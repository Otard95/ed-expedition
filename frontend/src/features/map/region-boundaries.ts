import * as THREE from "three";
import boundaryData from "../../assets/map/boundaries.json";

// Galactic centre in ED coordinates — matches the Go pipeline
const GC_X = 436.14;
const GC_Z = 26826.65;
const ARC_STEP_RAD = 0.008;

function fromPolar(r: number, theta: number): [number, number] {
	return [GC_X + r * Math.sin(theta), GC_Z + r * Math.cos(theta)];
}

// ED [x, z] → Three.js world coords (negate X, scale by 1/1000)
function toWorld(edX: number, edZ: number): [number, number, number] {
	return [-edX / 1000, 0, edZ / 1000];
}

function interpolateArc(
	r: number,
	theta0: number,
	theta1: number,
	x0: number,
	z0: number, // exact start position (forced)
	x1: number,
	z1: number, // exact end position (forced)
	positions: number[],
) {
	let dt = theta1 - theta0;
	while (dt > Math.PI) dt -= 2 * Math.PI;
	while (dt < -Math.PI) dt += 2 * Math.PI;

	const steps = Math.max(2, Math.ceil(Math.abs(dt) / ARC_STEP_RAD));
	let [px, pz] = [x0, z0];
	for (let i = 1; i <= steps; i++) {
		let cx: number, cz: number;
		if (i === steps) {
			[cx, cz] = [x1, z1]; // force exact endpoint — no float gap
		} else {
			const t = theta0 + (dt * i) / steps;
			[cx, cz] = fromPolar(r, t);
		}
		const [wx1, wy1, wz1] = toWorld(px, pz);
		const [wx2, wy2, wz2] = toWorld(cx, cz);
		positions.push(wx1, wy1, wz1, wx2, wy2, wz2);
		px = cx;
		pz = cz;
	}
}

const LABEL_FONT_PX = 48;
// Maximum canvas width before text wraps, in pixels at canvas resolution.
const LABEL_MAX_PX = 320;
// World units per canvas pixel — controls physical size on the galaxy plane.
const LABEL_PX_TO_WORLD = 0.014;

function makeLabelMesh(name: string): THREE.Mesh {
	const canvas = document.createElement("canvas");
	const ctx = canvas.getContext("2d")!;
	const font = `700 ${LABEL_FONT_PX}px 'Segoe UI', system-ui, sans-serif`;

	ctx.font = font;

	// Word-wrap: break on spaces to stay within LABEL_MAX_PX
	const words = name.split(" ");
	const lines: string[] = [];
	let line = "";
	for (const word of words) {
		const test = line ? `${line} ${word}` : word;
		if (ctx.measureText(test).width > LABEL_MAX_PX && line) {
			lines.push(line);
			line = word;
		} else {
			line = test;
		}
	}
	lines.push(line);

	const lineH = Math.round(LABEL_FONT_PX * 1.25);
	const pad = 8;
	const textW = Math.max(...lines.map((l) => ctx.measureText(l).width));
	canvas.width = Math.ceil(textW) + pad * 2;
	canvas.height = lines.length * lineH + pad * 2;

	ctx.font = font;
	ctx.fillStyle = "#CC6600";
	ctx.textAlign = "center";
	ctx.textBaseline = "top";
	lines.forEach((l, i) => ctx.fillText(l, canvas.width / 2, pad + i * lineH));

	const texture = new THREE.CanvasTexture(canvas);
	texture.minFilter = THREE.LinearFilter;
	texture.flipY = false;
	const material = new THREE.MeshBasicMaterial({
		map: texture,
		transparent: true,
		depthWrite: false,
		depthTest: false,
		side: THREE.DoubleSide,
	});
	const w = canvas.width * LABEL_PX_TO_WORLD;
	const h = canvas.height * LABEL_PX_TO_WORLD;
	const geo = new THREE.PlaneGeometry(w, h);
	const mesh = new THREE.Mesh(geo, material);
	// Rotate to lie flat on the XZ (galactic) plane
	mesh.rotation.x = -Math.PI / 2;
	mesh.scale.x = -1;
	mesh.renderOrder = 2;
	return mesh;
}

export function buildRegionLabels(): THREE.Mesh[] {
	const { vertices, names, edges, name_positions } =
		boundaryData as unknown as {
			vertices: [number, number][];
			names: string[];
			edges: number[][];
			edge_type: string[][];
			name_positions: ([number, number] | null)[];
		};

	return names.map((name, si) => {
		let wx: number, wz: number;
		const override = name_positions?.[si];
		if (override) {
			[wx, , wz] = toWorld(override[0], override[1]);
		} else {
			const verts = edges[si];
			let sumX = 0,
				sumZ = 0;
			for (const vi of verts) {
				const [r, t] = vertices[vi];
				const [edX, edZ] = fromPolar(r, t);
				const [wx2, , wz2] = toWorld(edX, edZ);
				sumX += wx2;
				sumZ += wz2;
			}
			wx = sumX / verts.length;
			wz = sumZ / verts.length;
		}
		const mesh = makeLabelMesh(name);
		mesh.position.set(wx, 0, wz);
		return mesh;
	});
}

export function buildRegionBoundaryLines(): THREE.LineSegments {
	const { vertices, edges, edge_type } = boundaryData as unknown as {
		vertices: [number, number][];
		names: string[];
		edges: number[][];
		edge_type: string[][];
		name_positions: ([number, number] | null)[];
	};

	const positions: number[] = [];
	const seen = new Set<string>();

	for (let si = 0; si < edges.length; si++) {
		const verts = edges[si];
		const types = edge_type[si];
		const n = verts.length;

		for (let i = 0; i < n; i++) {
			const ai = verts[i];
			const bi = verts[(i + 1) % n];

			// Each shared edge only rendered once
			const key = `${Math.min(ai, bi)}:${Math.max(ai, bi)}`;
			if (seen.has(key)) continue;
			seen.add(key);

			const [rA, tA] = vertices[ai];
			const [rB, tB] = vertices[bi];

			if (types[i] === "radial") {
				const [x1, z1] = fromPolar(rA, tA);
				const [x2, z2] = fromPolar(rB, tB);
				const [wx1, wy1, wz1] = toWorld(x1, z1);
				const [wx2, wy2, wz2] = toWorld(x2, z2);
				positions.push(wx1, wy1, wz1, wx2, wy2, wz2);
			} else {
				const r = (rA + rB) / 2;
				const [x1, z1] = fromPolar(rA, tA);
				const [x2, z2] = fromPolar(rB, tB);
				interpolateArc(r, tA, tB, x1, z1, x2, z2, positions);
			}
		}
	}

	const geo = new THREE.BufferGeometry();
	geo.setAttribute("position", new THREE.Float32BufferAttribute(positions, 3));

	const mat = new THREE.LineBasicMaterial({
		color: new THREE.Color("#CC6600"),
		opacity: 0.6,
		transparent: true,
		depthTest: false,
	});

	const lines = new THREE.LineSegments(geo, mat);
	lines.renderOrder = 1;
	return lines;
}
