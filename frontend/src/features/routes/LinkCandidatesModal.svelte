<script lang="ts">
  import Modal from "../../components/Modal.svelte";
  import Button from "../../components/Button.svelte";
  import Card from "../../components/Card.svelte";
  import Table from "../../components/Table.svelte";
  import CircleFilled from "../../components/icons/CircleFilled.svelte";
  import CircleHollow from "../../components/icons/CircleHollow.svelte";
  import {
    wouldCycle,
    type EditViewRoute,
    type LinkCandidate,
  } from "../../lib/routes/edit";
  import { models } from "../../../wailsjs/go/models";

  export let linkCandidates: LinkCandidate[] = [];
  export let open: boolean = false;
  export let direction: "from" | "to";
  export let routes: EditViewRoute[];
  export let currentRouteId: string;
  export let currentRouteIdx: number;
  export let currentJumpIndex: number;
  export let onSelect: (routeId: string, jumpIndex: number) => void;
  export let onClose: () => void;

  const CONTEXT_BEFORE = 3;
  const CONTEXT_AFTER = 2;

  $: candidates = (linkCandidates || [])
    .filter((candidate) => candidate.route.id !== currentRouteId)
    .map((candidate) => {
      const from =
        direction === "from"
          ? { route_id: currentRouteId, jump_index: currentJumpIndex }
          : { route_id: candidate.route.id, jump_index: candidate.jumpIndex };
      const to =
        direction === "to"
          ? { route_id: currentRouteId, jump_index: currentJumpIndex }
          : { route_id: candidate.route.id, jump_index: candidate.jumpIndex };

      // We can mutate this to save on computation because when we do eventually
      // create a link the parent ExpeditionEdit will re-fetch and re-compute
      // all the EditViewRoute and dependents, including LinkCandidate's
      candidate.wouldCycle = wouldCycle(
        models.Link.createFrom({
          id: "",
          from,
          to,
        }),
        routes,
      );

      return candidate;
    });

  function getContext(candidate: LinkCandidate) {
    const contextStart = Math.max(0, candidate.jumpIndex - CONTEXT_BEFORE);
    const contextEnd = Math.min(
      candidate.route.jumps.length,
      candidate.jumpIndex + CONTEXT_AFTER + 1,
    );

    const context = [];
    for (let i = contextStart; i < contextEnd; i++) {
      context.push({
        index: i,
        jump: candidate.route.jumps[i],
        isMatch: i === candidate.jumpIndex,
      });
    }

    return context;
  }

  function handleSelect(candidate: LinkCandidate) {
    onSelect(candidate.route.id, candidate.jumpIndex);
    onClose();
  }
</script>

<Modal
  {open}
  onRequestClose={onClose}
  title="Create Link {direction === 'from'
    ? 'From'
    : 'To'}: Route {currentRouteIdx + 1}"
>
  <div class="candidates-container">
    <div class="candidates-list stack-md">
    {#if candidates.length === 0}
      <p class="no-candidates">No matching systems found in other routes</p>
    {:else}
      {#each candidates as candidate}
        <div class:cycle-warning={candidate.wouldCycle}>
          <Card>
          <div class="candidate-header">
            <span class="route-label">Route {candidate.routeIdx + 1}</span>
            <span class="route-name">{candidate.route.name}</span>
          </div>
          {#if candidate.wouldCycle}
            <div class="cycle-warning-banner">
              <span class="warning-icon">⚠️</span>
              <span class="warning-text"
                >Creating this link will form a cycle - the expedition will loop
                indefinitely</span
              >
            </div>
          {/if}
          <Table
            columns={[
              { name: "#", align: "left" },
              { name: "System", align: "left" },
              { name: "Scoopable", align: "center" },
              { name: "Distance (LY)", align: "right" },
              { name: "Fuel", align: "right" },
            ]}
            data={getContext(candidate)}
            compact={true}
            let:item
          >
            <tr class:highlight={item.isMatch}>
              <td class="align-left">{item.index + 1}</td>
              <td class="align-left">{item.jump.system_name}</td>
              <td class="align-center">
                <span
                  class="scoopable"
                  class:must-refuel={item.jump.must_refuel}
                  class:can-scoop={item.jump.scoopable}
                >
                  {#if item.jump.scoopable}
                    <CircleFilled size="1rem" />
                  {:else}
                    <CircleHollow size="1rem" />
                  {/if}
                </span>
              </td>
              <td class="align-right numeric"
                >{item.jump.distance.toFixed(2)}</td
              >
              <td class="align-right numeric">
                {#if item.jump.fuel_in_tank !== undefined}
                  {item.jump.fuel_in_tank.toFixed(2)}
                {:else}
                  -
                {/if}
              </td>
            </tr>
          </Table>
          <div class="candidate-actions">
            <Button
              variant="primary"
              size="small"
              onClick={() => handleSelect(candidate)}
            >
              Select Jump {candidate.jumpIndex + 1}
            </Button>
          </div>
        </Card>
        </div>
      {/each}
    {/if}
  </div>

    <div class="modal-footer">
      <Button variant="secondary" onClick={onClose}>Cancel</Button>
    </div>
  </div>
</Modal>

<style>
  .candidates-container {
    max-width: 800px;
    width: 90vw;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
  }

  .candidates-list {
    flex: 1;
    overflow-y: auto;
    min-height: 0;
    padding: 1.5rem;
  }

  .no-candidates {
    color: var(--ed-text-dim);
    font-style: italic;
    text-align: center;
    padding: 2rem;
  }

  .candidate-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .route-label {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--ed-orange);
    text-transform: uppercase;
  }

  .route-name {
    font-size: 0.875rem;
    color: var(--ed-text-primary);
  }

  .candidate-actions {
    margin-top: 0.75rem;
    display: flex;
    justify-content: flex-end;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    padding: 1rem 1.5rem;
    border-top: 1px solid var(--ed-border);
    flex-shrink: 0;
  }

  .scoopable {
    color: var(--ed-text-dim);
    display: inline-flex;
    align-items: center;
  }

  .scoopable.must-refuel {
    color: var(--ed-orange);
  }

  .cycle-warning-banner {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem;
    margin-bottom: 0.75rem;
    background: rgba(255, 165, 0, 0.1);
    border: 1px solid var(--ed-orange);
    border-radius: 4px;
  }

  .warning-icon {
    font-size: 1.25rem;
    flex-shrink: 0;
  }

  .warning-text {
    font-size: 0.875rem;
    color: var(--ed-orange);
    line-height: 1.4;
  }

  .cycle-warning > :global(*) {
    border-color: var(--ed-orange) !important;
  }
</style>
