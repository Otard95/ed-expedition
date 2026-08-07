import * as THREE from "three";

export const COORD_SCALE = 1 / 1000;

export const REFERENCE_POINTS = [
	{ name: "Sol", x: 0.0, y: 0.0, z: 0.0 },
	{ name: "Sagittarius A*", x: 25.2, y: -21.0, z: 25899.9 },
	// { name: 'Galactic Centre', x:   436.1, y:    0.0, z:  26826.7 },
	{ name: "Colonia", x: -9530.5, y: -910.3, z: 19808.1 },
	{ name: "Beagle Point", x: -1111.6, y: -134.3, z: 65269.7 },
] as const;

export const REF_COLOR = new THREE.Color("#FF7800");

export const ROUTE_COLORS = [
	new THREE.Color("#4FC3F7"), // sky blue
	new THREE.Color("#81C784"), // sage green
	new THREE.Color("#FFD54F"), // golden yellow
	new THREE.Color("#CE93D8"), // lavender
	new THREE.Color("#4DD0E1"), // cyan
	new THREE.Color("#FF8A65"), // coral
	new THREE.Color("#F06292"), // pink
	new THREE.Color("#AED581"), // lime
];
