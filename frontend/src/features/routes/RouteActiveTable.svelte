<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Table from "../../components/Table.svelte";
  import Copy from "../../components/icons/Copy.svelte";
  import Checkmark from "../../components/icons/Checkmark.svelte";
  import CircleFilled from "../../components/icons/CircleFilled.svelte";
  import CircleHollow from "../../components/icons/CircleHollow.svelte";
  import Neutron from "../../components/icons/Neutron.svelte";
  import Arrow from "../../components/icons/Arrow.svelte";
  import Chevron from "../../components/icons/Chevron.svelte";
  import Target from "../../components/icons/Target.svelte";
  import Route from "../../components/icons/Route.svelte";
  import { ClipboardSetText } from "../../../wailsjs/runtime/runtime";
  import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
  import { ActiveJump } from "../../lib/routes/active";

  export let jumps: ActiveJump[];
  export let currentIndex: number = 0;

  let copiedSystemId: number | null = null;
  let targetedSystemId: number | null = null;

  async function copySystemName(systemName: string, systemId: number) {
    try {
      await ClipboardSetText(systemName);
      copiedSystemId = systemId;
      setTimeout(() => {
        copiedSystemId = null;
      }, 2000);
    } catch (err) {
      console.error("Failed to copy system name:", err);
    }
  }

  onMount(() => {
    EventsOn("Target", (targetData: any) => {
      targetedSystemId = targetData.SystemAddress;
    });
  });

  onDestroy(() => {
    EventsOff("Target");
  });
</script>

<Table
  columns={[
    { name: "", align: "center" },
    { name: "#", align: "left" },
    { name: "System", align: "left" },
    { name: "Scoopable", align: "center" },
    { name: "Neutron", align: "center" },
    { name: "Distance (LY)", align: "right" },
    { name: "Fuel", align: "right" },
  ]}
  data={jumps}
  let:item
  let:index
>
  <tr class:current={index === currentIndex}>
    <td class="align-center status-indicator">
      {#if index === currentIndex}
        <Chevron direction="right" size="1rem" color="var(--ed-orange)" />
      {:else if index === currentIndex + 1}
        <Target
          color={item.system_id === targetedSystemId
            ? "var(--ed-orange)"
            : "var(--ed-text-dim)"}
        />
      {:else if item.on_route}
        <Route color="var(--ed-text-dim)" />
      {/if}
    </td>
    <td class="align-left jump-index text-dim">{index + 1}</td>
    <td class="align-left">
      <div class="system-name-cell">
        <span>{item.system_name}</span>
        <button
          class="copy-btn text-dim"
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
        class="scoopable text-dim"
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
    <td class="align-center">
      {#if item.has_neutron}
        <Neutron color="var(--ed-orange)" />
      {/if}
    </td>
    <td class="align-right numeric">{item.distance.toFixed(2)}</td>
    <td class="align-right numeric fuel-cell">
      {#if item.fuel_in_tank !== undefined && item.fuel_used !== undefined}
        {item.fuel_in_tank.toFixed(2)}
        {#if index !== 0}
          <span class="fuel-used text-dim">
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
  </tr>
</Table>

<style>
  .status-indicator {
    width: 2rem;
    padding: 0.5rem 0.25rem;
    vertical-align: middle;
  }

  .status-indicator :global(svg) {
    vertical-align: middle;
  }

  .jump-index {
    font-weight: 600;
    font-size: 0.875rem;
  }

  .system-name-cell {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .copy-btn {
    background: none;
    border: none;
    padding: 0.25rem;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    border-radius: 2px;
    transition: all 0.2s;
    opacity: 0;
  }

  .system-name-cell:hover .copy-btn {
    opacity: 1;
  }

  .copy-btn:hover {
    background: var(--ed-bg-tertiary);
    color: var(--ed-text-secondary);
  }

  .copy-btn.copied {
    opacity: 1;
    color: var(--ed-success);
  }

  .scoopable {
    display: inline-flex;
    align-items: center;
  }

  .scoopable.must-refuel {
    color: var(--ed-orange);
  }

  .numeric {
    font-variant-numeric: tabular-nums;
    width: 1%;
    white-space: nowrap;
  }

  .fuel-cell {
    position: relative;
  }

  .fuel-used {
    position: absolute;
    bottom: 100%;
    right: 0%;
    transform: translate(-50%, 50%);
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    font-size: 0.75rem;
  }

  :global(tr.current) {
    background: hsl(from var(--ed-orange) h s l / 0.1);
  }

  :global(tr.current td:first-child) {
    padding-left: calc(1rem - 3px);
  }
</style>
