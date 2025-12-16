<script lang="ts">
  import Card from "../../components/Card.svelte";
  import Button from "../../components/Button.svelte";
  import Badge from "../../components/Badge.svelte";
  import ToggleChevron from "../../components/ToggleChevron.svelte";
  import Table from "../../components/Table.svelte";
  import Arrow from "../../components/Arrow.svelte";
  import Copy from "../../components/Copy.svelte";
  import Checkmark from "../../components/Checkmark.svelte";
  import CircleFilled from "../../components/CircleFilled.svelte";
  import CircleHollow from "../../components/CircleHollow.svelte";
  import ConfirmDialog from "../../components/ConfirmDialog.svelte";
  import Dropdown from "../../components/Dropdown.svelte";
  import DropdownItem from "../../components/DropdownItem.svelte";
  import LinkCandidatesModal from "./LinkCandidatesModal.svelte";
  import { EditViewRoute } from "../../lib/routes/edit";
  import { routeExpansion } from "../../lib/stores/routeExpansion";
  import { onDestroy } from "svelte";
  import { ClipboardSetText } from "../../../wailsjs/runtime/runtime";

  export let route: EditViewRoute;
  export let idx: number;
  export let expeditionId: string;

  export let onGotoJump: (
    route_id: string,
    jump_index: number,
    event: MouseEvent,
  ) => void;

  export let onRouteDeleted: ((routeId: string) => void) | undefined =
    undefined;
  export let onLinkCreated: (() => void) | undefined = undefined;
  export let allRoutes: EditViewRoute[] = [];
  export let collapsed: boolean = false;
  let showDeleteConfirm = false;
  let deleting = false;
  let copiedSystemId: number | null = null;

  let showLinkModal = false;
  let linkModalSystemId: number = 0;
  let linkModalSystemName: string = "";
  let linkModalDirection: "from" | "to" = "from";
  let linkModalCurrentJumpIndex: number = 0;
  let creatingLink = false;

  $: possibleLinkCandidates = getPossibleLinkCandidates(allRoutes);

  function getPossibleLinkCandidates(
    routes: EditViewRoute[],
  ): Record<number, Array<{ route: EditViewRoute; jumpIndex: number }>> {
    const map: Record<
      number,
      Array<{ route: EditViewRoute; jumpIndex: number }>
    > = {};

    routes.forEach((r) => {
      r.jumps.forEach((j, jumpIndex) => {
        if (!map[j.system_id]) {
          map[j.system_id] = [];
        }
        map[j.system_id].push({ route: r, jumpIndex });
      });
    });

    return Object.fromEntries(
      Object.entries(map).filter(([_, candidates]) => candidates.length > 1),
    ) as Record<number, Array<{ route: EditViewRoute; jumpIndex: number }>>;
  }

  function hasLinkCandidates(systemId: number): boolean {
    return !!possibleLinkCandidates[systemId];
  }

  function toggleCollapse() {
    collapsed = !collapsed;
  }

  // Listen for expand commands targeting this route
  const unsubscribe = routeExpansion.subscribe((command) => {
    if (command && command.routeId === route.id && collapsed) {
      collapsed = false;
    }
  });

  onDestroy(() => {
    unsubscribe();
  });

  function handleDeleteClick() {
    showDeleteConfirm = true;
  }

  async function confirmDelete() {
    if (deleting) return;

    deleting = true;
    try {
      const { RemoveRouteFromExpedition } = await import(
        "../../../wailsjs/go/main/App"
      );
      await RemoveRouteFromExpedition(expeditionId, route.id);
      showDeleteConfirm = false;
      if (onRouteDeleted) {
        onRouteDeleted(route.id);
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err);
      alert(`Failed to remove route: ${errorMsg}`);
      console.error("Failed to remove route:", err);
    } finally {
      deleting = false;
    }
  }

  async function copySystemName(systemName: string, systemId: number) {
    try {
      await ClipboardSetText(systemName);
      copiedSystemId = systemId;
      setTimeout(() => {
        copiedSystemId = null;
      }, 1500);
    } catch (err) {
      console.error("Failed to copy system name:", err);
    }
  }

  function openLinkModal(
    systemId: number,
    systemName: string,
    direction: "from" | "to",
    jumpIndex: number,
  ) {
    linkModalSystemId = systemId;
    linkModalSystemName = systemName;
    linkModalDirection = direction;
    linkModalCurrentJumpIndex = jumpIndex;
    showLinkModal = true;
  }

  async function handleLinkSelection(
    selectedRouteId: string,
    selectedJumpIndex: number,
  ) {
    if (creatingLink) return;

    creatingLink = true;
    try {
      const { CreateLink } = await import("../../../wailsjs/go/main/App");

      const from =
        linkModalDirection === "from"
          ? { route_id: route.id, jump_index: linkModalCurrentJumpIndex }
          : { route_id: selectedRouteId, jump_index: selectedJumpIndex };

      const to =
        linkModalDirection === "from"
          ? { route_id: selectedRouteId, jump_index: selectedJumpIndex }
          : { route_id: route.id, jump_index: linkModalCurrentJumpIndex };

      await CreateLink(expeditionId, from, to);
      showLinkModal = false;

      if (onLinkCreated) {
        onLinkCreated();
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err);
      alert(`Failed to create link: ${errorMsg}`);
      console.error("Failed to create link:", err);
    } finally {
      creatingLink = false;
    }
  }
</script>

<Card>
  <div class="route-header">
    <div class="route-info">
      <ToggleChevron {collapsed} onClick={toggleCollapse} />
      <span class="route-number">Route {idx + 1}</span>
      <span class="route-name">{route.name}</span>
      <span class="jump-count">{route.jumps.length} jumps</span>
    </div>
    <div class="route-actions">
      <Button variant="secondary" size="small" onClick={handleDeleteClick}
        >Remove</Button
      >
    </div>
  </div>
  {#if !collapsed}
    <hr />
    <Table
      columns={[
        { name: "#", align: "left" },
        { name: "System", align: "left" },
        { name: "Scoopable", align: "center" },
        { name: "Distance (LY)", align: "right" },
        { name: "Fuel", align: "right" },
        { name: "Link", align: "left" },
      ]}
      data={route.jumps}
      let:item
      let:index
    >
      <tr class:unreachable={!item.reachable}>
        <td class="align-left jump-index" id="jump-{route.id}-{index}"
          >{index + 1}</td
        >
        <td class="align-left">
          <div class="system-name-cell">
            <span>{item.system_name}</span>
            <button
              class="copy-btn"
              class:copied={copiedSystemId === item.system_id}
              on:click={() => copySystemName(item.system_name, item.system_id)}
              title="Copy system name"
            >
              {#if copiedSystemId === item.system_id}
                <Checkmark size="0.875rem" />
              {:else}
                <Copy size="0.875rem" />
              {/if}
            </button>
          </div>
        </td>
        <td class="align-center">
          <span
            class="scoopable"
            class:must-refuel={item.must_refuel}
            class:can-scoop={item.scoopable}
          >
            {#if item.scoopable}
              <CircleFilled size="1rem" />
            {:else}
              <CircleHollow size="1rem" />
            {/if}
          </span>
        </td>
        <td class="align-right numeric">{item.distance.toFixed(2)}</td>
        <td class="align-right numeric fuel-cell">
          {#if item.fuel_in_tank !== undefined && item.fuel_used !== undefined}
            {item.fuel_in_tank.toFixed(2)}
            {#if index !== 0}
              <span class="fuel-used">
                <Arrow
                  direction="down"
                  size="0.7rem"
                  color="hsl(from var(--ed-danger) h calc(s * 0.3) calc(l * 0.7))"
                />
                {item.fuel_used.toFixed(2)}
              </span>
            {/if}
          {:else}
            -
          {/if}
        </td>
        <td class="align-left">
          <div class="links-cell">
            <div class="badges-container">
              {#if item.start}
                <Badge variant="success">Start</Badge>
              {/if}
              {#if item.link}
                <Badge
                  variant={item.link.direction === "in" ? "info" : "warning"}
                  onClick={(e) =>
                    onGotoJump(item.link.other.id, item.link.other.i, e)}
                >
                  <Arrow
                    direction={item.link.direction === "in" ? "left" : "right"}
                    color={item.link.direction === "in"
                      ? "var(--ed-info)"
                      : "var(--ed-orange)"}
                  />
                  Route {item.link.other.label}, Jump {item.link.other.i}
                </Badge>
              {/if}
            </div>
            <div
              class="link-dropdown"
              class:has-candidates={hasLinkCandidates(item.system_id)}
            >
              <Dropdown>
                {#if hasLinkCandidates(item.system_id)}
                  <DropdownItem
                    onClick={() =>
                      openLinkModal(
                        item.system_id,
                        item.system_name,
                        "from",
                        index,
                      )}
                  >
                    Create link from here
                  </DropdownItem>
                  <DropdownItem
                    onClick={() =>
                      openLinkModal(
                        item.system_id,
                        item.system_name,
                        "to",
                        index,
                      )}
                  >
                    Create link to here
                  </DropdownItem>
                {/if}
                <DropdownItem
                  onClick={() =>
                    alert("Link to new route - not implemented yet")}
                >
                  Link to new route
                </DropdownItem>
              </Dropdown>
            </div>
          </div>
        </td>
      </tr>
    </Table>
  {/if}
</Card>

<ConfirmDialog
  bind:open={showDeleteConfirm}
  title="Remove Route"
  message={`Are you sure you want to remove <strong>"${route.name}"</strong> from this expedition?`}
  warningMessage="This will also remove any links involving this route."
  confirmLabel="Remove"
  confirmVariant="danger"
  loading={deleting}
  onConfirm={confirmDelete}
  onCancel={() => (showDeleteConfirm = false)}
/>

<LinkCandidatesModal
  bind:open={showLinkModal}
  systemId={linkModalSystemId}
  systemName={linkModalSystemName}
  direction={linkModalDirection}
  routes={allRoutes}
  currentRouteId={route.id}
  currentRouteIdx={idx}
  currentJumpIndex={linkModalCurrentJumpIndex}
  onSelect={handleLinkSelection}
  onClose={() => (showLinkModal = false)}
/>

<style>
  .route-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .route-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .route-number {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--ed-orange);
    text-transform: uppercase;
  }

  .route-name {
    font-size: 1rem;
    font-weight: 500;
    color: var(--ed-text-primary);
  }

  .unreachable > td:not(:last-child) {
    opacity: 0.3;
  }

  .jump-count {
    font-size: 0.875rem;
    color: var(--ed-text-secondary);
  }

  .route-actions {
    display: flex;
    gap: 0.5rem;
  }

  .jump-index {
    color: var(--ed-text-dim);
    font-variant-numeric: tabular-nums;
  }

  .numeric {
    font-variant-numeric: tabular-nums;
  }

  .scoopable {
    color: var(--ed-text-dim);
    display: inline-flex;
    align-items: center;
  }

  .scoopable.must-refuel {
    color: var(--ed-orange);
  }

  .links-cell {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 0.5rem;
    justify-content: space-between;
  }

  .badges-container {
    display: flex;
    gap: 0.25rem;
  }

  .link-dropdown {
    opacity: 0;
    transition: opacity 0.15s ease;
  }

  .link-dropdown.has-candidates {
    opacity: 1;
  }

  .link-dropdown.has-candidates :global(.toggle) {
    color: var(--ed-orange);
    border-color: var(--ed-orange);
  }

  :global(tr:hover) .link-dropdown {
    opacity: 1;
  }

  .system-name-cell {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .copy-btn {
    background: none;
    border: none;
    color: var(--ed-text-dim);
    cursor: pointer;
    padding: 0.25rem;
    font-size: 1rem;
    line-height: 1;
    transition: color 0.15s ease;
    opacity: 0;
  }

  .system-name-cell:hover .copy-btn {
    opacity: 1;
  }

  .copy-btn:hover {
    color: var(--ed-orange);
  }

  .copy-btn.copied {
    color: var(--ed-success);
    opacity: 1;
  }

  .fuel-cell {
    position: relative;
  }
  .fuel-used {
    position: absolute;
    bottom: 100%;
    right: 0%;
    transform: translate(-50%, 50%);
    background-color: var(--ed-bg-secondary);
    color: var(--ed-text-dim);
    font-size: 0.8rem;
    display: inline-flex;
    align-items: center;
    gap: -1rem;
  }

  :global(tr.highlight) {
    background: rgba(255, 120, 0, 0.3) !important;
    transition: background-color 0.3s ease;
  }

  hr {
    opacity: 0.3;
  }

  @keyframes blink {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.3;
    }
  }

  :global(tr.blink) {
    animation: blink 0.5s ease-in-out 2;
  }
</style>
