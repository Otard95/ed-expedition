import * as THREE from 'three';
import regionLinesUrl from '../../assets/map/regionlines_z3.png';

// Coordinate mapping constants — same tile system as galaxy-starfield.ts
const PLANE_SIZE = (81920 * 2) / 1000;   // 163.84 world units
const PLANE_CENTER_Z = 25000 / 1000;     // 25.0 world units

export async function buildRegionLinesPlane(): Promise<THREE.Mesh> {
  const img = await new Promise<HTMLImageElement>((resolve, reject) => {
    const i = new Image();
    i.onload = () => resolve(i);
    i.onerror = reject;
    i.src = regionLinesUrl;
  });

  // Replace near-white background pixels with transparent
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
  // Mirror horizontally to compensate for our negated X axis
  tex.repeat.set(-1, 1);
  tex.offset.set(1, 0);

  const geo = new THREE.PlaneGeometry(PLANE_SIZE, PLANE_SIZE);
  geo.rotateX(-Math.PI / 2);
  const mat = new THREE.MeshBasicMaterial({
    map: tex,
    transparent: true,
    depthWrite: false,
    opacity: 0.7,
    side: THREE.DoubleSide,
  });
  const mesh = new THREE.Mesh(geo, mat);
  mesh.position.set(0, -0.005, PLANE_CENTER_Z);
  return mesh;
}
