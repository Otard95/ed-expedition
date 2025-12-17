<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { models } from "../../wailsjs/go/models";
  import {
    LoadExpedition,
    LoadRoutes,
    RenameExpedition,
  } from "../../wailsjs/go/main/App";
  import ExpeditionStatusBadge from "../components/ExpeditionStatusBadge.svelte";
  import Card from "../components/Card.svelte";
  import Button from "../components/Button.svelte";
  import Modal from "../components/Modal.svelte";
  import RouteEditTable from "../features/routes/RouteEditTable.svelte";
  import LinksSection from "../features/links/LinksSection.svelte";
  import AddRouteWizard from "../features/routes/AddRouteWizard.svelte";
  import {
    EditViewLink,
    EditViewRoute,
    calculateReachable,
  } from "../lib/routes/edit";
  import { routeExpansion } from "../lib/stores/routeExpansion";

  export let params: { id: string };

  let showAddRouteModal = false;
  let canCloseAddRouteModal = true;
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

  async function handleRouteAdded(route: models.Route) {
    if (!expedition) return;

    handleAddRouteModalClose();

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
    showAddRouteModal = true;
  }

  function handleAddRouteModalClose() {
    showAddRouteModal = false;
    initialFromSystem = undefined;
  }
</script>

{#if loading}
  <div class="loading-state flex-center">
    <p>Loading expedition...</p>
  </div>
{:else if error}
  <div class="error-state flex-center">
    <p>Error: {error}</p>
  </div>
{:else if expedition}
  <div class="expedition-edit stack-lg">
    <div class="header">
      <Button variant="secondary" size="small" onClick={() => push("/")}>
        ← Back
      </Button>
      <div class="title-section">
        <input
          type="text"
          class="name-input"
          bind:value={expeditionName}
          on:blur={handleNameBlur}
          placeholder="Unnamed Expedition"
          disabled={savingName}
        />
        <ExpeditionStatusBadge status={expedition.status} />
      </div>
    </div>

    <div class="sections stack-lg">
      <div class="section stack-md">
        <div class="section-header flex-between">
          <h2>Routes</h2>
          <Button
            variant="primary"
            size="small"
            onClick={() => (showAddRouteModal = true)}>Add Route</Button
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
          <div class="routes-list stack-md">
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
                collapsed={route.id !== expedition?.start?.route_id}
              />
            {/each}
          </div>
        {/if}
      </div>

      <LinksSection {links} onGotoJump={scrollToJump} />
    </div>
  </div>
{/if}

{#if expedition}
  <Modal
    bind:open={showAddRouteModal}
    title="Add Route"
    onRequestClose={canCloseAddRouteModal
      ? handleAddRouteModalClose
      : undefined}
    showCloseButton={canCloseAddRouteModal}
  >
    <AddRouteWizard
      expeditionId={expedition.id}
      bind:canClose={canCloseAddRouteModal}
      initialFrom={initialFromSystem}
      onComplete={handleRouteAdded}
      onCancel={handleAddRouteModalClose}
    />
  </Modal>
{/if}

<style>
  h2 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--ed-orange);
    text-transform: uppercase;
    letter-spacing: 0.05em;
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
    text-transform: uppercase;
    letter-spacing: 0.05em;
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
    color: var(--ed-text-secondary);
    font-style: italic;
  }

  .error-state p {
    color: var(--ed-danger);
  }
</style>
