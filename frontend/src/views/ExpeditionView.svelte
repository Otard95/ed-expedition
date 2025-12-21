<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { models } from "../../wailsjs/go/models";
  import { LoadExpedition } from "../../wailsjs/go/main/App";
  import Card from "../components/Card.svelte";
  import Button from "../components/Button.svelte";
  import ExpeditionStatusBadge from "../components/ExpeditionStatusBadge.svelte";
  import Arrow from "../components/icons/Arrow.svelte";

  export let params: { id: string };

  let expedition: models.Expedition | null = null;
  let loading = true;
  let error: string | null = null;

  onMount(async () => {
    try {
      expedition = await LoadExpedition(params.id);
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load expedition";
      console.error("[ExpeditionView] Failed to load expedition:", err);
    } finally {
      loading = false;
    }
  });

  $: onRouteJumps =
    expedition?.jump_history.filter((j) => j.baked_index !== undefined) ?? [];
  $: detourJumps =
    expedition?.jump_history.filter((j) => j.baked_index === undefined) ?? [];
  $: totalDistance =
    expedition?.jump_history.reduce((sum, j) => sum + (j.distance || 0), 0) ?? 0;
  $: totalFuelUsed =
    expedition?.jump_history.reduce((sum, j) => sum + (j.fuel_used || 0), 0) ?? 0;
</script>

{#if loading}
  <div class="loading-state flex-center">
    <p class="text-secondary">Loading expedition...</p>
  </div>
{:else if error}
  <div class="error-state stack-md flex-center">
    <p class="text-danger">Error: {error}</p>
    <Button variant="secondary" size="small" onClick={() => push("/")}>
      <Arrow direction="left" size="0.75rem" /> Back to Index
    </Button>
  </div>
{:else if expedition}
  <div class="expedition-view stack-lg">
    <div class="header flex-between">
      <div class="title-section">
        <h1 class="text-uppercase-tracked">{expedition.name || "Unnamed Expedition"}</h1>
        <ExpeditionStatusBadge status={expedition.status} />
      </div>
      <Button variant="secondary" size="small" onClick={() => push("/")}>
        <Arrow direction="left" size="0.75rem" /> Back to Index
      </Button>
    </div>

    <Card>
      <div class="stats-grid">
        <div class="stat">
          <div class="stat-label text-uppercase-tracked">Total Jumps</div>
          <div class="stat-value">{expedition.jump_history.length}</div>
        </div>
        <div class="stat">
          <div class="stat-label text-uppercase-tracked">On Route</div>
          <div class="stat-value">{onRouteJumps.length}</div>
        </div>
        <div class="stat">
          <div class="stat-label text-uppercase-tracked">Detours</div>
          <div class="stat-value">{detourJumps.length}</div>
        </div>
        <div class="stat">
          <div class="stat-label text-uppercase-tracked">Total Distance</div>
          <div class="stat-value">{totalDistance.toFixed(2)} LY</div>
        </div>
        <div class="stat">
          <div class="stat-label text-uppercase-tracked">Fuel Used</div>
          <div class="stat-value">{totalFuelUsed.toFixed(2)} T</div>
        </div>
      </div>
    </Card>

    <Card>
      <h2 class="section-title text-uppercase-tracked">Jump History</h2>
      <div class="jump-history">
        {#if expedition.jump_history.length === 0}
          <p class="empty-state text-dim">No jumps recorded</p>
        {:else}
          <div class="jump-list stack-sm">
            {#each expedition.jump_history as jump, i}
              <div class="jump-entry" class:detour={jump.baked_index === undefined}>
                <div class="jump-number text-dim">{i + 1}</div>
                <div class="jump-details">
                  <div class="jump-system">{jump.system_name}</div>
                  <div class="jump-stats text-secondary">
                    {#if jump.distance}
                      <span>{jump.distance.toFixed(2)} LY</span>
                    {/if}
                    {#if jump.fuel_used}
                      <span>• {jump.fuel_used.toFixed(2)} T fuel</span>
                    {/if}
                    {#if jump.baked_index === undefined}
                      <span class="detour-badge text-uppercase-tracked text-dim">Detour</span>
                    {/if}
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </Card>
  </div>
{:else}
  <div class="not-found stack-md flex-center">
    <p class="text-secondary">Expedition not found</p>
    <Button variant="secondary" size="small" onClick={() => push("/")}>
      <Arrow direction="left" size="0.75rem" /> Back to Index
    </Button>
  </div>
{/if}

<style>
  h1 {
    margin: 0;
    font-size: 2rem;
    font-weight: 600;
    color: var(--ed-orange);
  }

  .title-section {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .loading-state,
  .error-state,
  .not-found {
    padding: 4rem 2rem;
    text-align: center;
  }

  .loading-state p {
    font-style: italic;
  }


  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 2rem;
    padding: 1.5rem;
  }

  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
  }

  .stat-label {
    color: hsl(from var(--ed-orange) h s calc(l * 0.7));
    font-size: 0.75rem;
  }

  .stat-value {
    color: var(--ed-text-primary);
    font-size: 1.75rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    white-space: nowrap;
  }

  .section-title {
    margin: 0 0 1rem 0;
    padding: 0 1.5rem;
    font-size: 1rem;
    font-weight: 600;
    color: var(--ed-orange);
  }

  .jump-history {
    padding: 0 1.5rem 1.5rem;
  }

  .empty-state {
    text-align: center;
    padding: 2rem;
    font-style: italic;
  }


  .jump-entry {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.75rem;
    background: var(--ed-bg-tertiary);
    border-radius: 4px;
    border-left: 3px solid var(--ed-orange);
  }

  .jump-entry.detour {
    border-left-color: var(--ed-text-dim);
    opacity: 0.8;
  }

  .jump-number {
    font-size: 0.875rem;
    font-weight: 600;
    min-width: 2rem;
    text-align: right;
  }

  .jump-details {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .jump-system {
    font-weight: 600;
    color: var(--ed-text-primary);
  }

  .jump-stats {
    font-size: 0.875rem;
    display: flex;
    gap: 0.5rem;
    align-items: center;
  }

  .detour-badge {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    background: var(--ed-bg-primary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    font-size: 0.75rem;
  }
</style>
