<script lang="ts">
  import Modal from "../../components/Modal.svelte";
  import Button from "../../components/Button.svelte";
  import TextInput from "../../components/TextInput.svelte";
  import { BrowseJournalDir } from "../../../wailsjs/go/main/App";

  export let open: boolean = false;
  export let onConfirm: (path: string) => void;

  let path: string = "";
  let error: string = "";

  async function handleBrowse() {
    try {
      const selected = await BrowseJournalDir();
      if (selected) {
        path = selected;
        error = "";
      }
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    }
  }

  function handleConfirm() {
    if (!path.trim()) {
      error = "Please select or enter a directory path";
      return;
    }
    error = "";
    onConfirm(path.trim());
  }

</script>

<Modal {open} title="Journal Directory" showCloseButton={false}>
  <div class="content">
    <p>
      ED Expedition needs to know where your
      <strong>Elite Dangerous journal files</strong> are located to track your
      jumps in real-time.
    </p>

    <div class="input-row">
      <div class="input-field">
        <TextInput
          bind:value={path}
          placeholder="/path/to/journal/directory"
          label="Journal directory"
        />
      </div>
      <Button variant="secondary" onClick={handleBrowse}>Browse</Button>
    </div>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    <div class="actions">
      <Button onClick={handleConfirm} disabled={!path.trim()}>Confirm</Button>
    </div>
  </div>
</Modal>

<style>
  .content {
    padding: 1.5rem;
    max-width: 480px;
  }

  .content p {
    margin: 0 0 1.25rem;
    color: var(--ed-text-primary);
    line-height: 1.5;
  }

  .content strong {
    color: var(--ed-orange);
  }

  .input-row {
    display: flex;
    gap: 0.75rem;
    align-items: flex-end;
    margin-bottom: 1.25rem;
  }

  .input-field {
    flex: 1;
  }

  .error {
    color: var(--ed-danger);
    font-size: 0.85rem;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
  }
</style>
