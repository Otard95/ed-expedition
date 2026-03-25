<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { DebugHilbertGroups } from "../../../wailsjs/go/main/App";
  import HilbertSystems3D from "./HilbertSystems3D.svelte";

  type Vec3 = { x: number; y: number; z: number };
  type SystemRecord = {
    id: number;
    name: string;
    x: number;
    y: number;
    z: number;
    star_class: number;
  };

  type RawSystemRecord = {
    id?: number;
    Id?: number;
    name?: string;
    Name?: string;
    x?: number;
    y?: number;
    z?: number;
    star_class?: number;
    Position?: Vec3;
    StarClass?: number;
  };

  type DebugSample = {
    position: Vec3;
    points: Vec3[];
    hilbertPoints: [number, number, number][];
    normalizedPosition: Vec3;
    durationsMs: Record<string, number>;
    radius: number;
    indices: number[];
    sortedIndices: number[];
    groups: number[][];
    avgDiff: number;
    diffs: number[];
    ranges: [number, number][];
    systems: SystemRecord[];
    systemsCount: number;
    numSystemsPreFilter: number;
  };

  type RenderedDot = {
    value: number;
    color: string;
    groupIndex: number;
    left: string;
  };

  type RenderedValue = {
    value: number;
    color: string;
    groupIndex: number;
  };

  type RenderedSample = DebugSample & {
    viewMin: number;
    viewMax: number;
    dots: RenderedDot[];
    values: RenderedValue[];
    rangesRendered: {
      left: string;
      width: string;
      color: string;
      groupIndex: number;
    }[];
  };

  type RawDebugSample = {
    position?: Vec3;
    Position?: Vec3;
    points?: Vec3[];
    Points?: Vec3[];
    hilbertPoints?: number[][];
    HilbertPoints?: number[][];
    durationsMs?: Record<string, number>;
    DurationsMs?: Record<string, number>;
    radius?: number;
    Radius?: number;
    indices?: number[];
    Indices?: number[];
    sortedIndices?: number[];
    SortedIndices?: number[];
    groups?: number[][];
    Groups?: number[][];
    ranges?: [number, number][];
    Ranges?: [number, number][];
    systems?: unknown[];
    Systems?: unknown[];
    numSystemsPreFilter?: number;
    NumSystemsPreFilter?: number;
    avgDiff?: number;
    AvgDiff?: number;
    diffs?: number[];
    Diffs?: number[];
  };

  const palette = [
    "#ef4444",
    "#22c55e",
    "#3b82f6",
    "#f59e0b",
    "#a855f7",
    "#06b6d4",
  ];

  const sampleOrigin = {
    x: -42213.8125,
    y: -29359.8125,
    z: -23405.0,
  };
  const galaxyOrigin = {
    x: -43000,
    y: -30000,
    z: -24000,
  };
  const coordScale = 10;
  const sampleRange = {
    x: 82717.625,
    y: 68878.15625,
    z: 89035.15625,
  };

  const minSharedRangeScale = 0.001;
  const maxSharedRangeScale = 1.1;

  let loading = true;
  let samples: DebugSample[] = [];
  let renderedSamples: RenderedSample[] = [];
  let error = "";
  let sampleCount = 12;
  let radius = 10;
  let sharedRangeZoom = 55;
  let pan = 0;
  let centerX = 0;
  let centerY = 0;
  let centerZ = 0;

  $: computedSharedRange =
    samples.length > 0
      ? Math.max(
          ...samples.map((sample) => {
            const min = sample.sortedIndices[0] ?? 0;
            const max =
              sample.sortedIndices[sample.sortedIndices.length - 1] ?? 0;
            return max - min;
          }),
        )
      : 0;
  $: sharedRangeScale =
    minSharedRangeScale *
    Math.pow(maxSharedRangeScale / minSharedRangeScale, sharedRangeZoom / 100);
  $: sharedRange = Math.max(
    1,
    Math.round(computedSharedRange * sharedRangeScale),
  );

  function centeredUnitRandom(): number {
    return (Math.random() + Math.random()) / 2;
  }

  function sampleAxis(origin: number, range: number): number {
    return origin + range * centeredUnitRandom();
  }

  function samplePosition(): Vec3 {
    return {
      x: sampleAxis(sampleOrigin.x, sampleRange.x),
      y: sampleAxis(sampleOrigin.y, sampleRange.y),
      z: sampleAxis(sampleOrigin.z, sampleRange.z),
    };
  }

  function manualPosition(): Vec3 {
    return {
      x: centerX,
      y: centerY,
      z: centerZ,
    };
  }

  function sampleDots(sample: DebugSample) {
    return sample.groups.flatMap((group, groupIndex) =>
      group.map((value) => ({
        value,
        color: palette[groupIndex % palette.length],
        groupIndex,
      })),
    );
  }

  function dotLeft(value: number, min: number, max: number): string {
    if (min === max) return "50%";
    return `${((value - min) / (max - min)) * 100}%`;
  }

  function rangeLeft(start: number, min: number, max: number): number {
    if (min === max) return 0;
    return ((start - min) / (max - min)) * 100;
  }

  function rangeWidth(
    start: number,
    end: number,
    min: number,
    max: number,
  ): number {
    if (min === max) return 0;
    return ((end - start) / (max - min)) * 100;
  }

  function sliderFill(value: number, min: number, max: number): string {
    const percent = ((value - min) / (max - min)) * 100;
    return `background: linear-gradient(90deg, var(--ed-orange) 0%, var(--ed-orange) ${percent}%, rgba(255, 255, 255, 0.08) ${percent}%, rgba(255, 255, 255, 0.08) 100%);`;
  }

  function normalizePosition(position: Vec3): Vec3 {
    return {
      x: (position.x - galaxyOrigin.x) * coordScale,
      y: (position.y - galaxyOrigin.y) * coordScale,
      z: (position.z - galaxyOrigin.z) * coordScale,
    };
  }

  $: renderedSamples = samples.map((sample) => {
    const min = sample.sortedIndices[0] ?? 0;
    const max = sample.sortedIndices[sample.sortedIndices.length - 1] ?? min;
    const sampleSpan = Math.max(1, max - min);
    const viewSpan = sharedRange;
    const panRange = Math.max(0, sampleSpan - sharedRange);
    const viewMin = min + (panRange * pan) / 100;
    const viewMax = viewMin + viewSpan;

    return {
      ...sample,
      viewMin,
      viewMax,
      dots: sample.groups.flatMap((group, groupIndex) =>
        group.map((value) => ({
          value,
          color: palette[groupIndex % palette.length],
          groupIndex,
          left: dotLeft(value, viewMin, viewMax),
        })),
      ),
      values: sample.groups.flatMap((group, groupIndex) =>
        group.map((value) => ({
          value,
          color: palette[groupIndex % palette.length],
          groupIndex,
        })),
      ),
      rangesRendered: sample.ranges.flatMap((range, groupIndex) => {
        const clippedStart = Math.max(viewMin, range[0]);
        const clippedEnd = Math.min(viewMax, range[1]);
        if (clippedEnd <= clippedStart) {
          return [];
        }

        return [
          {
            left: `${rangeLeft(clippedStart, viewMin, viewMax)}%`,
            width: `${rangeWidth(clippedStart, clippedEnd, viewMin, viewMax)}%`,
            color: palette[groupIndex % palette.length],
            groupIndex,
          },
        ];
      }),
    };
  });

  function normalizeSample(sample: RawDebugSample): DebugSample {
    const systems = (
      (sample.systems ?? sample.Systems ?? []) as RawSystemRecord[]
    ).map((system) => ({
      id: system.id ?? system.Id ?? 0,
      name: system.name ?? system.Name ?? "",
      ...normalizePosition({
        x: system.x ?? system.Position?.x ?? 0,
        y: system.y ?? system.Position?.y ?? 0,
        z: system.z ?? system.Position?.z ?? 0,
      }),
      star_class: system.star_class ?? system.StarClass ?? 0,
    }));

    return {
      position: sample.position ?? sample.Position ?? { x: 0, y: 0, z: 0 },
      points: sample.points ?? sample.Points ?? [],
      hilbertPoints: (
        (sample.hilbertPoints ?? sample.HilbertPoints ?? []) as number[][]
      ).map((point) => [point[0] ?? 0, point[1] ?? 0, point[2] ?? 0]),
      normalizedPosition: { x: 0, y: 0, z: 0 },
      durationsMs: sample.durationsMs ?? sample.DurationsMs ?? {},
      radius: sample.radius ?? sample.Radius ?? 0,
      indices: sample.indices ?? sample.Indices ?? [],
      sortedIndices: sample.sortedIndices ?? sample.SortedIndices ?? [],
      groups: sample.groups ?? sample.Groups ?? [],
      ranges: sample.ranges ?? sample.Ranges ?? [],
      systems,
      systemsCount: systems.length,
      numSystemsPreFilter:
        sample.numSystemsPreFilter ?? sample.NumSystemsPreFilter ?? 0,
      avgDiff: sample.avgDiff ?? sample.AvgDiff ?? 0,
      diffs: sample.diffs ?? sample.Diffs ?? [],
    };
  }

  async function regenerate() {
    loading = true;
    error = "";

    try {
      const positions = [
        manualPosition(),
        ...Array.from({ length: sampleCount }, () => samplePosition()),
      ];
      const rawSamples = await Promise.all(
        positions.map((position) =>
          DebugHilbertGroups(position.x, position.y, position.z, radius),
        ),
      );
      samples = (rawSamples as unknown as RawDebugSample[]).map((sample) => {
        const normalized = normalizeSample(sample);
        const queryPoint =
          normalized.hilbertPoints[normalized.hilbertPoints.length - 1];
        normalized.normalizedPosition = {
          x: queryPoint?.[0] ?? 0,
          y: queryPoint?.[1] ?? 0,
          z: queryPoint?.[2] ?? 0,
        };
        return normalized;
      });
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
      samples = [];
    } finally {
      loading = false;
    }
  }

  onMount(regenerate);
</script>

<div class="debug-view text-left">
  <div class="debug-header">
    <div class="controls">
      <div class="flex-col">
        <button class="back-button" on:click={() => push("/")}>Back</button>
        <label class="coord-control">
          <span>X</span>
          <input type="number" bind:value={centerX} />
        </label>
        <label class="coord-control">
          <span>Y</span>
          <input type="number" bind:value={centerY} />
        </label>
        <label class="coord-control">
          <span>Z</span>
          <input type="number" bind:value={centerZ} />
        </label>
        <label class="coord-control">
          <span>Random samples</span>
          <input type="number" min="0" step="1" bind:value={sampleCount} />
        </label>
      </div>
      <div class="flex-col">
        <label class="radius-control range-control">
          <span>Shared range ({(sharedRangeScale * 100).toFixed(1)}%)</span>
          <input
            type="range"
            min="0"
            max="100"
            step="1"
            bind:value={sharedRangeZoom}
            style={sliderFill(sharedRangeZoom, 0, 100)}
          />
          <div class="slider-scale">
            <span>0.1%</span>
            <span>110%</span>
          </div>
        </label>
        <label class="radius-control range-control">
          <span>Pan ({pan.toString()}%)</span>
          <input
            type="range"
            min="0"
            max="100"
            step="0.1"
            bind:value={pan}
            style={sliderFill(pan, 0, 100)}
          />
          <div class="slider-scale">
            <span>start</span>
            <span>end</span>
          </div>
        </label>
      </div>
      <div class="flex-col">
        <label class="radius-control">
          <span>Radius ({radius.toString()})</span>
          <input
            type="range"
            min="2"
            max="100"
            step="1"
            bind:value={radius}
            style={sliderFill(radius, 2, 100)}
          />
          <div class="slider-scale">
            <span>2</span>
            <span>50</span>
          </div>
        </label>
        <button on:click={regenerate} disabled={loading}>
          {#if loading}Generating...{:else}Regenerate{/if}
        </button>
      </div>
    </div>
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if renderedSamples[0]}
    <HilbertSystems3D
      systems={renderedSamples[0].systems}
      hilbertPoints={renderedSamples[0].hilbertPoints}
      title="Manual sample systems"
      center={renderedSamples[0].normalizedPosition}
      radius={renderedSamples[0].radius}
    />
  {/if}

  <div class="sample-list">
    {#each renderedSamples as sample, sampleIndex}
      <section class="sample-card">
        <div class="sample-meta">
          <div>
            <strong
              >{sampleIndex === 0
                ? "Manual sample"
                : `Sample ${sampleIndex + 1}`}</strong
            >
            <span>
              ({sample.position.x.toFixed(1)}, {sample.position.y.toFixed(1)}, {sample.position.z.toFixed(
                1,
              )})
            </span>
          </div>
          <div>
            <span>radius {sample.radius}</span>
            <span>avg diff {sample.avgDiff.toString()}</span>
            <span>{sample.groups.length} groups</span>
            <span>{sample.numSystemsPreFilter} pre-filter</span>
            <span>{sample.systemsCount} systems</span>
            <span>query {sample.durationsMs.query ?? 0}ms</span>
            <span>total {sample.durationsMs.total ?? 0}ms</span>
          </div>
        </div>

        <div class="number-line">
          <div class="axis"></div>
          {#each sample.rangesRendered as range}
            <div
              class="range-band"
              title={`range ${range.groupIndex + 1}`}
              style={`left: ${range.left}; width: ${range.width}; background: ${range.color};`}
            ></div>
          {/each}
          {#each sample.dots as dot}
            <div
              class="dot"
              title={`group ${dot.groupIndex + 1}: ${dot.value.toString()}`}
              style={`left: ${dot.left}; background: ${dot.color};`}
            ></div>
          {/each}
        </div>

        <div class="sample-values">
          <div>
            <span class="label">sorted</span>
            <code>
              {#each sample.values as item, itemIndex}
                <span class="value-chip" style={`color: ${item.color};`}
                  >{item.value.toString()}</span
                >{#if itemIndex < sample.values.length - 1}
                  <span class="value-separator">,</span>
                {/if}
              {/each}
            </code>
          </div>
          <div>
            <span class="label">diffs</span>
            <code
              >{sample.diffs.map((value) => value.toString()).join(", ")}</code
            >
          </div>
          <div>
            <span class="label">points</span>
            <code
              >{sample.points
                .map(
                  (point, pointIndex) =>
                    `${pointIndex + 1}: (${point.x.toFixed(1)}, ${point.y.toFixed(1)}, ${point.z.toFixed(1)})`,
                )
                .join(", ")}</code
            >
          </div>
          <div>
            <span class="label">hilbert points</span>
            <code
              >{sample.hilbertPoints
                .map(
                  (point, pointIndex) =>
                    `${pointIndex + 1}: (${point[0].toString()}, ${point[1].toString()}, ${point[2].toString()})`,
                )
                .join(", ")}</code
            >
          </div>
          <div>
            <span class="label">durations (ms)</span>
            <code
              >{Object.entries(sample.durationsMs)
                .map(([key, value]) => `${key}: ${value.toString()}`)
                .join(", ")}</code
            >
          </div>
        </div>
      </section>
    {/each}
  </div>
</div>

<style>
  .debug-view {
    min-height: 100vh;
    color: var(--ed-text-primary);
    box-sizing: border-box;
  }

  .debug-header {
    position: sticky;
    top: 0;
    z-index: 10;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 1.5rem;
    padding: 1rem;
    border: 1px solid rgba(255, 120, 0, 0.18);
    border-radius: 0.75rem;
    background: rgba(10, 12, 18, 0.96);
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
    backdrop-filter: blur(10px);
  }

  .controls {
    display: flex;
    align-items: flex-end;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .radius-control {
    display: grid;
    gap: 0.35rem;
    color: var(--ed-text-secondary);
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    min-width: 12rem;
  }

  .radius-control input {
    width: 14rem;
    accent-color: var(--ed-orange);
    margin: 0;
    height: 0.45rem;
    border-radius: 999px;
    appearance: none;
    -webkit-appearance: none;
  }

  .radius-control input::-webkit-slider-runnable-track {
    height: 0.45rem;
    border-radius: 999px;
    background: transparent;
  }

  .radius-control input::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 1rem;
    height: 1rem;
    margin-top: -0.275rem;
    border: 2px solid rgba(255, 255, 255, 0.22);
    border-radius: 999px;
    background: var(--ed-orange-bright);
    box-shadow: 0 0 0.75rem rgba(255, 120, 0, 0.25);
  }

  .radius-control input::-moz-range-track {
    height: 0.45rem;
    border-radius: 999px;
    background: transparent;
  }

  .radius-control input::-moz-range-thumb {
    width: 1rem;
    height: 1rem;
    border: 2px solid rgba(255, 255, 255, 0.22);
    border-radius: 999px;
    background: var(--ed-orange-bright);
    box-shadow: 0 0 0.75rem rgba(255, 120, 0, 0.25);
  }

  .range-control {
    min-width: 14rem;
  }

  .coord-control {
    min-width: 9rem;
  }

  .coord-control input {
    width: 9rem;
    padding: 0.7rem 0.8rem;
    border: 1px solid var(--ed-border);
    border-radius: 0.35rem;
    background: rgba(255, 255, 255, 0.04);
    color: var(--ed-text-primary);
    font: inherit;
  }

  .slider-scale {
    display: flex;
    justify-content: space-between;
    color: var(--ed-text-dim);
    font-size: 0.72rem;
    letter-spacing: 0;
    text-transform: none;
  }

  p {
    margin: 0;
    max-width: 60rem;
    color: var(--ed-text-secondary);
  }

  button {
    border: 1px solid var(--ed-border-accent);
    background: rgba(255, 120, 0, 0.1);
    color: var(--ed-orange-bright);
    padding: 0.75rem 1rem;
    border-radius: 0.35rem;
    font: inherit;
    cursor: pointer;
  }

  button:disabled {
    opacity: 0.6;
    cursor: wait;
  }

  .back-button {
    align-self: flex-start;
  }

  .error {
    color: var(--ed-danger);
    margin: 0 0 1rem;
  }

  .sample-list {
    display: grid;
    gap: 1rem;
  }

  .sample-card {
    padding: 1rem;
    border: 1px solid rgba(255, 120, 0, 0.18);
    background: rgba(10, 12, 18, 0.92);
    border-radius: 0.75rem;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
  }

  .sample-meta {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 0.75rem;
    color: var(--ed-text-secondary);
    font-size: 0.95rem;
  }

  .sample-meta > div {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    align-items: baseline;
  }

  .number-line {
    position: relative;
    height: 3.25rem;
    margin: 1rem 0;
    overflow: hidden;
  }

  .axis {
    position: absolute;
    left: 0;
    right: 0;
    top: 50%;
    height: 2px;
    transform: translateY(-50%);
    background: linear-gradient(
      90deg,
      transparent,
      var(--ed-text-dim),
      transparent
    );
  }

  .range-band {
    position: absolute;
    top: 50%;
    height: 0.8rem;
    border-radius: 999px;
    opacity: 0.28;
    transform: translateY(-50%);
    box-shadow: 0 0 0.5rem rgba(255, 255, 255, 0.05);
  }

  .dot {
    position: absolute;
    top: 50%;
    width: 0.3rem;
    height: 0.3rem;
    border-radius: 999px;
    border: 2px solid rgba(255, 255, 255, 0.22);
    transform: translate(-50%, -50%);
    box-shadow: 0 0 0.3rem rgba(255, 255, 255, 0.08);
  }

  .sample-values {
    display: grid;
    gap: 0.5rem;
    font-size: 0.85rem;
  }

  .sample-values > div {
    display: grid;
    gap: 0.25rem;
  }

  .label {
    color: var(--ed-orange-dim);
    text-transform: uppercase;
    letter-spacing: 0.08em;
    font-size: 0.72rem;
  }

  code {
    display: block;
    padding: 0.5rem 0.65rem;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 0.4rem;
    color: var(--ed-text-primary);
    white-space: pre-wrap;
    word-break: break-word;
  }

  .value-chip {
    font-weight: 700;
  }

  .value-separator {
    color: var(--ed-text-dim);
  }

  @media (max-width: 720px) {
    .debug-view {
      padding: 1rem;
    }

    .debug-header,
    .sample-meta {
      flex-direction: column;
    }

    .controls {
      width: 100%;
      align-items: stretch;
      flex-direction: column;
    }

    .radius-control input {
      width: 100%;
      box-sizing: border-box;
    }
  }
</style>
