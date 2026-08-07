import * as THREE from 'three';

export function buildCrosshair(): THREE.LineSegments {
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
