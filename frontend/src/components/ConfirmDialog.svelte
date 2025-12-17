<script lang="ts">
  import Modal from "./Modal.svelte";
  import Button from "./Button.svelte";

  export let open: boolean = false;
  export let title: string;
  export let message: string;
  export let warningMessage: string | null = null;
  export let confirmLabel: string = "Confirm";
  export let cancelLabel: string = "Cancel";
  export let confirmVariant: "primary" | "danger" = "primary";
  export let loading: boolean = false;
  export let onConfirm: () => void;
  export let onCancel: () => void;
</script>

<Modal
  bind:open
  {title}
  onRequestClose={onCancel}
>
  <div class="confirm-dialog stack-md">
    <p class="message">{@html message}</p>
    {#if warningMessage}
      <p class="warning">{warningMessage}</p>
    {/if}

    <div class="modal-actions">
      <Button variant="secondary" onClick={onCancel} disabled={loading}>
        {cancelLabel}
      </Button>
      <Button variant={confirmVariant} onClick={onConfirm} disabled={loading}>
        {loading ? 'Loading...' : confirmLabel}
      </Button>
    </div>
  </div>
</Modal>

<style>
  .message {
    margin: 0;
    color: var(--ed-text-primary);
  }

  .message :global(strong) {
    color: var(--ed-orange);
  }

  .warning {
    margin: 0;
    color: var(--ed-danger);
    font-weight: 500;
  }

  .modal-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
    margin-top: 0.5rem;
  }
</style>
