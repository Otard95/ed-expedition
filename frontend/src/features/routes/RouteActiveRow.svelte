<script lang="ts">
  import { onMount, onDestroy } from "svelte";
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
  import { models } from "../../../wailsjs/go/models";

  export let index: number = -1;
  export let jump: ActiveJump | null = null;
  export let isCurrent: boolean = false;
  export let isNext: boolean = false;
  export let isPrevOnRoute: boolean = true;

  let targetedSystemId: number | null = null;
  let confirmCopy: boolean = false;

  async function copySystemName() {
    try {
      await ClipboardSetText(jump.system_name);
      confirmCopy = true;
      setTimeout(() => {
        confirmCopy = false;
      }, 2000);
    } catch (err) {
      console.error("Failed to copy system name:", err);
    }
  }

  onMount(() => {
    EventsOn("Target", (targetData: any) => {
      targetedSystemId = targetData.SystemAddress;
    });
    if (isCurrent) {
      EventsOn("CurrentJump", (jumpHistEntry: models.JumpHistoryEntry) => {
        if (jump.system_id === jumpHistEntry.system_id) {
          jump.fuel_in_tank = jumpHistEntry.fuel_in_tank;
        }
      });
    }
  });

  onDestroy(() => {
    EventsOff("Target");
    EventsOff("CurrentJump");
  });

  function getActualColorClass(actual?: number, expected?: number): string {
    if (typeof actual !== "number" || typeof expected !== "number") return "";
    const diff = actual - expected;
    if (diff < -1.0) return "text-danger";
    if (diff < -0.5) return "text-warning";
    return "";
  }
</script>

<tr class:current={isCurrent}>
  <td class="align-center status-indicator">
    {#if isCurrent}
      <Chevron direction="right" size="1rem" color="var(--ed-orange)" />
    {:else if isNext}
      <Target
        color={jump.system_id === targetedSystemId
          ? "var(--ed-orange)"
          : "var(--ed-text-dim)"}
      />
    {:else if jump.on_route}
      <Route color="var(--ed-text-dim)" />
    {/if}
  </td>
  <td class="align-left jump-index text-dim">{index + 1}</td>
  <td class="align-left">
    <div class="system-name-cell">
      <span>{jump.system_name}</span>
      <button
        class="copy-btn text-dim"
        class:copied={confirmCopy}
        on:click={copySystemName}
        title="Copy system name"
      >
        {#if confirmCopy}
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
      class:must-refuel={jump.must_refuel}
      class:can-scoop={jump.scoopable}
    >
      {#if jump.scoopable}
        <CircleFilled size="1rem" />
      {:else}
        <CircleHollow size="1rem" />
      {/if}
    </span>
  </td>
  <td class="align-center">
    {#if jump.has_neutron}
      <Neutron color="var(--ed-orange)" />
    {/if}
  </td>
  <td class="align-right numeric">{jump.distance.toFixed(2)}</td>
  <td class="align-right numeric fuel-cell">
    <span class="flex-between flex-gap-sm">
      <span
        class="fuel-actual {getActualColorClass(
          jump.fuel_in_tank,
          jump.expected_fuel,
        )}"
        >{jump.fuel_in_tank !== undefined
          ? jump.fuel_in_tank.toFixed(2)
          : "-"}</span
      >
      <span class="text-dim">/</span>
      <span class="fuel-expected"
        >{jump.expected_fuel !== undefined
          ? jump.expected_fuel.toFixed(2)
          : "-"}</span
      >
    </span>
    {#if index !== 0}
      {#if isNext && !isPrevOnRoute}
        <span class="fuel-used text-dim">
          <Arrow
            direction="down"
            size="0.7rem"
            color="hsl(from var(--ed-danger) h calc(s * 0.3) calc(l * 0.7))"
          />
          ???
        </span>
      {:else if jump.fuel_used !== undefined}
        <span class="fuel-used text-dim">
          <Arrow
            direction="down"
            size="0.7rem"
            color="hsl(from var(--ed-danger) h calc(s * 0.3) calc(l * 0.7))"
          />
          {jump.fuel_used.toFixed(2)}
        </span>
      {/if}
    {/if}
  </td>
</tr>

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

  .fuel-actual {
    text-align: right;
    min-width: 2.5rem;
  }

  .fuel-expected {
    text-align: left;
    min-width: 2.5rem;
  }

  .fuel-used {
    position: absolute;
    bottom: 100%;
    left: 50%;
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
