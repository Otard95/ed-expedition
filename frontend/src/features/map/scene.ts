import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import {
	CSS2DObject,
	CSS2DRenderer,
} from "three/examples/jsm/renderers/CSS2DRenderer.js";
import type { EditViewRoute, EditViewRouteJump } from "../../lib/routes/edit";
import { buildCrosshair } from "./crosshair";
import { buildGalaxyStarField } from "./galaxy-starfield";
import {
	buildCustomMarkerObjects,
	buildReferencePointObjects,
	buildRoutePointObjects,
	computeBoundingBox,
} from "./points";
import { buildRegionLinesPlane } from "./region-lines";

export interface MapDebugInfo {
	zoom: number;
	starSize: number;
	pivotWorld: { x: number; y: number; z: number };
	pivotED: { x: number; y: number; z: number };
}

export class MapScene {
	private scene: THREE.Scene;
	private camera: THREE.PerspectiveCamera;
	private renderer: THREE.WebGLRenderer;
	private labelRenderer: CSS2DRenderer;
	private controls: OrbitControls;
	private animationId: number | null = null;
	private raycaster = new THREE.Raycaster();
	// Parallel arrays: each entry in routePoints[i] corresponds to routes[i]
	private routePoints: THREE.Points[] = [];
	private loadedRoutes: EditViewRoute[] = [];
	private customMarkerGroup = new THREE.Group();
	private customMarkerPoints: THREE.Points | null = null;
	private loadedCustomMarkers: {
		name: string;
		x: number;
		y: number;
		z: number;
	}[] = [];
	private hasLoaded = false;
	private backgroundGroup = new THREE.Group();
	private starField: THREE.Points | null = null;
	private crosshair: THREE.LineSegments;
	private keys = new Set<string>();
	private boundKeyDown: (e: KeyboardEvent) => void;
	private boundKeyUp: (e: KeyboardEvent) => void;
	private boundMouseDown: (e: MouseEvent) => void;
	private boundMouseMove: (e: MouseEvent) => void;
	private boundMouseUp: (e: MouseEvent) => void;
	private rightDragLastY: number | null = null;
	onFrame: ((debug: MapDebugInfo) => void) | null = null;

	constructor(
		private readonly container: HTMLDivElement,
		labelContainer: HTMLDivElement,
	) {
		this.scene = new THREE.Scene();
		this.scene.background = new THREE.Color("#000000");

		const { clientWidth: w, clientHeight: h } = container;

		this.camera = new THREE.PerspectiveCamera(60, w / h, 0.001, 10000);

		this.renderer = new THREE.WebGLRenderer({ antialias: true });
		this.renderer.setPixelRatio(window.devicePixelRatio);
		this.renderer.setSize(w, h);
		container.appendChild(this.renderer.domElement);

		this.labelRenderer = new CSS2DRenderer();
		this.labelRenderer.setSize(w, h);
		this.labelRenderer.domElement.style.position = "absolute";
		this.labelRenderer.domElement.style.top = "0";
		this.labelRenderer.domElement.style.left = "0";
		this.labelRenderer.domElement.style.pointerEvents = "none";
		labelContainer.appendChild(this.labelRenderer.domElement);

		this.controls = new OrbitControls(this.camera, this.renderer.domElement);
		this.controls.enableDamping = true;
		this.controls.dampingFactor = 0.05;
		this.controls.screenSpacePanning = true;
		this.controls.enablePan = false;
		this.controls.maxDistance = 90;

		// Star field and region lines live in a persistent group that survives scene.clear()
		buildGalaxyStarField().then((points) => {
			this.starField = points;
			this.backgroundGroup.add(points);
		});
		buildRegionLinesPlane().then((mesh) => this.backgroundGroup.add(mesh));

		this.crosshair = buildCrosshair();

		this.boundKeyDown = (e) => this.keys.add(e.key.toLowerCase());
		this.boundKeyUp = (e) => this.keys.delete(e.key.toLowerCase());
		document.addEventListener("keydown", this.boundKeyDown);
		document.addEventListener("keyup", this.boundKeyUp);

		this.boundMouseDown = (e) => {
			if (e.button === 2) this.rightDragLastY = e.clientY;
		};
		this.boundMouseMove = (e) => {
			if (this.rightDragLastY === null) return;
			const dy = e.clientY - this.rightDragLastY;
			this.rightDragLastY = e.clientY;
			const dist = this.camera.position.distanceTo(this.controls.target);
			const delta = (dy / this.container.clientHeight) * dist;
			this.camera.position.y += delta;
			this.controls.target.y += delta;
		};
		this.boundMouseUp = (e) => {
			if (e.button === 2) this.rightDragLastY = null;
		};
		this.renderer.domElement.addEventListener("mousedown", this.boundMouseDown);
		this.renderer.domElement.addEventListener("mousemove", this.boundMouseMove);
		this.renderer.domElement.addEventListener("mouseup", this.boundMouseUp);
	}

	load(routes: EditViewRoute[]) {
		const savedCameraPos = this.hasLoaded ? this.camera.position.clone() : null;
		const savedTarget = this.hasLoaded ? this.controls.target.clone() : null;

		this.scene.clear();

		this.loadedRoutes = routes;
		this.routePoints = buildRoutePointObjects(routes);
		this.routePoints.forEach((p) => this.scene.add(p));

		const { points: refPoints, labels } = buildReferencePointObjects();
		this.scene.add(refPoints);

		for (const { position, name } of labels) {
			const el = document.createElement("div");
			el.className = "map-label";
			el.textContent = name;
			const obj = new CSS2DObject(el);
			obj.position.copy(position);
			obj.position.y += 0.5;
			this.scene.add(obj);
		}

		this.scene.add(this.customMarkerGroup);
		this.scene.add(this.backgroundGroup);
		this.scene.add(this.crosshair);

		if (savedCameraPos && savedTarget) {
			this.camera.position.copy(savedCameraPos);
			this.controls.target.copy(savedTarget);
			this.controls.update();
		} else {
			this.fitCamera(routes);
		}

		this.hasLoaded = true;
	}

	setCustomMarkers(
		markers: { name: string; x: number; y: number; z: number }[],
	) {
		this.customMarkerGroup.clear();
		this.customMarkerPoints = null;
		this.loadedCustomMarkers = markers;
		if (markers.length === 0) return;

		const { points, labels } = buildCustomMarkerObjects(markers);
		this.customMarkerPoints = points;
		this.customMarkerGroup.add(points);

		for (const { position, name } of labels) {
			const el = document.createElement("div");
			el.className = "map-label map-label--custom";
			el.textContent = name;
			const obj = new CSS2DObject(el);
			obj.position.copy(position);
			obj.position.y += 0.5;
			this.customMarkerGroup.add(obj);
		}
	}

	private fitCamera(routes: EditViewRoute[]) {
		const box = computeBoundingBox(routes);
		const center = new THREE.Vector3();
		box.getCenter(center);
		center.y = 0;

		const distance = 200;
		this.camera.position.set(
			center.x,
			center.y + distance * 0.95,
			center.z - distance * 0.2,
		);
		this.camera.lookAt(center);
		this.controls.target.copy(center);
		this.controls.update();
	}

	start() {
		const animate = () => {
			this.animationId = requestAnimationFrame(animate);
			this.applyWASD();
			this.controls.update();
			this.crosshair.position.copy(this.controls.target);
			const dist = this.camera.position.distanceTo(this.controls.target);
			if (this.starField) {
				const mat = this.starField.material as THREE.PointsMaterial;
				// Gaussian bell in log-distance space: peaks at PEAK_DIST, falls off both sides.
				// If stars visibly shrink past zoom=50 this confirms sizing works; if not, it's density illusion.
				const STAR_MIN_PX = 0.8,
					STAR_MAX_PX = 1.3;
				const STAR_PEAK_DIST = 25,
					STAR_SIGMA = 0.001;
				const bell = Math.exp(
					-0.5 * (Math.log(dist / STAR_PEAK_DIST) / STAR_SIGMA) ** 2,
				);
				mat.size = STAR_MIN_PX + (STAR_MAX_PX - STAR_MIN_PX) * bell;
			}
			this.crosshair.scale.setScalar(dist * 0.02);
			if (this.onFrame) {
				const t = this.controls.target;
				this.onFrame({
					zoom: dist,
					starSize: this.starField
						? (this.starField.material as THREE.PointsMaterial).size
						: 0,
					pivotWorld: { x: t.x, y: t.y, z: t.z },
					pivotED: { x: -t.x * 1000, y: t.y * 1000, z: t.z * 1000 },
				});
			}
			this.renderer.render(this.scene, this.camera);
			this.labelRenderer.render(this.scene, this.camera);
		};
		animate();
	}

	resize() {
		const { clientWidth: w, clientHeight: h } = this.container;
		this.camera.aspect = w / h;
		this.camera.updateProjectionMatrix();
		this.renderer.setSize(w, h);
		this.labelRenderer.setSize(w, h);
	}

	private applyWASD() {
		if (this.keys.size === 0) return;

		const dist = this.camera.position.distanceTo(this.controls.target);
		const speed = dist * 0.015;

		const forward = new THREE.Vector3();
		this.camera.getWorldDirection(forward);
		forward.y = 0;
		if (forward.lengthSq() < 0.0001) return;
		forward.normalize();

		const right = new THREE.Vector3()
			.crossVectors(forward, new THREE.Vector3(0, 1, 0))
			.normalize();

		const delta = new THREE.Vector3();
		if (this.keys.has("w")) delta.addScaledVector(forward, speed);
		if (this.keys.has("s")) delta.addScaledVector(forward, -speed);
		if (this.keys.has("d")) delta.addScaledVector(right, speed);
		if (this.keys.has("a")) delta.addScaledVector(right, -speed);

		this.camera.position.add(delta);
		this.controls.target.add(delta);
	}

	pick(
		mouseX: number,
		mouseY: number,
		rect: DOMRect,
	):
		| {
				kind: "route";
				jump: EditViewRouteJump;
				route: EditViewRoute;
				jumpIndex: number;
		  }
		| { kind: "custom"; name: string; x: number; y: number; z: number }
		| null {
		const x = ((mouseX - rect.left) / rect.width) * 2 - 1;
		const y = -((mouseY - rect.top) / rect.height) * 2 + 1;

		this.raycaster.setFromCamera(new THREE.Vector2(x, y), this.camera);

		for (let i = 0; i < this.routePoints.length; i++) {
			const material = this.routePoints[i].material as THREE.PointsMaterial;
			this.raycaster.params.Points!.threshold = material.size / 2;
			const hits = this.raycaster.intersectObject(this.routePoints[i]);
			if (hits.length === 0) continue;

			const pointIndex = hits[0].index!;
			const route = this.loadedRoutes[i];

			let seen = 0;
			for (let j = 0; j < route.jumps.length; j++) {
				if (!route.jumps[j].position) continue;
				if (seen === pointIndex)
					return { kind: "route", jump: route.jumps[j], route, jumpIndex: j };
				seen++;
			}
		}

		if (this.customMarkerPoints) {
			const material = this.customMarkerPoints.material as THREE.PointsMaterial;
			this.raycaster.params.Points!.threshold = material.size / 2;
			const hits = this.raycaster.intersectObject(this.customMarkerPoints);
			if (hits.length > 0) {
				const marker = this.loadedCustomMarkers[hits[0].index!];
				if (marker) return { kind: "custom", ...marker };
			}
		}

		return null;
	}

	destroy() {
		if (this.animationId !== null) cancelAnimationFrame(this.animationId);
		document.removeEventListener("keydown", this.boundKeyDown);
		document.removeEventListener("keyup", this.boundKeyUp);
		this.renderer.domElement.removeEventListener(
			"mousedown",
			this.boundMouseDown,
		);
		this.renderer.domElement.removeEventListener(
			"mousemove",
			this.boundMouseMove,
		);
		this.renderer.domElement.removeEventListener("mouseup", this.boundMouseUp);
		this.controls.dispose();
		this.renderer.dispose();

		if (this.renderer.domElement.parentNode) {
			this.renderer.domElement.parentNode.removeChild(this.renderer.domElement);
		}
		if (this.labelRenderer.domElement.parentNode) {
			this.labelRenderer.domElement.parentNode.removeChild(
				this.labelRenderer.domElement,
			);
		}
	}
}
