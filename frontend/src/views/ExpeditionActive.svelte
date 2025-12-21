<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { push } from "svelte-spa-router";
  import {
    LoadActiveExpedition,
    EndActiveExpedition,
  } from "../../wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";
  import { models } from "../../wailsjs/go/models";
  import Card from "../components/Card.svelte";
  import Button from "../components/Button.svelte";
  import Modal from "../components/Modal.svelte";
  import ConfirmDialog from "../components/ConfirmDialog.svelte";
  import ExpeditionStatusBadge from "../components/ExpeditionStatusBadge.svelte";
  import Arrow from "../components/icons/Arrow.svelte";
  import IntersectionObserver from "../components/IntersectionObserver.svelte";
  import Tooltip from "../components/Tooltip.svelte";
  import RouteActiveTable from "../features/routes/RouteActiveTable.svelte";
  import { ActiveJump } from "../lib/routes/active";

  let expedition: models.Expedition | null = null;
  let bakedRoute: models.Route | null = null;
  let loading = true;
  let error: string | null = null;
  let showCompletionModal = false;
  let completedExpedition: models.Expedition | null = null;
  let now = new Date();
  let showEndConfirm = false;
  let endingExpedition = false;

  $: completionStats = completedExpedition
    ? calculateCompletionStats(completedExpedition)
    : null;

  function calculateCompletionStats(exp: models.Expedition) {
    const totalJumps = exp.jump_history.length;
    let onRouteJumps = 0;
    let totalDistance = 0;
    let longestJump = 0;

    for (const jump of exp.jump_history) {
      if (jump.baked_index !== undefined) onRouteJumps++;
      const distance = jump.distance || 0;
      totalDistance += distance;
      if (distance > longestJump) longestJump = distance;
    }

    const detourJumps = totalJumps - onRouteJumps;
    const averageJump = totalJumps > 0 ? totalDistance / totalJumps : 0;
    const routeAccuracy =
      totalJumps > 0 ? (onRouteJumps / totalJumps) * 100 : 0;

    let duration = "Unknown";
    let startDate = "Unknown";
    let endDate = "Unknown";

    if (exp.jump_history.length > 0) {
      const firstJump = new Date(exp.jump_history[0].timestamp);
      const lastJump = new Date(
        exp.jump_history[exp.jump_history.length - 1].timestamp,
      );
      const durationMs = lastJump.getTime() - firstJump.getTime();

      const hours = Math.floor(durationMs / (1000 * 60 * 60));
      const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60));

      if (hours > 0) {
        duration = `${hours}h ${minutes}m`;
      } else {
        duration = `${minutes}m`;
      }

      startDate = firstJump.toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
      endDate = lastJump.toLocaleDateString(undefined, {
        month: "short",
        day: "numeric",
        year: "numeric",
        hour: "2-digit",
        minute: "2-digit",
      });
    }

    return {
      totalJumps,
      onRouteJumps,
      detourJumps,
      totalDistance,
      longestJump,
      averageJump,
      routeAccuracy,
      duration,
      startDate,
      endDate,
    };
  }

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
      ? (
          (expedition.current_baked_index / bakedRoute.jumps.length) *
          100
        ).toFixed(1)
      : "0.0";
  $: jumpsLeft =
    expedition && bakedRoute
      ? bakedRoute.jumps.length - expedition.current_baked_index
      : 0;

  $: startDate = expedition?.jump_history[0]?.timestamp
    ? new Date(expedition.jump_history[0].timestamp).toLocaleDateString(
        undefined,
        {
          month: "short",
          day: "numeric",
          hour: "2-digit",
          minute: "2-digit",
        },
      )
    : null;

  $: duration = (() => {
    if (!expedition?.jump_history.length) return null;
    const firstJump = new Date(expedition.jump_history[0].timestamp);
    const durationMs = now.getTime() - firstJump.getTime();
    const hours = Math.floor(durationMs / (1000 * 60 * 60));
    const minutes = Math.floor((durationMs % (1000 * 60 * 60)) / (1000 * 60));
    return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;
  })();

  $: totalDistance =
    expedition?.jump_history.reduce((sum, j) => sum + (j.distance || 0), 0) ??
    0;

  $: currentJumpIndex = expedition
    ? Math.max(expedition.jump_history.length - 1, 0)
    : 0;

  $: if (currentJumpIndex >= 0 && !loading) {
    setTimeout(() => {
      const currentRow = document.querySelector("tr.current");
      if (currentRow) {
        currentRow.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, 100);
  }

  onMount(async () => {
    try {
      const result = await LoadActiveExpedition();
      expedition = result.Expedition;
      bakedRoute = result.BakedRoute;
    } catch (err) {
      error =
        err instanceof Error ? err.message : "Failed to load active expedition";
      console.error("[ExpeditionActive] Failed to load:", err);
    } finally {
      loading = false;
    }

    EventsOn("JumpHistory", (jumpData: any) => {
      if (expedition) {
        const newJump = models.JumpHistoryEntry.createFrom(jumpData);
        expedition.jump_history = [...expedition.jump_history, newJump];
        expedition.current_baked_index =
          newJump.baked_index ?? expedition.current_baked_index;
      }
    });

    EventsOn("CompleteExpedition", (expeditionData: any) => {
      completedExpedition = models.Expedition.createFrom(expeditionData);
      showCompletionModal = true;
    });

    const durationInterval = setInterval(() => {
      now = new Date();
    }, 10000);

    return () => clearInterval(durationInterval);
  });

  onDestroy(() => {
    EventsOff("JumpHistory");
    EventsOff("CompleteExpedition");
  });

  async function handleEndExpedition() {
    endingExpedition = true;
    try {
      await EndActiveExpedition();
      push("/");
    } catch (err) {
      console.error("[ExpeditionActive] Failed to end expedition:", err);
    } finally {
      endingExpedition = false;
      showEndConfirm = false;
    }
  }
</script>

{#if loading}
  <div class="loading-state flex-center">
    <p class="text-secondary">Loading active expedition...</p>
  </div>
{:else if error}
  <div class="error-state flex-col flex-gap-md flex-center">
    <p class="text-danger">Error: {error}</p>
    <Button variant="secondary" size="small" onClick={() => push("/")}>
      <Arrow direction="left" size="0.75rem" /> Back to Index
    </Button>
  </div>
{:else if expedition && bakedRoute}
  <div class="expedition-active flex-col flex-gap-lg">
    <div class="header flex-between">
      <div class="title-section">
        <h1 class="text-uppercase-tracked">
          {expedition.name || "Unnamed Expedition"}
        </h1>
        <ExpeditionStatusBadge status={expedition.status} />
      </div>
      <div class="header-actions">
        <Button variant="secondary" size="small" onClick={() => push("/")}>
          <Arrow direction="left" size="0.75rem" /> Back to Index
        </Button>
        <Button
          variant="danger"
          size="small"
          onClick={() => (showEndConfirm = true)}
        >
          End Expedition
        </Button>
      </div>
    </div>

    <IntersectionObserver
      let:ratio
      options={{ threshold: [1], rootMargin: "-9px 0px 0px 0px" }}
      class="stats-card-container"
    >
      <div class:sticky={ratio < 1}>
        <Card>
          <div class="stats">
            <div class="stat-compact">
              <div class="stat-label-small text-uppercase-tracked">
                Progress
              </div>
              <div class="stat-value-compact">
                {progressPercent}%
              </div>
            </div>
            <div class="stat-compact">
              <div class="stat-label-small text-uppercase-tracked">
                Jumps Left
              </div>
              <div class="stat-value-compact">
                {jumpsLeft}
              </div>
            </div>
            <div class="stat-compact">
              <div class="stat-label-small text-uppercase-tracked">
                Jumps <Tooltip
                  text="On Route / Detour / Total"
                  direction="down"
                  nowrap
                  size="0.75rem"
                />
              </div>
              <div class="stat-value-compact">
                {onRouteCount} <span class="slash text-dim">/</span>
                {detourCount} <span class="slash text-dim">/</span>
                {totalJumps}
              </div>
            </div>
            {#if startDate}
              <div class="stat-compact">
                <div class="stat-label-small text-uppercase-tracked">
                  Started
                </div>
                <div class="stat-value-compact small">{startDate}</div>
              </div>
            {/if}
            {#if duration}
              <div class="stat-compact">
                <div class="stat-label-small text-uppercase-tracked">
                  Duration
                </div>
                <div class="stat-value-compact">{duration}</div>
              </div>
            {/if}
            <div class="stat-compact">
              <div class="stat-label-small text-uppercase-tracked">
                Distance
              </div>
              <div class="stat-value-compact">
                {totalDistance.toFixed(1)} LY
              </div>
            </div>
          </div>
        </Card>
      </div>
    </IntersectionObserver>

    <Card>
      <RouteActiveTable jumps={allJumps} currentIndex={currentJumpIndex} />
    </Card>
  </div>
{:else}
  <div class="no-active flex-col flex-gap-md flex-center">
    <p class="text-secondary">No active expedition</p>
    <Button variant="secondary" size="small" onClick={() => push("/")}>
      <Arrow direction="left" size="0.75rem" /> Back to Index
    </Button>
  </div>
{/if}

<Modal
  open={showCompletionModal}
  title="Expedition Complete!"
  showCloseButton={false}
>
  <div class="completion-content flex-col flex-gap-md">
    <div class="celebration-text flex-col flex-gap-md">
      <p class="hype">🎉 Outstanding work, Commander! 🎉</p>
      <p class="text-secondary">
        You've successfully completed your expedition! Your flight data has been
        logged and archived in the expedition database.
      </p>
      <p class="expedition-name">
        {completedExpedition?.name || "Unnamed Expedition"}
      </p>
    </div>

    {#if completionStats}
      <div class="completion-stats-container flex-col flex-gap-md">
        <!-- Time Stats -->
        <div class="stats-group">
          <div class="stats-group-title text-uppercase-tracked">Time</div>
          <div class="stats-group-content">
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Started
              </div>
              <div class="completion-stat-value small">
                {completionStats.startDate}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Ended
              </div>
              <div class="completion-stat-value small">
                {completionStats.endDate}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Duration
              </div>
              <div class="completion-stat-value">
                {completionStats.duration}
              </div>
            </div>
          </div>
        </div>

        <!-- Jump Stats -->
        <div class="stats-group">
          <div class="stats-group-title text-uppercase-tracked">Jumps</div>
          <div class="stats-group-content">
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Total
              </div>
              <div class="completion-stat-value">
                {completionStats.totalJumps}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                On Route
              </div>
              <div class="completion-stat-value">
                {completionStats.onRouteJumps}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Detours
              </div>
              <div class="completion-stat-value">
                {completionStats.detourJumps}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Accuracy
              </div>
              <div class="completion-stat-value">
                {completionStats.routeAccuracy.toFixed(1)}%
              </div>
            </div>
          </div>
        </div>

        <!-- Distance Stats -->
        <div class="stats-group">
          <div class="stats-group-title text-uppercase-tracked">Distance</div>
          <div class="stats-group-content">
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Total
              </div>
              <div class="completion-stat-value">
                {completionStats.totalDistance.toFixed(2)} LY
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Average
              </div>
              <div class="completion-stat-value">
                {completionStats.averageJump.toFixed(2)} LY
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Longest
              </div>
              <div class="completion-stat-value">
                {completionStats.longestJump.toFixed(2)} LY
              </div>
            </div>
          </div>
        </div>
      </div>
    {/if}

    <div class="action-buttons flex-row">
      <Button
        variant="primary"
        onClick={() => {
          if (completedExpedition) {
            push(`/expeditions/${completedExpedition.id}/view`);
          }
        }}
      >
        View Expedition
      </Button>
      <Button variant="secondary" onClick={() => push("/")}>
        Back to Index
      </Button>
    </div>
  </div>
</Modal>

<ConfirmDialog
  bind:open={showEndConfirm}
  title="End Expedition"
  message="Are you sure you want to end <strong>{expedition?.name ||
    'this expedition'}</strong>?"
  warningMessage="This cannot be undone. The expedition will be marked as ended and removed from active tracking."
  confirmLabel="End Expedition"
  confirmVariant="danger"
  loading={endingExpedition}
  onConfirm={handleEndExpedition}
  onCancel={() => (showEndConfirm = false)}
/>

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

  .header-actions {
    display: flex;
    gap: 0.75rem;
  }

  .loading-state,
  .error-state,
  .no-active {
    padding: 4rem 2rem;
    text-align: center;
  }

  .loading-state p {
    font-style: italic;
  }

  :global(.stats-card-container) {
    position: sticky;
    top: 8px;
    z-index: 10;
    transition: all 0.2s ease;
  }

  .stats {
    padding: 1rem 1.5rem;
    display: flex;
    align-items: stretch;
    gap: 3rem;
    justify-content: center;
    transition: all 0.2s ease;
  }

  .sticky .stats {
    padding: 0 1rem;
    gap: 5rem;
  }

  .stat-compact {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: space-between;
    gap: 0.25rem;
    transition: gap 0.2s ease;
  }

  .sticky .stat-compact {
    gap: 0.15rem;
  }

  .stat-label-small {
    color: hsl(from var(--ed-orange) h s calc(l * 0.7));
    font-size: 0.75rem;
    transition: font-size 0.2s ease;
  }

  .stat-value-compact {
    color: var(--ed-text-primary);
    font-size: 1.5rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    transition: font-size 0.2s ease;
  }

  .stat-value-compact.small {
    font-size: 1rem;
  }

  .sticky .stat-value-compact {
    font-size: 1.1rem;
  }

  .sticky .stat-value-compact.small {
    font-size: 0.85rem;
  }

  .sticky .stat-label-small {
    font-size: 0.65rem;
  }

  .slash {
    font-weight: 400;
  }

  .completion-content {
    max-width: 700px;
    text-align: center;
    padding: 1.5rem;
  }

  .hype {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--ed-orange);
    margin: 0;
  }

  .celebration-text p {
    margin: 0;
    line-height: 1.6;
  }

  .expedition-name {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--ed-text-primary);
    margin-top: 0.5rem;
  }

  .action-buttons {
    margin-top: 2rem;
    display: flex;
    gap: 1rem;
    justify-content: center;
  }

  .stats-group {
    background: var(--ed-bg-tertiary);
    border-radius: 4px;
    border: 1px solid var(--ed-border);
    overflow: hidden;
  }

  .stats-group-title {
    padding: 0.5rem 1rem;
    background: hsl(from var(--ed-bg-tertiary) h s calc(l * 0.9));
    border-bottom: 1px solid var(--ed-border);
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--ed-orange);
  }

  .stats-group-content {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
    gap: 1.5rem;
    padding: 1.5rem;
  }

  .completion-stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
  }

  .completion-stat-label {
    color: hsl(from var(--ed-orange) h s calc(l * 0.7));
    font-size: 0.75rem;
  }

  .completion-stat-value {
    color: var(--ed-text-primary);
    font-size: 1.5rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }

  .completion-stat-value.small {
    font-size: 1rem;
  }
</style>
