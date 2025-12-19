<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { LoadActiveExpedition } from "../../wailsjs/go/main/App";
  import type { models } from "../../wailsjs/go/models";
  import Card from "../components/Card.svelte";
  import Button from "../components/Button.svelte";
  import ExpeditionStatusBadge from "../components/ExpeditionStatusBadge.svelte";
  import Arrow from "../components/icons/Arrow.svelte";
  import RouteActiveTable from "../features/routes/RouteActiveTable.svelte";
  import { ActiveJump } from "../lib/routes/active";

  let expedition: models.Expedition | null = null;
  let bakedRoute: models.Route | null = null;
  let loading = true;
  let error: string | null = null;

  $: allJumps =
    expedition && bakedRoute
      ? [
          ...expedition.jump_history,
          ...bakedRoute.jumps.slice(expedition.current_baked_index + 1),
        ].map((jump) => new ActiveJump(jump, bakedRoute.jumps))
      : [];

  $: onRouteCount =
    expedition?.jump_history.filter((j) => j.baked_index !== undefined)
      .length ?? 0;
  $: detourCount =
    expedition?.jump_history.filter((j) => j.baked_index === undefined)
      .length ?? 0;
  $: totalJumps = onRouteCount + detourCount;
  $: progressPercent =
    expedition && bakedRoute
      ? ((expedition.current_baked_index / bakedRoute.jumps.length) * 100).toFixed(
          1,
        )
      : "0.0";
  $: jumpsLeft =
    expedition && bakedRoute
      ? bakedRoute.jumps.length - expedition.current_baked_index
      : 0;

  onMount(async () => {
    try {
      const result = await LoadActiveExpedition("");
      expedition = result.Expedition;
      bakedRoute = result.BakedRoute;
    } catch (err) {
      error =
        err instanceof Error ? err.message : "Failed to load active expedition";
      console.error("[ExpeditionActive] Failed to load:", err);
    } finally {
      loading = false;
    }
  });
</script>

{#if loading}
  <div class="loading-state flex-center">
    <p>Loading active expedition...</p>
  </div>
{:else if error}
  <div class="error-state flex-center">
    <p>Error: {error}</p>
    <Button variant="secondary" size="small" onClick={() => push("/")}>
      <Arrow direction="left" size="0.75rem" /> Back to Index
    </Button>
  </div>
{:else if expedition && bakedRoute}
  <div class="expedition-active stack-lg">
    <div class="header flex-between">
      <div class="title-section">
        <h1>{expedition.name || "Unnamed Expedition"}</h1>
        <ExpeditionStatusBadge status={expedition.status} />
      </div>
      <Button variant="secondary" size="small" onClick={() => push("/")}>
        <Arrow direction="left" size="0.75rem" /> Back to Index
      </Button>
    </div>

    <Card>
      <div class="stats">
        <div class="stat-compact">
          <div class="stat-label-small">Progress</div>
          <div class="stat-value-compact">
            {progressPercent}%
          </div>
        </div>
        <div class="stat-compact">
          <div class="stat-label-small">Jumps Left</div>
          <div class="stat-value-compact">
            {jumpsLeft}
          </div>
        </div>
        <div class="stat-compact">
          <div class="stat-label-small">On Route / Detour / Total</div>
          <div class="stat-value-compact">
            {onRouteCount} <span class="slash">/</span> {detourCount} <span
              class="slash">/</span
            >
            {totalJumps}
          </div>
        </div>
      </div>
    </Card>

    <Card>
      <RouteActiveTable
        jumps={allJumps}
        currentIndex={Math.max(expedition.jump_history.length - 1, 0)}
      />
    </Card>
  </div>
{:else}
  <div class="no-active flex-center">
    <div class="stack-md" style="align-items: center;">
      <p>No active expedition</p>
      <Button variant="primary" onClick={() => push("/")}>
        Go to Expeditions
      </Button>
    </div>
  </div>
{/if}

<style>
  h1 {
    margin: 0;
    font-size: 2rem;
    font-weight: 600;
    color: var(--ed-orange);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .title-section {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .loading-state,
  .error-state,
  .no-active {
    padding: 4rem 2rem;
    text-align: center;
  }

  .loading-state p {
    color: var(--ed-text-secondary);
    font-style: italic;
  }

  .error-state p,
  .no-active p {
    color: var(--ed-text-secondary);
    margin-bottom: 1rem;
  }

  .stats {
    padding: 1rem 1.5rem;
    display: flex;
    gap: 3rem;
    justify-content: center;
  }

  .stat-compact {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
  }

  .stat-label-small {
    color: hsl(from var(--ed-orange) h s calc(l * 0.7));
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .stat-value-compact {
    color: var(--ed-text-primary);
    font-size: 1.5rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }

  .slash {
    color: var(--ed-text-dim);
    font-weight: 400;
  }
</style>
