<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { MapScene } from './scene';
  import type { EditViewRoute } from '../../lib/routes/edit';
  import { models } from '../../../wailsjs/go/models';
  import CircleFilled from '../../components/icons/CircleFilled.svelte';
  import CircleHollow from '../../components/icons/CircleHollow.svelte';
  import Neutron from '../../components/icons/Neutron.svelte';
  import FSDInjection from '../../components/icons/FSDInjection.svelte';
  import BoostLevel from '../../components/icons/BoostLevel.svelte';
  import Copy from '../../components/icons/Copy.svelte';
  import Tooltip from '../../components/Tooltip.svelte';
  import Checkmark from '../../components/icons/Checkmark.svelte';
  import X from '../../components/icons/X.svelte';
  import { ClipboardSetText } from '../../../wailsjs/runtime/runtime';
  import { AutocompleteSystems, LookupSystem } from '../../../wailsjs/go/main/App';
  import AutocompleteInput from '../../components/AutocompleteInput.svelte';

  export let routes: EditViewRoute[];

  let container: HTMLDivElement;
  let labelContainer: HTMLDivElement;
  let scene: MapScene | null = null;

  type TooltipData = {
    x: number;
    y: number;
    systemName: string;
    routeName: string;
    scoopable: boolean;
    mustRefuel: boolean;
    distance: number;
    fsdBoost: number | undefined;
  };

  type CustomMarker = { name: string; x: number; y: number; z: number };
  let customMarkers: CustomMarker[] = [];
  let searchValue = '';

  async function handleSystemSelected(name: string) {
    if (!name.trim()) return;
    if (customMarkers.some(m => m.name.toLowerCase() === name.toLowerCase())) {
      searchValue = '';
      return;
    }
    const result = await LookupSystem(name);
    if (!result.found) return;
    customMarkers = [...customMarkers, { name: result.name, x: result.position.x, y: result.position.y, z: result.position.z }];
    searchValue = '';
  }

  function removeMarker(name: string) {
    customMarkers = customMarkers.filter(m => m.name !== name);
  }

  $: if (scene) scene.setCustomMarkers(customMarkers);

  let hoverTooltip: TooltipData | null = null;
  let pinnedTooltip: TooltipData | null = null;
  let copied = false;

  let hoverCloseTimer: ReturnType<typeof setTimeout> | null = null;

  // Captured on mousedown so cursor drift before click fires doesn't lose the hit.
  let pendingPin: TooltipData | null = null;
  let mouseDownX = 0;
  let mouseDownY = 0;

  function makeTooltipData(e: MouseEvent, hit: ReturnType<typeof scene.pick>): TooltipData | null {
    if (!hit) return null;
    if (hit.kind === 'custom') return {
      x: e.offsetX,
      y: e.offsetY,
      systemName: hit.name,
      routeName: '',
      scoopable: false,
      mustRefuel: false,
      distance: 0,
      fsdBoost: undefined,
    };
    return {
      x: e.offsetX,
      y: e.offsetY,
      systemName: hit.jump.system_name,
      routeName: hit.route.name,
      scoopable: hit.jump.scoopable,
      mustRefuel: hit.jump.must_refuel,
      distance: hit.jump.distance,
      fsdBoost: hit.jump.fsd_boost,
    };
  }

  function scheduleHoverClose() {
    if (hoverCloseTimer) clearTimeout(hoverCloseTimer);
    hoverCloseTimer = setTimeout(() => {
      hoverTooltip = null;
      hoverCloseTimer = null;
    }, 150);
  }

  function cancelHoverClose() {
    if (hoverCloseTimer) {
      clearTimeout(hoverCloseTimer);
      hoverCloseTimer = null;
    }
  }

  function handleMouseMove(e: MouseEvent) {
    if (!scene) return;
    const hit = scene.pick(e.clientX, e.clientY, container.getBoundingClientRect());
    if (hit) {
      cancelHoverClose();
      const data = makeTooltipData(e, hit);
      if (!hoverTooltip) {
        // Fresh detection: set position so the tooltip stays put and the
        // cursor can travel to it without the tooltip running away.
        hoverTooltip = data;
      } else {
        // Already showing: update data but keep position frozen.
        hoverTooltip = { ...data, x: hoverTooltip.x, y: hoverTooltip.y };
      }
    } else {
      scheduleHoverClose();
    }
  }

  function handleMouseLeave() {
    scheduleHoverClose();
    pendingPin = null;
  }

  function handleMouseDown(e: MouseEvent) {
    if (!scene) return;
    mouseDownX = e.clientX;
    mouseDownY = e.clientY;
    const hit = scene.pick(e.clientX, e.clientY, container.getBoundingClientRect());
    pendingPin = makeTooltipData(e, hit);
  }

  function handleClick(e: MouseEvent) {
    // Ignore drags (OrbitControls camera rotation/pan)
    const dx = e.clientX - mouseDownX;
    const dy = e.clientY - mouseDownY;
    if (Math.sqrt(dx * dx + dy * dy) > 5) return;

    const data = pendingPin;
    pendingPin = null;

    if (!data) {
      pinnedTooltip = null;
      return;
    }
    if (pinnedTooltip?.systemName === data.systemName) {
      pinnedTooltip = null;
    } else {
      pinnedTooltip = { ...data, x: e.offsetX, y: e.offsetY };
      copied = false;
    }
  }

  async function handleCopy() {
    if (!pinnedTooltip) return;
    try {
      await ClipboardSetText(pinnedTooltip.systemName);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch (err) {
      console.error('Failed to copy system name:', err);
    }
  }

  function unpin() {
    pinnedTooltip = null;
  }

  onMount(() => {
    scene = new MapScene(container, labelContainer);
    scene.load(routes);
    scene.start();

    const observer = new ResizeObserver(() => scene?.resize());
    observer.observe(container);

    return () => {
      observer.disconnect();
    };
  });

  onDestroy(() => {
    scene?.destroy();
    scene = null;
  });

  $: if (scene) {
    scene.load(routes);
  }
</script>

<div class="map-wrapper">
  <div
    class="canvas-container"
    bind:this={container}
    on:mousemove={handleMouseMove}
    on:mouseleave={handleMouseLeave}
    on:mousedown={handleMouseDown}
    on:click={handleClick}
  ></div>
  <div class="label-layer" bind:this={labelContainer}></div>

  <div class="controls-hint">
    <Tooltip direction="up-left" size="1.25rem">
      <div class="hint-rows">
        <div class="hint-row"><kbd>Left drag</kbd><span>Orbit</span></div>
        <div class="hint-row"><kbd>Scroll</kbd><span>Zoom</span></div>
        <div class="hint-row"><kbd>Right drag</kbd><span>Elevation</span></div>
        <div class="hint-row"><kbd>WASD</kbd><span>Pan</span></div>
        <div class="hint-row"><kbd>Hover</kbd><span>System info</span></div>
        <div class="hint-row"><kbd>Click</kbd><span>Pin info</span></div>
      </div>
    </Tooltip>
  </div>

  <div class="markers-bar">
    <div class="markers-search">
      <AutocompleteInput
        bind:value={searchValue}
        placeholder="Add system…"
        fetchSuggestions={AutocompleteSystems}
        onSelect={handleSystemSelected}
        dropUp
      />
    </div>
    {#if customMarkers.length > 0}
      <ul class="markers-list">
        {#each customMarkers as marker}
          <li class="marker-chip">
            <span>{marker.name}</span>
            <button class="marker-remove" on:click={() => removeMarker(marker.name)}>
              <X size="0.625rem" />
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  {#if hoverTooltip && hoverTooltip.systemName !== pinnedTooltip?.systemName}
    <div
      class="tooltip"
      style="left: {hoverTooltip.x + 14}px; top: {hoverTooltip.y + 14}px"
      on:mouseenter={cancelHoverClose}
      on:mouseleave={scheduleHoverClose}
    >
      <span class="tooltip-system">{hoverTooltip.systemName}</span>
      {#if hoverTooltip.routeName}<span class="tooltip-route">{hoverTooltip.routeName}</span>{/if}
      {#if hoverTooltip.distance > 0 || hoverTooltip.fsdBoost != null}
        <div class="tooltip-stats">
          {#if hoverTooltip.distance > 0}<span>{hoverTooltip.distance.toFixed(1)} ly</span>{/if}
          <span class="scoop-indicator" class:must-refuel={hoverTooltip.mustRefuel}>
            {#if hoverTooltip.scoopable}<CircleFilled size="0.75rem" />{:else}<CircleHollow size="0.75rem" />{/if}
          </span>
          {#if hoverTooltip.fsdBoost === models.FSDBoost.NEUTRON}
            <Neutron size="0.875rem" color="var(--ed-orange)" />
          {:else if hoverTooltip.fsdBoost === models.FSDBoost.INJECTION_BASIC || hoverTooltip.fsdBoost === models.FSDBoost.INJECTION_STANDARD || hoverTooltip.fsdBoost === models.FSDBoost.INJECTION_PREMIUM}
            <FSDInjection size="0.875rem" color="var(--ed-orange)" />
            <BoostLevel level={hoverTooltip.fsdBoost === models.FSDBoost.INJECTION_BASIC ? 1 : hoverTooltip.fsdBoost === models.FSDBoost.INJECTION_STANDARD ? 2 : 3} color="var(--ed-orange)" />
          {/if}
        </div>
      {/if}
    </div>
  {/if}

  {#if pinnedTooltip}
    <div class="tooltip tooltip--pinned" style="left: {pinnedTooltip.x + 14}px; top: {pinnedTooltip.y + 14}px">
      <div class="tooltip-header">
        <span class="tooltip-system">{pinnedTooltip.systemName}</span>
        <div class="tooltip-actions">
          <button class="tooltip-btn" class:copied title="Copy system name" on:click|stopPropagation={handleCopy}>
            {#if copied}
              <Checkmark size="0.75rem" />
            {:else}
              <Copy size="0.75rem" />
            {/if}
          </button>
          <button class="tooltip-btn" title="Close" on:click|stopPropagation={unpin}>
            <X size="0.75rem" />
          </button>
        </div>
      </div>
      {#if pinnedTooltip.routeName}<span class="tooltip-route">{pinnedTooltip.routeName}</span>{/if}
      {#if pinnedTooltip.distance > 0 || pinnedTooltip.fsdBoost != null}
        <div class="tooltip-stats">
          {#if pinnedTooltip.distance > 0}<span>{pinnedTooltip.distance.toFixed(1)} ly</span>{/if}
          <span class="scoop-indicator" class:must-refuel={pinnedTooltip.mustRefuel}>
            {#if pinnedTooltip.scoopable}<CircleFilled size="0.75rem" />{:else}<CircleHollow size="0.75rem" />{/if}
          </span>
          {#if pinnedTooltip.fsdBoost === models.FSDBoost.NEUTRON}
            <Neutron size="0.875rem" color="var(--ed-orange)" />
          {:else if pinnedTooltip.fsdBoost === models.FSDBoost.INJECTION_BASIC || pinnedTooltip.fsdBoost === models.FSDBoost.INJECTION_STANDARD || pinnedTooltip.fsdBoost === models.FSDBoost.INJECTION_PREMIUM}
            <FSDInjection size="0.875rem" color="var(--ed-orange)" />
            <BoostLevel level={pinnedTooltip.fsdBoost === models.FSDBoost.INJECTION_BASIC ? 1 : pinnedTooltip.fsdBoost === models.FSDBoost.INJECTION_STANDARD ? 2 : 3} color="var(--ed-orange)" />
          {/if}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .map-wrapper {
    position: relative;
    width: 100%;
    height: 100%;
  }

  .canvas-container {
    width: 100%;
    height: 100%;
  }

  .label-layer {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }

  .tooltip--pinned {
    pointer-events: auto;
    border-color: var(--ed-orange);
    box-shadow: 0 0 8px rgba(255, 120, 0, 0.25);
  }

  .tooltip-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .tooltip-header .tooltip-system {
    flex: 1;
    min-width: 0;
  }

  .tooltip-actions {
    display: flex;
    gap: 0.125rem;
    flex-shrink: 0;
  }

  .tooltip-btn {
    background: none;
    border: none;
    color: var(--ed-text-secondary);
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.125rem 0.25rem;
    border-radius: 2px;
    font-size: 0.6875rem;
    transition: color 0.15s;
  }

  .tooltip-btn:hover {
    color: var(--ed-text-primary);
  }

  .tooltip-btn.copied {
    color: var(--ed-success);
  }

  .tooltip {
    position: absolute;
    pointer-events: auto;
    z-index: 10;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border-accent);
    border-radius: 2px;
    padding: 0.375rem 0.625rem;
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    max-width: 240px;
  }

  .tooltip-system {
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--ed-text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tooltip-route {
    font-size: 0.6875rem;
    color: var(--ed-text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .tooltip-stats {
    display: flex;
    gap: 0.5rem;
    font-size: 0.6875rem;
    color: var(--ed-text-secondary);
    margin-top: 0.1rem;
  }

  .scoop-indicator {
    display: inline-flex;
    align-items: center;
    color: var(--ed-text-dim);
  }

  .scoop-indicator.must-refuel {
    color: var(--ed-orange);
  }

  .controls-hint {
    position: absolute;
    bottom: 0.75rem;
    right: 0.75rem;
    z-index: 10;
  }

  .hint-rows {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
  }

  .hint-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1.5rem;
    font-size: 0.75rem;
  }

  .hint-row span {
    color: var(--ed-text-secondary);
  }

  kbd {
    background: var(--ed-bg-tertiary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    padding: 0.1rem 0.35rem;
    font-family: inherit;
    font-size: 0.6875rem;
    color: var(--ed-text-primary);
    white-space: nowrap;
  }

  .markers-bar {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 10;
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    padding: 0.5rem;
    background: linear-gradient(to top, rgba(0,0,0,0.7) 0%, transparent 100%);
    pointer-events: none;
  }

  .markers-search {
    width: 200px;
    pointer-events: auto;
  }

  /* Override AutocompleteInput sizing to fit the bar */
  .markers-search :global(input) {
    font-size: 0.8125rem;
    padding: 0.3rem 0.6rem;
  }

  .markers-search :global(.error-text) {
    display: none;
  }

  .markers-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
    list-style: none;
    margin: 0;
    padding: 0;
    pointer-events: auto;
    align-self: center;
  }

  .marker-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    background: rgba(10, 10, 10, 0.85);
    border: 1px solid #A0C4FF;
    border-radius: 2px;
    padding: 0.2rem 0.4rem;
    font-size: 0.6875rem;
    color: #A0C4FF;
  }

  .marker-remove {
    background: none;
    border: none;
    color: var(--ed-text-dim);
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    padding: 0;
    transition: color 0.15s;
  }

  .marker-remove:hover {
    color: var(--ed-text-primary);
  }

  /*
   * CSS2DRenderer injects divs directly into .label-layer.
   * :global is needed because these elements are created imperatively by Three.js.
   */
  :global(.map-label) {
    font-family: inherit;
    font-size: 0.6875rem;
    font-weight: 600;
    color: #FF7800;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    text-shadow: 0 0 4px #000, 0 0 8px #000;
    white-space: nowrap;
    pointer-events: none;
    user-select: none;
  }

  :global(.map-label--custom) {
    color: #A0C4FF;
  }
</style>
