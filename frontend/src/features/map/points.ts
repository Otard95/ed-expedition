import * as THREE from 'three';
import type { EditViewRoute } from '../../lib/routes/edit';
import { COORD_SCALE, REFERENCE_POINTS, REF_COLOR, ROUTE_COLORS } from './constants';

function makeCircleTexture(size = 64): THREE.Texture {
  const canvas = document.createElement('canvas');
  canvas.width = size;
  canvas.height = size;
  const ctx = canvas.getContext('2d')!;
  const c = size / 2;
  const r = size / 2 - 1;
  ctx.fillStyle = 'rgba(255,255,255,1.0)';
  ctx.beginPath();
  ctx.arc(c, c, r, 0, Math.PI * 2);
  ctx.fill();
  return new THREE.CanvasTexture(canvas);
}

const circleTexture = makeCircleTexture();

function clamp(v: number, min: number, max: number) {
  return Math.max(min, Math.min(max, v));
}

function worldSizeForRoute(route: EditViewRoute): number {
  let totalDist = 0;
  let count = 0;
  let prev: [number, number, number] | null = null;

  for (const jump of route.jumps) {
    if (!jump.position) continue;
    const cur = toWorldCoords(jump.position.x, jump.position.y, jump.position.z);
    if (prev) {
      const dx = cur[0] - prev[0];
      const dy = cur[1] - prev[1];
      const dz = cur[2] - prev[2];
      totalDist += Math.sqrt(dx * dx + dy * dy + dz * dz);
      count++;
    }
    prev = cur;
  }

  if (count === 0) return 0.05;

  return clamp((totalDist / count) / 3, 0.03, 0.5);
}

function toWorldCoords(x: number, y: number, z: number): [number, number, number] {
  // Negate X to match the in-game galaxy map orientation (negative X = West = left on screen)
  return [-x * COORD_SCALE, y * COORD_SCALE, z * COORD_SCALE];
}

export function buildRoutePointObjects(routes: EditViewRoute[]): THREE.Points[] {
  return routes.map((route, idx) => {
    const positions: number[] = [];

    for (const jump of route.jumps) {
      if (!jump.position) continue;
      const [wx, wy, wz] = toWorldCoords(jump.position.x, jump.position.y, jump.position.z);
      positions.push(wx, wy, wz);
    }

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));

    const material = new THREE.PointsMaterial({
      color: ROUTE_COLORS[idx % ROUTE_COLORS.length],
      size: worldSizeForRoute(route),
      sizeAttenuation: true,
      map: circleTexture,
      alphaTest: 0.5,
    });

    return new THREE.Points(geometry, material);
  });
}

export function buildReferencePointObjects(): { points: THREE.Points; labels: { position: THREE.Vector3; name: string }[] } {
  const positions: number[] = [];
  const labels: { position: THREE.Vector3; name: string }[] = [];

  for (const ref of REFERENCE_POINTS) {
    const [wx, wy, wz] = toWorldCoords(ref.x, ref.y, ref.z);
    positions.push(wx, wy, wz);
    labels.push({ position: new THREE.Vector3(wx, wy, wz), name: ref.name });
  }

  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));

  const material = new THREE.PointsMaterial({
    color: REF_COLOR,
    size: 0.2,
    sizeAttenuation: true,
    map: circleTexture,
    alphaTest: 0.5,
  });

  return { points: new THREE.Points(geometry, material), labels };
}

export function buildCustomMarkerObjects(markers: { name: string; x: number; y: number; z: number }[]): { points: THREE.Points; labels: { position: THREE.Vector3; name: string }[] } {
  const positions: number[] = [];
  const labels: { position: THREE.Vector3; name: string }[] = [];

  for (const m of markers) {
    const [wx, wy, wz] = toWorldCoords(m.x, m.y, m.z);
    positions.push(wx, wy, wz);
    labels.push({ position: new THREE.Vector3(wx, wy, wz), name: m.name });
  }

  const geometry = new THREE.BufferGeometry();
  geometry.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));

  const material = new THREE.PointsMaterial({
    color: new THREE.Color('#A0C4FF'),
    size: 0.2,
    sizeAttenuation: true,
    map: circleTexture,
    alphaTest: 0.5,
  });

  return { points: new THREE.Points(geometry, material), labels };
}

export function computeBoundingBox(routes: EditViewRoute[]): THREE.Box3 {
  const box = new THREE.Box3();

  for (const route of routes) {
    for (const jump of route.jumps) {
      if (!jump.position) continue;
      const [wx, wy, wz] = toWorldCoords(jump.position.x, jump.position.y, jump.position.z);
      box.expandByPoint(new THREE.Vector3(wx, wy, wz));
    }
  }

  // Always include Sol so the box is never empty
  box.expandByPoint(new THREE.Vector3(0, 0, 0));

  return box;
}
