import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { CSS2DRenderer, CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
import type { EditViewRoute, EditViewRouteJump } from '../../lib/routes/edit';
import { buildRoutePointObjects, buildReferencePointObjects, buildCustomMarkerObjects, computeBoundingBox } from './points';
import galaxyUrl from '../../assets/map/galaxy_z3.png';
import regionLinesUrl from '../../assets/map/regionlines_z3.png';

// From galmap.js galmapCoords(): ED_X = (lng-128)*640, ED_Z = (128+lat)*640+25000
// At zoom 3 (scale=8): pixel = latLng * 8
// Full image spans ED X ±81920 ly, ED Z from -56920 to +106920 ly (total 163840 ly square)
// Plane center in ED: (0, 25000) → Three.js (0, 0, 25.0)
const GALAXY_PLANE_SIZE = (81920 * 2) / 1000; // 163.84 world units
const GALAXY_PLANE_CENTER_Z = 25000 / 1000;   // 25.0 world units

// The image X axis (left→right) maps to increasing ED X, but our Three.js X
// is negated (-ED_X), so the texture is horizontally mirrored relative to the
// plane UVs — we fix this with texture.repeat.x = -1, offset.x = 1.
function buildGalaxyPlane(): THREE.Mesh {
  const loader = new THREE.TextureLoader();
  const tex = loader.load(galaxyUrl);
  tex.colorSpace = THREE.SRGBColorSpace;
  // Three.js loads PNGs with Y flipped; the tile image already has y=0 at top
  // (high Z = galactic north), which matches our coordinate system, so no flipY needed.
  tex.flipY = false;
  // Mirror horizontally to compensate for our negated X axis.
  tex.repeat.set(-1, 1);
  tex.offset.set(1, 0);

  const geo = new THREE.PlaneGeometry(GALAXY_PLANE_SIZE, GALAXY_PLANE_SIZE);
  geo.rotateX(-Math.PI / 2);
  const mat = new THREE.MeshBasicMaterial({
    map: tex,
    blending: THREE.AdditiveBlending,
    depthWrite: false,
    transparent: true,
    opacity: 0.6,
    side: THREE.DoubleSide,
  });
  const mesh = new THREE.Mesh(geo, mat);
  mesh.position.set(0, -0.01, GALAXY_PLANE_CENTER_Z);
  return mesh;
}

async function buildRegionLinesPlane(): Promise<THREE.Mesh> {
  const img = await new Promise<HTMLImageElement>((resolve, reject) => {
    const i = new Image();
    i.onload = () => resolve(i);
    i.onerror = reject;
    i.src = regionLinesUrl;
  });

  // Paint onto a canvas and replace near-white pixels with transparent.
  const canvas = document.createElement('canvas');
  canvas.width = img.width;
  canvas.height = img.height;
  const ctx = canvas.getContext('2d')!;
  ctx.drawImage(img, 0, 0);
  const data = ctx.getImageData(0, 0, canvas.width, canvas.height);
  for (let i = 0; i < data.data.length; i += 4) {
    const r = data.data[i], g = data.data[i + 1], b = data.data[i + 2];
    if (r > 230 && g > 230 && b > 230) data.data[i + 3] = 0;
  }
  ctx.putImageData(data, 0, 0);

  const tex = new THREE.CanvasTexture(canvas);
  tex.colorSpace = THREE.SRGBColorSpace;
  tex.flipY = false;
  tex.repeat.set(-1, 1);
  tex.offset.set(1, 0);

  const geo = new THREE.PlaneGeometry(GALAXY_PLANE_SIZE, GALAXY_PLANE_SIZE);
  geo.rotateX(-Math.PI / 2);
  const mat = new THREE.MeshBasicMaterial({
    map: tex,
    transparent: true,
    depthWrite: false,
    opacity: 0.7,
    side: THREE.DoubleSide,
  });
  const mesh = new THREE.Mesh(geo, mat);
  mesh.position.set(0, -0.005, GALAXY_PLANE_CENTER_Z);
  return mesh;
}

function buildCrosshair(): THREE.LineSegments {
  const geo = new THREE.BufferGeometry().setFromPoints([
    new THREE.Vector3(-1, 0,  0), new THREE.Vector3(1, 0,  0),
    new THREE.Vector3( 0, 0, -1), new THREE.Vector3(0, 0,  1),
    new THREE.Vector3( 0, -1, 0), new THREE.Vector3(0, 1,  0),
  ]);
  const mat = new THREE.LineBasicMaterial({
    color: 0xffffff,
    opacity: 0.5,
    transparent: true,
    depthTest: false,
  });
  const lines = new THREE.LineSegments(geo, mat);
  lines.renderOrder = 999;
  return lines;
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
  private loadedCustomMarkers: { name: string; x: number; y: number; z: number }[] = [];
  private hasLoaded = false;
  private backgroundGroup = new THREE.Group();
  private crosshair: THREE.LineSegments;
  private keys = new Set<string>();
  private boundKeyDown: (e: KeyboardEvent) => void;
  private boundKeyUp: (e: KeyboardEvent) => void;
  private boundMouseDown: (e: MouseEvent) => void;
  private boundMouseMove: (e: MouseEvent) => void;
  private boundMouseUp: (e: MouseEvent) => void;
  private rightDragLastY: number | null = null;

  constructor(
    private readonly container: HTMLDivElement,
    labelContainer: HTMLDivElement,
  ) {
    this.scene = new THREE.Scene();
    this.scene.background = new THREE.Color('#000000');

    const { clientWidth: w, clientHeight: h } = container;

    this.camera = new THREE.PerspectiveCamera(60, w / h, 0.001, 10000);

    this.renderer = new THREE.WebGLRenderer({ antialias: true });
    this.renderer.setPixelRatio(window.devicePixelRatio);
    this.renderer.setSize(w, h);
    container.appendChild(this.renderer.domElement);

    this.labelRenderer = new CSS2DRenderer();
    this.labelRenderer.setSize(w, h);
    this.labelRenderer.domElement.style.position = 'absolute';
    this.labelRenderer.domElement.style.top = '0';
    this.labelRenderer.domElement.style.left = '0';
    this.labelRenderer.domElement.style.pointerEvents = 'none';
    labelContainer.appendChild(this.labelRenderer.domElement);

    this.controls = new OrbitControls(this.camera, this.renderer.domElement);
    this.controls.enableDamping = true;
    this.controls.dampingFactor = 0.05;
    this.controls.screenSpacePanning = true;

    // Galaxy and region planes live in a persistent group that survives scene.clear().
    this.backgroundGroup.add(buildGalaxyPlane());
    buildRegionLinesPlane().then(mesh => this.backgroundGroup.add(mesh));

    this.crosshair = buildCrosshair();
    this.scene.add(this.crosshair);

    this.boundKeyDown = (e) => this.keys.add(e.key.toLowerCase());
    this.boundKeyUp   = (e) => this.keys.delete(e.key.toLowerCase());
    document.addEventListener('keydown', this.boundKeyDown);
    document.addEventListener('keyup',   this.boundKeyUp);

    this.controls.enablePan = false;

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
    this.renderer.domElement.addEventListener('mousedown', this.boundMouseDown);
    this.renderer.domElement.addEventListener('mousemove', this.boundMouseMove);
    this.renderer.domElement.addEventListener('mouseup',   this.boundMouseUp);
  }

  load(routes: EditViewRoute[]) {
    const savedCameraPos = this.hasLoaded ? this.camera.position.clone() : null;
    const savedTarget = this.hasLoaded ? this.controls.target.clone() : null;

    this.scene.clear();

    this.loadedRoutes = routes;
    this.routePoints = buildRoutePointObjects(routes);
    this.routePoints.forEach(p => this.scene.add(p));

    const { points: refPoints, labels } = buildReferencePointObjects();
    this.scene.add(refPoints);

    for (const { position, name } of labels) {
      const el = document.createElement('div');
      el.className = 'map-label';
      el.textContent = name;

      const obj = new CSS2DObject(el);
      obj.position.copy(position);
      // Small upward offset so the label sits above the point
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

  setCustomMarkers(markers: { name: string; x: number; y: number; z: number }[]) {
    this.customMarkerGroup.clear();
    this.customMarkerPoints = null;
    this.loadedCustomMarkers = markers;
    if (markers.length === 0) return;

    const { points, labels } = buildCustomMarkerObjects(markers);
    this.customMarkerPoints = points;
    this.customMarkerGroup.add(points);

    for (const { position, name } of labels) {
      const el = document.createElement('div');
      el.className = 'map-label map-label--custom';
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
    const size = new THREE.Vector3();
    box.getCenter(center);
    box.getSize(size);
    center.y = 0; // Always orbit around the galactic plane

    const maxDim = Math.max(size.x, size.y, size.z);
    const fovRad = (this.camera.fov * Math.PI) / 180;
    const distance = (maxDim / 2) / Math.tan(fovRad / 2) * 1.5;

    this.camera.position.set(center.x, center.y + distance * 0.6, center.z - distance * 0.8);
    this.camera.lookAt(center);
    this.controls.target.copy(center);
    this.controls.update();
  }

  start() {
    const animate = () => {
      this.animationId = requestAnimationFrame(animate);
      this.applyWASD();
      this.controls.update();
      this.backgroundGroup.position.y = this.controls.target.y;
      this.crosshair.position.copy(this.controls.target);
      // Scale crosshair so it stays a constant apparent size regardless of zoom
      const dist = this.camera.position.distanceTo(this.controls.target);
      this.crosshair.scale.setScalar(dist * 0.02);
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

    const right = new THREE.Vector3().crossVectors(forward, new THREE.Vector3(0, 1, 0)).normalize();

    const delta = new THREE.Vector3();
    if (this.keys.has('w')) delta.addScaledVector(forward,  speed);
    if (this.keys.has('s')) delta.addScaledVector(forward, -speed);
    if (this.keys.has('d')) delta.addScaledVector(right,    speed);
    if (this.keys.has('a')) delta.addScaledVector(right,   -speed);

    this.camera.position.add(delta);
    this.controls.target.add(delta);
  }

  pick(mouseX: number, mouseY: number, rect: DOMRect): 
    | { kind: 'route'; jump: EditViewRouteJump; route: EditViewRoute; jumpIndex: number }
    | { kind: 'custom'; name: string; x: number; y: number; z: number }
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

      // Map point index back to jump index, skipping jumps without positions
      let seen = 0;
      for (let j = 0; j < route.jumps.length; j++) {
        if (!route.jumps[j].position) continue;
        if (seen === pointIndex) return { kind: 'route', jump: route.jumps[j], route, jumpIndex: j };
        seen++;
      }
    }

    if (this.customMarkerPoints) {
      const material = this.customMarkerPoints.material as THREE.PointsMaterial;
      this.raycaster.params.Points!.threshold = material.size / 2;
      const hits = this.raycaster.intersectObject(this.customMarkerPoints);
      if (hits.length > 0) {
        const marker = this.loadedCustomMarkers[hits[0].index!];
        if (marker) return { kind: 'custom', ...marker };
      }
    }

    return null;
  }

  destroy() {
    if (this.animationId !== null) cancelAnimationFrame(this.animationId);
    document.removeEventListener('keydown', this.boundKeyDown);
    document.removeEventListener('keyup',   this.boundKeyUp);
    this.renderer.domElement.removeEventListener('mousedown', this.boundMouseDown);
    this.renderer.domElement.removeEventListener('mousemove', this.boundMouseMove);
    this.renderer.domElement.removeEventListener('mouseup',   this.boundMouseUp);
    this.controls.dispose();
    this.renderer.dispose();

    if (this.renderer.domElement.parentNode) {
      this.renderer.domElement.parentNode.removeChild(this.renderer.domElement);
    }
    if (this.labelRenderer.domElement.parentNode) {
      this.labelRenderer.domElement.parentNode.removeChild(this.labelRenderer.domElement);
    }
  }
}
