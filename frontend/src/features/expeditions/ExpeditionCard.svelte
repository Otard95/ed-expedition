<script lang="ts">
  import type { models } from "../../../wailsjs/go/models";
  import Card from "../../components/Card.svelte";
  import ExpeditionStatusBadge from "../../components/ExpeditionStatusBadge.svelte";
  import Button from "../../components/Button.svelte";
  import Dropdown from "../../components/Dropdown.svelte";
  import DropdownItem from "../../components/DropdownItem.svelte";
  import { formatRelativeTime } from "../../lib/utils/dateFormat";

  export let expedition: models.ExpeditionSummary;

  $: isActive = expedition.status === "active";

  function handleView() {
    console.log("View expedition:", expedition.id);
  }

  function handleClone() {
    console.log("Clone expedition:", expedition.id);
  }

  function handleDelete() {
    console.log("Delete expedition:", expedition.id);
  }

  function handleStart() {
    console.log("Start expedition:", expedition.id);
  }

  function handleEnd() {
    console.log("End expedition:", expedition.id);
  }
</script>

<Card variant={isActive ? "active" : "default"} padding="1.5rem">
  <div class="expedition-card">
    <div class="content">
      <h3 class="name">{expedition.name}</h3>
      <div class="meta">
        <ExpeditionStatusBadge status={expedition.status} />
        <span class="dates">
          Created {formatRelativeTime(expedition.created_at)}
          {#if expedition.last_updated !== expedition.created_at}
            · Updated {formatRelativeTime(expedition.last_updated)}
          {/if}
        </span>
      </div>
    </div>
    <div class="actions">
      <Button variant="primary" size="small" onClick={handleView}>View</Button>
      <Dropdown>
        {#if expedition.status === "planned"}
          <DropdownItem onClick={handleStart}>Start</DropdownItem>
        {/if}
        {#if expedition.status === "active"}
          <DropdownItem onClick={handleEnd}>End</DropdownItem>
        {/if}
        <DropdownItem onClick={handleClone}>Clone</DropdownItem>
        <DropdownItem variant="danger" onClick={handleDelete}
          >Delete</DropdownItem
        >
      </Dropdown>
    </div>
  </div>
</Card>

<style>
  .expedition-card {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 2rem;
  }

  .content {
    flex: 1;
    min-width: 0;
  }

  .name {
    margin: 0 0 0.75rem 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--ed-text-primary);
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .dates {
    font-size: 0.875rem;
    color: var(--ed-text-secondary);
  }

  .actions {
    display: flex;
    gap: 0.5rem;
    flex-shrink: 0;
  }
</style>
