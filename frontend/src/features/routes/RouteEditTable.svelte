<script lang="ts">
  import Card from "../../components/Card.svelte";
  import Button from "../../components/Button.svelte";
  import Badge from "../../components/Badge.svelte";
  import ToggleChevron from "../../components/ToggleChevron.svelte";
  import Table from "../../components/Table.svelte";
  import Arrow from "../../components/Arrow.svelte";
  import { Icons } from "../../lib/icons";
  import { EditViewRoute } from "../../lib/routes/edit";
  import { routeExpansion } from "../../lib/stores/routeExpansion";
  import { onDestroy } from "svelte";

  export let route: EditViewRoute;
  export let idx: number;

  export let onGotoJump: (
    route_id: string,
    jump_index: number,
    event: MouseEvent,
  ) => void;

  let collapsed = false;

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
      <Button variant="secondary" size="small">Remove</Button>
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
        <td class="align-left">{item.system_name}</td>
        <td class="align-center">
          <span class="scoopable" class:yes={item.scoopable}>
            {item.scoopable ? Icons.SCOOPABLE : Icons.NOT_SCOOPABLE}
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
        </td>
      </tr>
    </Table>
  {/if}
</Card>

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

  .unreachable {
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

  :global(.jump-index) {
    color: var(--ed-text-dim);
    font-variant-numeric: tabular-nums;
  }

  :global(.numeric) {
    font-variant-numeric: tabular-nums;
  }

  :global(.scoopable) {
    font-size: 1.25rem;
    color: var(--ed-text-dim);
  }

  :global(.scoopable.yes) {
    color: var(--ed-orange);
  }

  :global(.links-cell) {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
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
