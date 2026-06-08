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
  import ConfirmDialog from "../components/ConfirmDialog.svelte";
  import ExpeditionStatusBadge from "../components/ExpeditionStatusBadge.svelte";
  import Arrow from "../components/icons/Arrow.svelte";
  import IntersectionObserver from "../components/IntersectionObserver.svelte";
  import RouteActiveTable from "../features/routes/RouteActiveTable.svelte";
  import ExpeditionCompletionModal from "../features/expeditions/ExpeditionCompletionModal.svelte";
  import ActiveExpeditionStats from "../features/expeditions/ActiveExpeditionStats.svelte";
  import { computeActiveStats } from "../lib/expedition/active";

  let expedition: models.Expedition | null = null;
  let bakedRoute: models.Route | null = null;
  let loading = true;
  let error: string | null = null;
  let showCompletionModal = false;
  let completedExpedition: models.Expedition | null = null;
  let showEndConfirm = false;
  let endingExpedition = false;

  $: stats =
    expedition && bakedRoute
      ? computeActiveStats(expedition, bakedRoute)
      : null;

  $: if (stats && stats.currentJumpIndex >= 0 && !loading) {
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
        <ActiveExpeditionStats {expedition} {stats} compact={ratio < 1} />
      </div>
    </IntersectionObserver>

    <Card>
      <RouteActiveTable
        jumps={stats.allJumps}
        currentIndex={stats.currentJumpIndex}
      />
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

<ExpeditionCompletionModal
  open={showCompletionModal}
  expedition={completedExpedition}
/>

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
</style>
