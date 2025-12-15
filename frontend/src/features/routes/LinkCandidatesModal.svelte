<script lang="ts">
  import Modal from "../../components/Modal.svelte";
  import Button from "../../components/Button.svelte";
  import Card from "../../components/Card.svelte";
  import Table from "../../components/Table.svelte";
  import CircleFilled from "../../components/CircleFilled.svelte";
  import CircleHollow from "../../components/CircleHollow.svelte";
  import type { EditViewRoute } from "../../lib/routes/edit";

  export let open: boolean = false;
  export let systemId: number;
  export let systemName: string;
  export let direction: 'from' | 'to';
  export let routes: EditViewRoute[];
  export let currentRouteId: string;
  export let currentRouteIdx: number;
  export let onSelect: (routeId: string, jumpIndex: number) => void;
  export let onClose: () => void;

  interface LinkCandidate {
    routeId: string;
    routeName: string;
    routeIdx: number;
    jumpIndex: number;
    totalJumps: number;
    route: EditViewRoute;
  }

  interface CandidateExpansion {
    startOffset: number;
    endOffset: number;
  }

  let expansionState: Record<string, CandidateExpansion> = {};

  $: candidates = findCandidates(routes, currentRouteId, systemId);

  function findCandidates(routes: EditViewRoute[], excludeRouteId: string, targetSystemId: number): LinkCandidate[] {
    const results: LinkCandidate[] = [];

    routes.forEach((route, routeIdx) => {
      if (route.id === excludeRouteId) return;

      route.jumps.forEach((jump, jumpIndex) => {
        if (jump.system_id === targetSystemId) {
          const candidateKey = `${route.id}-${jumpIndex}`;
          if (!expansionState[candidateKey]) {
            expansionState[candidateKey] = { startOffset: 2, endOffset: 2 };
          }

          results.push({
            routeId: route.id,
            routeName: route.name,
            routeIdx: routeIdx + 1,
            jumpIndex,
            totalJumps: route.jumps.length,
            route
          });
        }
      });
    });

    return results;
  }

  function getContext(candidate: LinkCandidate) {
    const candidateKey = `${candidate.routeId}-${candidate.jumpIndex}`;
    const expansion = expansionState[candidateKey] || { startOffset: 2, endOffset: 2 };

    const contextStart = Math.max(0, candidate.jumpIndex - expansion.startOffset);
    const contextEnd = Math.min(candidate.route.jumps.length, candidate.jumpIndex + expansion.endOffset + 1);

    const context = [];
    for (let i = contextStart; i < contextEnd; i++) {
      context.push({
        index: i,
        jump: candidate.route.jumps[i],
        isMatch: i === candidate.jumpIndex
      });
    }

    return context;
  }

  function canExpandUp(candidate: LinkCandidate): boolean {
    const candidateKey = `${candidate.routeId}-${candidate.jumpIndex}`;
    const expansion = expansionState[candidateKey];
    return candidate.jumpIndex - expansion.startOffset > 0;
  }

  function canExpandDown(candidate: LinkCandidate): boolean {
    const candidateKey = `${candidate.routeId}-${candidate.jumpIndex}`;
    const expansion = expansionState[candidateKey];
    return candidate.jumpIndex + expansion.endOffset + 1 < candidate.totalJumps;
  }

  function expandUp(candidate: LinkCandidate) {
    const candidateKey = `${candidate.routeId}-${candidate.jumpIndex}`;
    expansionState[candidateKey].startOffset += 3;
    expansionState = expansionState; // Trigger reactivity
  }

  function expandDown(candidate: LinkCandidate) {
    const candidateKey = `${candidate.routeId}-${candidate.jumpIndex}`;
    expansionState[candidateKey].endOffset += 3;
    expansionState = expansionState; // Trigger reactivity
  }

  function handleSelect(candidate: LinkCandidate) {
    onSelect(candidate.routeId, candidate.jumpIndex);
    onClose();
  }
</script>

<Modal {open} onRequestClose={onClose} title="Create Link {direction === 'from' ? 'From' : 'To'}: Route {currentRouteIdx + 1}" class="link-candidates-modal">
  <div class="candidates-list">
    {#if candidates.length === 0}
      <p class="no-candidates">No matching systems found in other routes</p>
    {:else}
      {#each candidates as candidate}
        <Card class="candidate-card">
          <div class="candidate-header">
            <span class="route-label">Route {candidate.routeIdx}</span>
            <span class="route-name">{candidate.routeName}</span>
          </div>
          {#if canExpandUp(candidate)}
            <button class="expand-btn" on:click={() => expandUp(candidate)}>
              ⋯ more
            </button>
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
                <span class="scoopable" class:must-refuel={item.jump.must_refuel} class:can-scoop={item.jump.scoopable}>
                  {#if item.jump.scoopable}
                    <CircleFilled size="1rem" />
                  {:else}
                    <CircleHollow size="1rem" />
                  {/if}
                </span>
              </td>
              <td class="align-right numeric">{item.jump.distance.toFixed(2)}</td>
              <td class="align-right numeric">
                {#if item.jump.fuel_in_tank !== undefined}
                  {item.jump.fuel_in_tank.toFixed(2)}
                {:else}
                  -
                {/if}
              </td>
            </tr>
          </Table>
          {#if canExpandDown(candidate)}
            <button class="expand-btn" on:click={() => expandDown(candidate)}>
              ⋯ more
            </button>
          {/if}
          <div class="candidate-actions">
            <Button variant="primary" size="small" onClick={() => handleSelect(candidate)}>
              Select Jump {candidate.jumpIndex + 1}
            </Button>
          </div>
        </Card>
      {/each}
    {/if}
  </div>

  <div class="modal-footer">
    <Button variant="secondary" onClick={onClose}>Cancel</Button>
  </div>
</Modal>

<style>
  :global(.link-candidates-modal) {
    max-width: 800px;
    width: 90vw;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
  }

  :global(.link-candidates-modal .modal-body) {
    display: flex;
    flex-direction: column;
    min-height: 0;
    overflow: hidden;
  }

  .candidates-list {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 1rem;
    min-height: 0;
  }

  .no-candidates {
    color: var(--ed-text-dim);
    font-style: italic;
    text-align: center;
    padding: 2rem;
  }

  :global(.candidate-card) {
    padding: 1rem !important;
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
    padding-top: 1rem;
    border-top: 1px solid var(--ed-border);
    flex-shrink: 0;
  }

  .highlight {
    background: rgba(255, 120, 0, 0.2) !important;
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

  .expand-btn {
    width: 100%;
    background: none;
    border: none;
    color: var(--ed-text-dim);
    font-size: 0.75rem;
    padding: 0.25rem;
    cursor: pointer;
    text-align: center;
    transition: color 0.15s ease;
  }

  .expand-btn:hover {
    color: var(--ed-orange);
  }
</style>
