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
import { buildRegionBoundaryLines, buildRegionLabels } from "./region-boundaries";

// Geometric centre of the ED region boundary system (not exactly Sag A*)
const SAG_X = 436.14;
const SAG_Z = 26826.65;

function toPolar(edX: number, edZ: number) {
	const dx = edX - SAG_X, dz = edZ - SAG_Z;
	return { r: Math.sqrt(dx * dx + dz * dz), theta: Math.atan2(dx, dz) };
}

export interface MapDebugInfo {
	zoom: number;
	starSize: number;
	starBrightness: number;
	pivotWorld: { x: number; y: number; z: number };
	pivotED:    { x: number; y: number; z: number };
	pivotPolar: { r: number; theta: number };
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
	private highlightedMarkerIndex: number = -1;
	private hasLoaded = false;
	private backgroundGroup = new THREE.Group();
	private starField: THREE.Points | null = null;
	private starFieldVisible = true;
	private regionObjects: THREE.Object3D[] = [];
	private regionLabelMeshes: THREE.Mesh[] = [];
	private regionVisible = true;
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
		private readonly labelContainer: HTMLDivElement,
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
			points.visible = this.starFieldVisible;
			this.starField = points;
			this.backgroundGroup.add(points);
		});
		const boundaryLines = buildRegionBoundaryLines();
		this.regionObjects.push(boundaryLines);
		this.backgroundGroup.add(boundaryLines);

		for (const mesh of buildRegionLabels()) {
			this.regionObjects.push(mesh);
			this.regionLabelMeshes.push(mesh);
			this.backgroundGroup.add(mesh);
		}

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
			obj.position.y += 0.25;
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
			obj.position.y += 0.25;
			this.customMarkerGroup.add(obj);
		}
	}

	private fitCamera(routes: EditViewRoute[]) {
		const box = computeBoundingBox(routes);
		const center = new THREE.Vector3();
		const size = new THREE.Vector3();
		box.getCenter(center);
		box.getSize(size);
		center.y = 0;

		const maxDim = Math.max(size.x, size.y, size.z);
		const fovRad = (this.camera.fov * Math.PI) / 180;
		const distance = (maxDim / 2 / Math.tan(fovRad / 2)) * 1.5;

		this.camera.position.set(
			center.x,
			center.y + distance * 0.6,
			center.z - distance * 0.8,
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
				// Size: bell curve in log-distance space, peaks at PEAK_DIST
				const STAR_MIN_PX = 0.8, STAR_MAX_PX = 1.3;
				const STAR_PEAK_DIST = 25, STAR_SIGMA = 0.9;
				const bell = Math.exp(-0.5 * ((Math.log(dist / STAR_PEAK_DIST)) / STAR_SIGMA) ** 2);
				mat.size = STAR_MIN_PX + (STAR_MAX_PX - STAR_MIN_PX) * bell;

				// Colour brightness: dims the star field as you zoom out to prevent additive saturation
				const brightness = Math.max(0.15, Math.min(0.85, (8 / dist) * 0.85));
				mat.color.setScalar(brightness);
			}
			this.crosshair.scale.setScalar(dist * 0.02);

			// Region name labels: fade opacity dist 20→9.5, shrink scale dist 25→15
			const regionLabelOpacity = Math.max(0.1, Math.min(1, 0.1 + (dist - 9.5) / (20 - 9.5) * 0.9));
			// Scale from 1.0 at dist=25 down to ~0.71 at dist=15
			// (equivalent to LABEL_PX_TO_WORLD going from 0.014 → 0.01)
			const regionLabelScale = Math.max(0.71, Math.min(1, 0.71 + (dist - 15) / (25 - 15) * 0.29));
			for (const mesh of this.regionLabelMeshes) {
				(mesh.material as THREE.MeshBasicMaterial).opacity = regionLabelOpacity;
				// Preserve the -1 X scale that corrects the mirror orientation
				mesh.scale.set(-regionLabelScale, regionLabelScale, regionLabelScale);
			}

			const labelOpacity = 1 - Math.max(0, Math.min(1, (dist - 4.5) / 3.5));
			this.labelContainer.querySelectorAll<HTMLElement>('.map-label--custom').forEach(el => {
				el.style.opacity = String(labelOpacity);
			});
			if (this.onFrame) {
				const t = this.controls.target;
				const sfMat = this.starField?.material as THREE.PointsMaterial | undefined;
				const edX = -t.x * 1000, edZ = t.z * 1000;
				this.onFrame({
					zoom: dist,
					starSize: sfMat?.size ?? 0,
					starBrightness: sfMat?.color.r ?? 0,
					pivotWorld: { x: t.x, y: t.y, z: t.z },
					pivotED: { x: edX, y: t.y * 1000, z: edZ },
					pivotPolar: toPolar(edX, edZ),
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
		const tag = (document.activeElement as HTMLElement)?.tagName;
		if (tag === 'INPUT' || tag === 'TEXTAREA') return;

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

	setHighlightedMarker(name: string | null) {
		if (!this.customMarkerPoints) return;
		const colors = this.customMarkerPoints.geometry.attributes.color as THREE.BufferAttribute;
		const base = new THREE.Color('#A0C4FF');
		const highlight = new THREE.Color('#FF7800');

		// Restore previous highlight
		if (this.highlightedMarkerIndex >= 0) {
			colors.setXYZ(this.highlightedMarkerIndex, base.r, base.g, base.b);
		}

		const idx = name ? this.loadedCustomMarkers.findIndex((m) => m.name === name) : -1;
		this.highlightedMarkerIndex = idx;

		if (idx >= 0) {
			colors.setXYZ(idx, highlight.r, highlight.g, highlight.b);
		}
		colors.needsUpdate = true;
	}

	getCameraState() {
		return {
			targetX: this.controls.target.x,
			targetY: this.controls.target.y,
			targetZ: this.controls.target.z,
			posX: this.camera.position.x,
			posY: this.camera.position.y,
			posZ: this.camera.position.z,
		};
	}

	setCameraState(cam: { targetX: number; targetY: number; targetZ: number; posX: number; posY: number; posZ: number }) {
		this.controls.target.set(cam.targetX, cam.targetY, cam.targetZ);
		this.camera.position.set(cam.posX, cam.posY, cam.posZ);
		this.controls.update();
	}

	setBoundaryMarks(marks: { id: number; x: number; z: number }[]) {
		// Remove previous mark objects
		this.scene.children
			.filter((c) => c.userData['boundaryMark'])
			.forEach((c) => this.scene.remove(c));

		const color = new THREE.Color('#FFFF00');
		for (const m of marks) {
			// Dot at the mark position
			const geo = new THREE.BufferGeometry();
			const wx = -m.x / 1000, wz = m.z / 1000;
			geo.setAttribute('position', new THREE.BufferAttribute(new Float32Array([wx, 0, wz]), 3));
			const mat = new THREE.PointsMaterial({ color, size: 8, sizeAttenuation: false, depthTest: false });
			const pt = new THREE.Points(geo, mat);
			pt.userData['boundaryMark'] = true;
			pt.renderOrder = 10;
			this.scene.add(pt);

			// Numbered CSS2D label
			const el = document.createElement('div');
			el.className = 'map-label map-label--boundary-mark';
			el.textContent = String(m.id);
			const obj = new CSS2DObject(el);
			obj.position.set(wx, 0.1, wz);
			obj.userData['boundaryMark'] = true;
			obj.renderOrder = 10;
			this.scene.add(obj);
		}
	}

	setStarFieldVisible(visible: boolean) {
		this.starFieldVisible = visible;
		if (this.starField) this.starField.visible = visible;
	}

	setRegionsVisible(visible: boolean) {
		this.regionVisible = visible;
		for (const obj of this.regionObjects) obj.visible = visible;
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
