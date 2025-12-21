<script lang="ts">
  import { push } from "svelte-spa-router";
  import type { models } from "../../../wailsjs/go/models";
  import {
    DeleteExpedition,
    StartExpedition,
  } from "../../../wailsjs/go/main/App";
  import Card from "../../components/Card.svelte";
  import ExpeditionStatusBadge from "../../components/ExpeditionStatusBadge.svelte";
  import Button from "../../components/Button.svelte";
  import Dropdown from "../../components/Dropdown.svelte";
  import DropdownItem from "../../components/DropdownItem.svelte";
  import ConfirmDialog from "../../components/ConfirmDialog.svelte";
  import { formatRelativeTime } from "../../lib/utils/dateFormat";

  export let expedition: models.ExpeditionSummary;
  export let onDelete: ((id: string) => void) | undefined = undefined;

  let showDeleteConfirm = false;
  let deleting = false;

  $: isActive = expedition.status === "active";
  $: expeditionName = expedition.name || "Unnamed Expedition";

  function handleClone() {
    console.log("Clone expedition:", expedition.id);
  }

  function handleDeleteClick() {
    showDeleteConfirm = true;
  }

  async function confirmDelete() {
    if (deleting) return;

    deleting = true;
    try {
      await DeleteExpedition(expedition.id);
      showDeleteConfirm = false;
      if (onDelete) {
        onDelete(expedition.id);
      }
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err);
      alert(`Failed to delete expedition: ${errorMsg}`);
      console.error("Failed to delete expedition:", err);
    } finally {
      deleting = false;
    }
  }

  async function handleStart() {
    try {
      await StartExpedition(expedition.id);
      // Reload page to refresh expedition list
      window.location.reload();
    } catch (err) {
      console.error("Failed to start expedition:", err);
      alert(
        `Failed to start expedition: ${err instanceof Error ? err.message : String(err)}`,
      );
    }
  }

  function handleEnd() {
    console.log("End expedition:", expedition.id);
  }
</script>

<Card variant={isActive ? "active" : "default"} padding="1.5rem">
  <div class="expedition-card flex-between">
    <div class="content">
      <h3 class="name">{expedition.name}</h3>
      <div class="meta">
        <ExpeditionStatusBadge status={expedition.status} />
        <span class="dates text-secondary">
          Created {formatRelativeTime(expedition.created_at)}
          {#if expedition.last_updated !== expedition.created_at}
            · Updated {formatRelativeTime(expedition.last_updated)}
          {/if}
        </span>
      </div>
    </div>
    <div class="actions">
      {#if expedition.status === "planned"}
        <Button
          variant="primary"
          size="small"
          onClick={() => push(`/expeditions/${expedition.id}`)}
        >
          Edit
        </Button>
      {:else if expedition.status === "active"}
        <Button variant="primary" size="small" onClick={() => push("/active")}>
          View
        </Button>
      {:else}
        <Button
          variant="primary"
          size="small"
          onClick={() => push(`/expeditions/${expedition.id}/view`)}
        >
          View
        </Button>
      {/if}
      <Dropdown>
        {#if expedition.status === "planned"}
          <DropdownItem onClick={handleStart}>Start</DropdownItem>
        {/if}
        {#if expedition.status === "active"}
          <DropdownItem onClick={handleEnd}>End</DropdownItem>
        {/if}
        <DropdownItem onClick={handleClone}>Clone</DropdownItem>
        <DropdownItem variant="danger" onClick={handleDeleteClick}
          >Delete</DropdownItem
        >
      </Dropdown>
    </div>
  </div>
</Card>

<ConfirmDialog
  bind:open={showDeleteConfirm}
  title="Delete Expedition"
  message="Are you sure you want to delete <strong>'{expeditionName}'</strong>?"
  warningMessage="This action cannot be undone."
  confirmLabel="Delete"
  confirmVariant="danger"
  loading={deleting}
  onConfirm={confirmDelete}
  onCancel={() => (showDeleteConfirm = false)}
/>

<style>
  .expedition-card {
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
  }

  .actions {
    display: flex;
    gap: 0.5rem;
    flex-shrink: 0;
  }
</style>
