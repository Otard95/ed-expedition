<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { models } from "../../wailsjs/go/models";
  import {
    LoadExpedition,
    LoadRoutes,
    RenameExpedition,
    StartExpedition,
  } from "../../wailsjs/go/main/App";
  import ExpeditionStatusBadge from "../components/ExpeditionStatusBadge.svelte";
  import Card from "../components/Card.svelte";
  import Button from "../components/Button.svelte";
  import Arrow from "../components/icons/Arrow.svelte";
  import X from "../components/icons/X.svelte";
  import RouteEditTable from "../features/routes/RouteEditTable.svelte";
  import LinksSection from "../features/links/LinksSection.svelte";
  import AddRouteWizard from "../features/routes/AddRouteWizard.svelte";
  import {
    EditViewLink,
    EditViewRoute,
    calculateReachable,
  } from "../lib/routes/edit";
  import { routeExpansion } from "../lib/stores/routeExpansion";
  import { createRouteCollapseStore } from "../lib/stores/routeCollapseState";

  export let params: { id: string };

  $: collapseStore = createRouteCollapseStore(params.id);

  let showRoutePanel = false;
  let canCloseRoutePanel = true;
  let initialFromSystem: string | undefined = undefined;

  let expedition: models.Expedition | null = null;
  let rawRoutes: models.Route[] = [];
  let loading = true;
  let error: string | null = null;
  let expeditionName = "";
  let savingName = false;

  onMount(async () => {
    console.log("[ExpeditionEdit] onMount - loading expedition:", params.id);
    try {
      expedition = await LoadExpedition(params.id);
      console.log("[ExpeditionEdit] Loaded expedition:", expedition);
      expeditionName = expedition.name || "";
      rawRoutes = await LoadRoutes(params.id);
      console.log("[ExpeditionEdit] Loaded routes:", rawRoutes);
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load expedition";
      console.error("[ExpeditionEdit] Failed to load expedition:", err);
    } finally {
      loading = false;
      console.log("[ExpeditionEdit] Loading complete");
    }
  });

  $: links =
    expedition && rawRoutes.length > 0
      ? expedition.links.map((l) => new EditViewLink(l, rawRoutes))
      : [];

  $: routeIdToIdx = rawRoutes.reduce((acc, r, i) => {
    acc[r.id] = i;
    return acc;
  }, {});

  $: routes =
    expedition && rawRoutes.length > 0
      ? calculateReachable(
          expedition.start,
          rawRoutes.map(
            (r) =>
              new EditViewRoute(
                r,
                expedition.start,
                expedition.links,
                routeIdToIdx,
              ),
          ),
        )
      : [];

  function scrollToJump(routeId: string, jumpIndex: number, event: MouseEvent) {
    // Signal the target route to expand if needed
    routeExpansion.expandRoute(routeId);

    // Wait a tick for the route to expand, then scroll
    setTimeout(() => {
      const element = document.getElementById(`jump-${routeId}-${jumpIndex}`);
      if (element) {
        // Get the parent row element
        const row = element.closest("tr");
        if (row) {
          row.scrollIntoView({ behavior: "smooth", block: "center" });
          // Add both highlight and blink to the target row
          row.classList.add("highlight", "blink");
          setTimeout(() => row.classList.remove("blink"), 1000);
          setTimeout(() => row.classList.remove("highlight"), 2000);
        }
      }
    }, 100);
  }

  async function handleNameBlur() {
    if (!expedition || savingName) return;

    const trimmedName = expeditionName.trim();
    if (trimmedName === expedition.name) return;

    savingName = true;
    try {
      await RenameExpedition(expedition.id, trimmedName);
      expedition = await LoadExpedition(expedition.id);
      expeditionName = expedition.name || "";
    } catch (err) {
      console.error("Failed to rename expedition:", err);
      expeditionName = expedition.name || "";
    } finally {
      savingName = false;
    }
  }

  async function handleRouteAdded() {
    if (!expedition) return;

    handleRoutePanelClose();

    console.log("[ExpeditionEdit] Reloading expedition data...");

    try {
      expedition = await LoadExpedition(expedition.id);
      expeditionName = expedition.name || "";
      rawRoutes = await LoadRoutes(expedition.id);
      console.log("[ExpeditionEdit] Reloaded expedition:", expedition);
      console.log("[ExpeditionEdit] Reloaded routes:", rawRoutes);
    } catch (err) {
      console.error("Failed to reload expedition data:", err);
    }
  }

  async function handleRouteDeleted(routeId: string) {
    if (!expedition) return;

    console.log("[ExpeditionEdit] Route deleted, reloading expedition data...");

    try {
      expedition = await LoadExpedition(expedition.id);
      expeditionName = expedition.name || "";
      rawRoutes = await LoadRoutes(expedition.id);
      console.log("[ExpeditionEdit] Reloaded expedition:", expedition);
      console.log("[ExpeditionEdit] Reloaded routes:", rawRoutes);
    } catch (err) {
      console.error("Failed to reload expedition data:", err);
    }
  }

  async function handleLinkCreated() {
    if (!expedition) return;

    console.log("[ExpeditionEdit] Link created, reloading expedition data...");

    try {
      expedition = await LoadExpedition(expedition.id);
      expeditionName = expedition.name || "";
      rawRoutes = await LoadRoutes(expedition.id);
      console.log("[ExpeditionEdit] Reloaded expedition:", expedition);
      console.log("[ExpeditionEdit] Reloaded routes:", rawRoutes);
    } catch (err) {
      console.error("Failed to reload expedition data:", err);
    }
  }

  function handleLinkToNewRoute(systemName: string) {
    initialFromSystem = systemName;
    showRoutePanel = true;
  }

  function handleRoutePanelClose() {
    showRoutePanel = false;
    initialFromSystem = undefined;
  }

  async function handleStartExpedition() {
    if (!expedition) return;

    try {
      await StartExpedition(expedition.id);
      // Redirect to active expedition view
      push("/active");
    } catch (err) {
      console.error("Failed to start expedition:", err);
      alert(
        `Failed to start expedition: ${err instanceof Error ? err.message : String(err)}`,
      );
    }
  }
</script>

<div class="page-layout" class:panel-open={showRoutePanel}>
  {#if loading}
    <div class="loading-state flex-center">
      <p class="text-secondary">Loading expedition...</p>
    </div>
  {:else if error}
    <div class="error-state flex-col flex-gap-md flex-center">
      <p class="text-danger">Error: {error}</p>
      <Button variant="secondary" size="small" onClick={() => push("/")}>
        <Arrow direction="left" size="0.75rem" /> Back to Index
      </Button>
    </div>
  {:else if expedition}
    <div class="expedition-edit main-scroll flex-col flex-gap-lg">
      <div class="header">
        <Button variant="secondary" size="small" onClick={() => push("/")}>
          <Arrow direction="left" size="0.75rem" /> Back
        </Button>
        <div class="title-section">
          <input
            type="text"
            class="name-input text-uppercase-tracked"
            bind:value={expeditionName}
            on:blur={handleNameBlur}
            placeholder="Unnamed Expedition"
            disabled={savingName}
          />
          <ExpeditionStatusBadge status={expedition.status} />
        </div>
        {#if expedition.status === "planned"}
          <Button
            variant="primary"
            size="medium"
            onClick={handleStartExpedition}
            class="start-button"
          >
            Start this Expedition
          </Button>
        {/if}
      </div>

      <div class="sections flex-col flex-gap-lg">
        <div class="section flex-col flex-gap-md">
          <div class="section-header flex-between">
            <h2 class="text-uppercase-tracked">Routes</h2>
            <Button
              variant="primary"
              size="small"
              onClick={() => (showRoutePanel = true)}>Add Route</Button
            >
          </div>
          {#if routes.length === 0}
            <Card>
              <p class="empty-message">
                No routes added yet. Add a route to begin planning your
                expedition.
              </p>
            </Card>
          {:else}
            <div class="routes-list flex-col flex-gap-md">
              {#each routes as route, idx}
                <RouteEditTable
                  {route}
                  {idx}
                  expeditionId={params.id}
                  onGotoJump={scrollToJump}
                  onRouteDeleted={handleRouteDeleted}
                  onLinkCreated={handleLinkCreated}
                  onLinkToNewRoute={handleLinkToNewRoute}
                  allRoutes={routes}
                  {collapseStore}
                  defaultCollapsed={route.id !== expedition?.start?.route_id}
                />
              {/each}
            </div>
          {/if}
        </div>

        <LinksSection {links} onGotoJump={scrollToJump} />
      </div>
    </div>
  {/if}

  {#if showRoutePanel && expedition}
    <div class="route-panel">
      <div class="panel-header">
        <span class="panel-title text-uppercase-tracked">Add Route</span>
        {#if canCloseRoutePanel}
          <button class="panel-close" on:click={handleRoutePanelClose}>
            <X size="1rem" />
          </button>
        {/if}
      </div>
      <div class="panel-body">
        <AddRouteWizard
          expeditionId={expedition.id}
          bind:canClose={canCloseRoutePanel}
          initialFrom={initialFromSystem}
          onComplete={handleRouteAdded}
          onCancel={handleRoutePanelClose}
        />
      </div>
    </div>
  {/if}
</div>

<style>
  h2 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--ed-orange);
  }

  .header {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .title-section {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex: 1;
  }

  .name-input {
    flex: 1;
    max-width: 600px;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    padding: 0.5rem 0.75rem;
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--ed-text-primary);
  }

  .name-input:focus {
    outline: none;
    border-color: var(--ed-orange);
  }

  .name-input::placeholder {
    color: var(--ed-text-dim);
  }

  .loading-state,
  .error-state {
    padding: 4rem 2rem;
  }

  .loading-state p {
    font-style: italic;
  }

  :global(.start-button) {
    box-shadow: 0 0 16px var(--ed-orange);
  }

  /* Panel layout */

  .page-layout {
    display: contents;
  }

  .page-layout.panel-open {
    display: flex;
    flex-direction: column;
    /* Fills the viewport minus App.svelte's 2rem top + 2rem bottom padding */
    height: calc(100vh - 4rem);
    gap: 0;
  }

  .main-scroll {
    /* Slight inset so scrollbar doesn't clip content flush to edge */
    padding: 0.125rem 0.25rem 1rem;
  }

  .page-layout.panel-open .main-scroll {
    flex: 1;
    overflow-y: auto;
    overscroll-behavior: contain;
  }

  .route-panel {
    flex-shrink: 0;
    height: 420px;
    border-top: 2px solid var(--ed-orange);
    background: var(--ed-bg-secondary);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.625rem 1.25rem;
    border-bottom: 1px solid var(--ed-border);
    flex-shrink: 0;
  }

  .panel-title {
    font-size: 0.8125rem;
    font-weight: 600;
    color: var(--ed-orange);
    letter-spacing: 0.08em;
  }

  .panel-close {
    background: none;
    border: none;
    color: var(--ed-text-secondary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.25rem;
    border-radius: 2px;
    transition: color 0.15s;
  }

  .panel-close:hover {
    color: var(--ed-text-primary);
  }

  .panel-body {
    flex: 1;
    overflow-y: auto;
    overscroll-behavior: contain;
  }

  /* Landscape: panel docks to the right */
  @media (min-aspect-ratio: 1/1) {
    .page-layout.panel-open {
      flex-direction: row;
      align-items: stretch;
    }

    .route-panel {
      border-top: none;
      border-left: 2px solid var(--ed-orange);
      width: min(40%, 520px);
      height: auto;
    }
  }
</style>
