<script lang="ts">
  import { onDestroy } from "svelte";

  export let fraction: number = 0;
  export let color: string = "var(--ed-info)";

  let displayProgress = 0;
  let rafId: number | null = null;
  let startValue = 0;
  let startTime = 0;
  let lastUpdateTime = 0;
  let updateInterval = 100;

  function animLoop() {
    const now = performance.now();
    const elapsed = now - startTime;
    const duration = updateInterval * 1.2;
    const t = Math.min(elapsed / duration, 1);

    displayProgress = startValue + (fraction - startValue) * t;

    if (t < 1) {
      rafId = requestAnimationFrame(animLoop);
    } else {
      rafId = null;
    }
  }

  function onFractionUpdate(f: number) {
    const now = performance.now();
    if (lastUpdateTime > 0) {
      updateInterval = now - lastUpdateTime;
    }
    lastUpdateTime = now;

    if (f < displayProgress) {
      displayProgress = f;
      startValue = f;
    } else {
      startValue = displayProgress;
    }
    startTime = now;

    if (rafId == null) {
      rafId = requestAnimationFrame(animLoop);
    }
  }

  $: onFractionUpdate(fraction);

  onDestroy(() => {
    if (rafId != null) cancelAnimationFrame(rafId);
  });
</script>

<div class="bar" style="--color: {color}; --progress: {Math.min(displayProgress, 1) * 100}%"></div>

<style>
  .bar {
    position: absolute;
    top: 0;
    left: 0;
    height: 4px;
    width: var(--progress);
    background: var(--color);
    z-index: 1;
  }
</style>
