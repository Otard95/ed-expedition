import * as THREE from 'three';
import galaxyBlurredUrl from '../../assets/map/galaxy_z3_blurred.png';

// Coordinate mapping derived from EDastro galmap.js galmapCoords():
// ED_X = (lng-128)*640, ED_Z = (128+lat)*640+25000, at zoom 3 px = latLng*8
function pixelToWorld(px: number, py: number, imgWidth: number): [number, number] {
  const lng = px / (imgWidth / 256);
  const lat = -py / (imgWidth / 256);
  const edX = (lng - 128) * 640;
  const edZ = (128 + lat) * 640 + 25000;
  return [-edX / 1000, edZ / 1000];
}

function mulberry32(seed: number): () => number {
  return () => {
    seed |= 0; seed = seed + 0x6D2B79F5 | 0;
    let t = Math.imul(seed ^ seed >>> 15, 1 | seed);
    t = t + Math.imul(t ^ t >>> 7, 61 | t) ^ t;
    return ((t ^ t >>> 14) >>> 0) / 4294967296;
  };
}

function boxMullerGaussian(rand: () => number): number {
  const u1 = Math.max(rand(), 1e-10);
  const u2 = rand();
  return Math.sqrt(-2 * Math.log(u1)) * Math.cos(2 * Math.PI * u2);
}

const SPECTRAL_COLORS = [
  new THREE.Color('#9BB0FF'), // O - blue
  new THREE.Color('#CAD7FF'), // B - blue-white
  new THREE.Color('#F8F7FF'), // A - white
  new THREE.Color('#FFF4E8'), // F - yellow-white
  new THREE.Color('#FFE680'), // G - yellow
  new THREE.Color('#FFBD6F'), // K - orange
  new THREE.Color('#FF8C3B'), // M - red-orange
];

// Boosted from real stellar proportions so hot blue stars remain visible
const BASE_SPECTRAL_WEIGHTS = [0.005, 0.02, 0.06, 0.10, 0.20, 0.25, 0.365];

const STEP = 1;              // sample every pixel → full 2048×2048 grid
const DENSITY = 1.5;         // placement probability multiplier (scaled with STEP)
const GAMMA = 2.0;           // power curve exponent — >1 increases arm contrast
const MAX_HALF_THICKNESS = 3.0; // world units ≈ 3000 ly max half-thickness (bulge)
// World units spanned by one sample cell — used to jitter star positions
// off the pixel grid so they don't form a visible lattice
const CELL_SIZE = (163.84 / 2048) * STEP; // ≈ 0.08 world units at STEP=1

export async function buildGalaxyStarField(): Promise<THREE.Points> {
  const rand = mulberry32(0xDEADBEEF);

  const img = await new Promise<HTMLImageElement>((resolve, reject) => {
    const i = new Image();
    i.onload = () => resolve(i);
    i.onerror = reject;
    i.src = galaxyBlurredUrl;
  });

  const canvas = document.createElement('canvas');
  canvas.width = img.width;
  canvas.height = img.height;
  const ctx = canvas.getContext('2d')!;
  ctx.drawImage(img, 0, 0);
  const { data } = ctx.getImageData(0, 0, img.width, img.height);

  const positions: number[] = [];
  const colors: number[] = [];

  for (let py = 0; py < img.height; py += STEP) {
    for (let px = 0; px < img.width; px += STEP) {
      const i = (py * img.width + px) * 4;
      const r = data[i]     / 255;
      const g = data[i + 1] / 255;
      const b = data[i + 2] / 255;

      const lum = 0.299 * r + 0.587 * g + 0.114 * b;
      if (lum < 0.01) continue;
      if (rand() > Math.pow(lum, GAMMA) * DENSITY) continue;

      const jitterX = (rand() - 0.5) * CELL_SIZE;
      const jitterZ = (rand() - 0.5) * CELL_SIZE;
      const [worldX, worldZ] = pixelToWorld(px, py, img.width);
      const finalX = worldX + jitterX;
      const finalZ = worldZ + jitterZ;

      // Y scatter: brighter pixels = thicker disc = more vertical spread
      const sigma = lum * MAX_HALF_THICKNESS * 0.5;
      const worldY = Math.max(-MAX_HALF_THICKNESS,
        Math.min(MAX_HALF_THICKNESS, boxMullerGaussian(rand) * sigma));

      // Bias spectral type toward hot (blue pixel) or cool (warm pixel) stars
      const maxC = Math.max(r, g, b);
      const sat  = maxC < 0.001 ? 0 : (maxC - Math.min(r, g, b)) / maxC;
      const blueBias = sat * (b - r);

      const weights = BASE_SPECTRAL_WEIGHTS.map((w, j) =>
        w * Math.exp(blueBias * (3 - j) * 0.8)
      );
      const total = weights.reduce((a, c) => a + c, 0);
      let roll = rand() * total;
      let type = weights.length - 1;
      for (let j = 0; j < weights.length; j++) {
        roll -= weights[j];
        if (roll <= 0) { type = j; break; }
      }

      const brightness = 0.65 + rand() * 0.35;
      // Mix toward white to keep hue variation but avoid an overly yellow/orange cast
      const c = SPECTRAL_COLORS[type].clone().lerp(new THREE.Color(1, 1, 1), 0.25).multiplyScalar(brightness);

      positions.push(finalX, worldY, finalZ);
      colors.push(c.r, c.g, c.b);
    }
  }

  const geo = new THREE.BufferGeometry();
  geo.setAttribute('position', new THREE.Float32BufferAttribute(positions, 3));
  geo.setAttribute('color',    new THREE.Float32BufferAttribute(colors, 3));

  const mat = new THREE.PointsMaterial({
    size: 1.5,            // screen-space pixels, updated dynamically by MapScene
    sizeAttenuation: false,
    vertexColors: true,
    blending: THREE.AdditiveBlending, // stars add colour rather than occlude — can never block anything
    // opacity is left at default (1.0); brightness is controlled via mat.color in MapScene
    depthWrite: false,
    depthTest: false,
  });

  const points = new THREE.Points(geo, mat);
  points.renderOrder = -1;
  return points;
}
