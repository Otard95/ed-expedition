<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import * as THREE from "three";
  import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";

  type SystemPoint = {
    id: number;
    name: string;
    x: number;
    y: number;
    z: number;
    star_class: number;
  };

  type Vec3 = {
    x: number;
    y: number;
    z: number;
  };

  export let systems: SystemPoint[] = [];
  export let hilbertPoints: [number, number, number][] = [];
  export let title = "Systems cloud";
  export let center: Vec3 = { x: 0, y: 0, z: 0 };
  export let radius = 0;

  const backgroundColor = "#080a0e";
  const scoopableColor = "#ff9f43";
  const nonScoopableColor = "#6ee7f9";
  const hilbertPointColor = "#ef4444";
  const pointSize = 30;
  const hilbertPointSize = 42;
  const pointOpacity = 0.95;
  const pointAlphaTest = 0.08;
  const pointTextureSize = 64;
  const pointTextureInnerRadius = 6;
  const axisScaleFactor = 0.35;
  const cameraDistanceFactor = 1.65;
  const cameraHeightFactor = 0.65;
  const cameraMinNearFactor = 500;
  const cameraFarFactor = 20;
  const minCameraFar = 1000;
  const sphereRadiusScale = 10;
  const sphereOpacity = 0.08;
  const sphereOutlineOpacity = 0.28;
  const sphereSegmentsWidth = 32;
  const sphereSegmentsHeight = 24;
  const ambientLightColor = "#ffffff";
  const ambientLightIntensity = 0.8;
  const directionalLightColor = "#ffffff";
  const directionalLightIntensity = 1.2;
  const directionalLightPosition = { x: 1, y: 1, z: 1 };
  const minCanvasWidth = 320;
  const minCanvasHeight = 260;
  const canvasAspectRatio = 0.56;

  let mountEl: HTMLDivElement;
  let renderer: THREE.WebGLRenderer | null = null;
  let scene: THREE.Scene | null = null;
  let camera: THREE.PerspectiveCamera | null = null;
  let controls: OrbitControls | null = null;
  let pointCloud: THREE.Points | null = null;
  let hilbertPointCloud: THREE.Points | null = null;
  let axesHelper: THREE.AxesHelper | null = null;
  let radiusSphere: THREE.Mesh | null = null;
  let pointTexture: THREE.CanvasTexture | null = null;
  let resizeObserver: ResizeObserver | null = null;
  let frameId = 0;
  let systemCount = 0;
  let hoverText = "Orbit around the centroid of the current systems cloud";
  let yaw = 0;
  let pitch = 0;
  let zoom = 1;

  function disposeSceneObjects() {
    if (pointCloud) {
      pointCloud.geometry.dispose();
      const material = pointCloud.material;
      if (Array.isArray(material)) {
        material.forEach((entry) => entry.dispose());
      } else {
        material.dispose();
      }
      scene?.remove(pointCloud);
      pointCloud = null;
    }

    if (hilbertPointCloud) {
      hilbertPointCloud.geometry.dispose();
      const material = hilbertPointCloud.material;
      if (Array.isArray(material)) {
        material.forEach((entry) => entry.dispose());
      } else {
        material.dispose();
      }
      scene?.remove(hilbertPointCloud);
      hilbertPointCloud = null;
    }

    if (axesHelper) {
      scene?.remove(axesHelper);
      axesHelper = null;
    }

    if (radiusSphere) {
      radiusSphere.geometry.dispose();
      const material = radiusSphere.material;
      if (Array.isArray(material)) {
        material.forEach((entry) => entry.dispose());
      } else {
        material.dispose();
      }
      scene?.remove(radiusSphere);
      radiusSphere = null;
    }
  }

  function getPointTexture() {
    if (pointTexture) {
      return pointTexture;
    }

    const size = pointTextureSize;
    const textureCanvas = document.createElement("canvas");
    textureCanvas.width = size;
    textureCanvas.height = size;
    const ctx = textureCanvas.getContext("2d");
    if (!ctx) {
      return null;
    }

    const gradient = ctx.createRadialGradient(
      size / 2,
      size / 2,
      pointTextureInnerRadius,
      size / 2,
      size / 2,
      size / 2,
    );
    gradient.addColorStop(0, "rgba(255,255,255,1)");
    gradient.addColorStop(0.45, "rgba(255,255,255,0.95)");
    gradient.addColorStop(1, "rgba(255,255,255,0)");

    ctx.fillStyle = gradient;
    ctx.beginPath();
    ctx.arc(size / 2, size / 2, size / 2, 0, Math.PI * 2);
    ctx.fill();

    pointTexture = new THREE.CanvasTexture(textureCanvas);
    return pointTexture;
  }

  function updateStats() {
    if (!camera || !controls) {
      yaw = 0;
      pitch = 0;
      zoom = 1;
      return;
    }

    const offset = camera.position.clone().sub(controls.target);
    const radius = offset.length() || 1;
    yaw = Math.atan2(offset.x, offset.z);
    pitch = Math.asin(THREE.MathUtils.clamp(offset.y / radius, -1, 1));
    zoom = controls.target.distanceTo(camera.position);
  }

  function fitCamera(points: THREE.Vector3[]) {
    if (!camera || !controls || !renderer || points.length === 0) {
      return;
    }

    const box = new THREE.Box3().setFromPoints(points);
    const size = box.getSize(new THREE.Vector3());
    const radius = Math.max(size.x, size.y, size.z, 1);
    const distance = radius * cameraDistanceFactor;
    const target = new THREE.Vector3(center.x, center.y, center.z);

    controls.target.copy(target);
    camera.position.set(
      target.x + distance,
      target.y + distance * cameraHeightFactor,
      target.z + distance,
    );
    camera.near = Math.max(0.1, distance / cameraMinNearFactor);
    camera.far = Math.max(minCameraFar, distance * cameraFarFactor);
    camera.updateProjectionMatrix();
    controls.update();
    updateStats();
  }

  function setSystemsData() {
    disposeSceneObjects();

    if (!scene) {
      return;
    }

    systemCount = systems.length;
    hoverText =
      systemCount === 0
        ? "No systems returned for this sample"
        : "Orbit around the centroid of the current systems cloud";

    if (systems.length === 0 && hilbertPoints.length === 0) {
      return;
    }

    const positions = new Float32Array(systems.length * 3);
    const colors = new Float32Array(systems.length * 3);
    const vectors: THREE.Vector3[] = [];

    systems.forEach((system, index) => {
      const point = new THREE.Vector3(system.x, system.y, system.z);
      vectors.push(point);

      positions[index * 3] = point.x;
      positions[index * 3 + 1] = point.y;
      positions[index * 3 + 2] = point.z;

      const color =
        system.star_class <= 0x16
          ? new THREE.Color(scoopableColor)
          : new THREE.Color(nonScoopableColor);

      colors[index * 3] = color.r;
      colors[index * 3 + 1] = color.g;
      colors[index * 3 + 2] = color.b;
    });

    const geometry = new THREE.BufferGeometry();
    geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute("color", new THREE.BufferAttribute(colors, 3));

    const material = new THREE.PointsMaterial({
      size: pointSize,
      sizeAttenuation: true,
      vertexColors: true,
      transparent: true,
      opacity: pointOpacity,
      alphaTest: pointAlphaTest,
      map: getPointTexture() ?? undefined,
    });

    pointCloud = new THREE.Points(geometry, material);
    scene.add(pointCloud);

    if (hilbertPoints.length > 0) {
      const hilbertPositions = new Float32Array(hilbertPoints.length * 3);
      const hilbertColors = new Float32Array(hilbertPoints.length * 3);

      hilbertPoints.forEach((point, index) => {
        hilbertPositions[index * 3] = point[0];
        hilbertPositions[index * 3 + 1] = point[1];
        hilbertPositions[index * 3 + 2] = point[2];

        const color = new THREE.Color(hilbertPointColor);
        hilbertColors[index * 3] = color.r;
        hilbertColors[index * 3 + 1] = color.g;
        hilbertColors[index * 3 + 2] = color.b;
      });

      const hilbertGeometry = new THREE.BufferGeometry();
      hilbertGeometry.setAttribute(
        "position",
        new THREE.BufferAttribute(hilbertPositions, 3),
      );
      hilbertGeometry.setAttribute(
        "color",
        new THREE.BufferAttribute(hilbertColors, 3),
      );

      const hilbertMaterial = new THREE.PointsMaterial({
        size: hilbertPointSize,
        sizeAttenuation: true,
        vertexColors: true,
        transparent: true,
        opacity: pointOpacity,
        alphaTest: pointAlphaTest,
        map: getPointTexture() ?? undefined,
      });

      hilbertPointCloud = new THREE.Points(hilbertGeometry, hilbertMaterial);
      hilbertPointCloud.renderOrder = 3;
      scene.add(hilbertPointCloud);
    }

    const box = new THREE.Box3().setFromPoints(vectors);
    const size = box.getSize(new THREE.Vector3());
    const helperSize = Math.max(size.x, size.y, size.z, 1) * axisScaleFactor;
    axesHelper = new THREE.AxesHelper(helperSize);
    axesHelper.position.set(center.x, center.y, center.z);
    scene.add(axesHelper);

    if (radius > 0) {
      const sphereGeometry = new THREE.SphereGeometry(
        radius * sphereRadiusScale,
        sphereSegmentsWidth,
        sphereSegmentsHeight,
      );
      const sphereMaterial = new THREE.MeshBasicMaterial({
        color: scoopableColor,
        transparent: true,
        opacity: sphereOpacity,
        side: THREE.DoubleSide,
        depthWrite: false,
        wireframe: false,
      });
      radiusSphere = new THREE.Mesh(sphereGeometry, sphereMaterial);
      radiusSphere.position.set(center.x, center.y, center.z);
      radiusSphere.renderOrder = 1;
      scene.add(radiusSphere);

      const sphereEdges = new THREE.EdgesGeometry(sphereGeometry);
      const sphereOutline = new THREE.LineSegments(
        sphereEdges,
        new THREE.LineBasicMaterial({
          color: scoopableColor,
          transparent: true,
          opacity: sphereOutlineOpacity,
          depthWrite: false,
        }),
      );
      sphereOutline.renderOrder = 2;
      radiusSphere.add(sphereOutline);
    }

    const fitVectors = vectors.concat(
      hilbertPoints.map((point) => new THREE.Vector3(point[0], point[1], point[2])),
    );
    fitCamera(fitVectors);
  }

  function render() {
    if (!renderer || !scene || !camera || !controls) {
      return;
    }

    controls.update();
    updateStats();
    renderer.render(scene, camera);
    frameId = requestAnimationFrame(render);
  }

  function resizeRenderer() {
    if (!mountEl || !renderer || !camera) {
      return;
    }

    const width = Math.max(mountEl.clientWidth, minCanvasWidth);
    const height = Math.max(
      Math.round(width * canvasAspectRatio),
      minCanvasHeight,
    );
    renderer.setSize(width, height, false);
    camera.aspect = width / height;
    camera.updateProjectionMatrix();
  }

  onMount(() => {
    scene = new THREE.Scene();
    scene.background = new THREE.Color(backgroundColor);

    camera = new THREE.PerspectiveCamera(55, 1, 0.1, 100000000);
    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false });
    renderer.setPixelRatio(window.devicePixelRatio);
    mountEl.appendChild(renderer.domElement);

    controls = new OrbitControls(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.08;
    controls.rotateSpeed = 0.7;
    controls.zoomSpeed = 0.9;
    controls.panSpeed = 0.7;

    const ambient = new THREE.AmbientLight(
      ambientLightColor,
      ambientLightIntensity,
    );
    const directional = new THREE.DirectionalLight(
      directionalLightColor,
      directionalLightIntensity,
    );
    directional.position.set(
      directionalLightPosition.x,
      directionalLightPosition.y,
      directionalLightPosition.z,
    );
    scene.add(ambient, directional);

    resizeRenderer();
    setSystemsData();

    resizeObserver = new ResizeObserver(() => resizeRenderer());
    resizeObserver.observe(mountEl);

    render();
  });

  onDestroy(() => {
    cancelAnimationFrame(frameId);
    resizeObserver?.disconnect();
    controls?.dispose();
    disposeSceneObjects();
    pointTexture?.dispose();
    pointTexture = null;
    renderer?.dispose();
    renderer?.domElement.remove();
  });

  $: if (scene && renderer && camera && controls) {
    systems;
    hilbertPoints;
    center;
    radius;
    setSystemsData();
    resizeRenderer();
  }
</script>

<section class="cloud-card">
  <div class="cloud-header">
    <div>
      <h2>{title}</h2>
      <p>
        {systemCount} systems. Drag to orbit, scroll to zoom, right-drag to pan.
      </p>
    </div>
    <div class="cloud-stats">
      <span>yaw {yaw.toFixed(2)}</span>
      <span>pitch {pitch.toFixed(2)}</span>
      <span>dist {zoom.toFixed(1)}</span>
    </div>
  </div>

  <div bind:this={mountEl} class="canvas-host"></div>

  <div class="cloud-footer">
    <span>{hoverText}</span>
  </div>
</section>

<style>
  .cloud-card {
    padding: 1rem;
    border: 1px solid rgba(255, 120, 0, 0.18);
    background: rgba(10, 12, 18, 0.92);
    border-radius: 0.75rem;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
  }

  .cloud-header,
  .cloud-stats,
  .cloud-footer {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    align-items: center;
  }

  h2 {
    margin: 0 0 0.25rem;
    color: var(--ed-orange-bright);
    font-size: 1.25rem;
  }

  p,
  .cloud-stats,
  .cloud-footer {
    margin: 0;
    color: var(--ed-text-secondary);
    font-size: 0.9rem;
  }

  .cloud-stats {
    flex-wrap: wrap;
  }

  .canvas-host {
    width: 100%;
    min-height: 260px;
    margin: 1rem 0 0.75rem;
    border-radius: 0.6rem;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.06);
  }

  .canvas-host :global(canvas) {
    display: block;
    width: 100%;
    height: auto;
  }

  @media (max-width: 720px) {
    .cloud-header,
    .cloud-footer {
      flex-direction: column;
      align-items: flex-start;
    }
  }
</style>
