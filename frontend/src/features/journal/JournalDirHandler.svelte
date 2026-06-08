<script lang="ts">
  import { onMount } from "svelte";
  import { toasts } from "../../lib/stores/toast";
  import {
    GetJournalDirStatus,
    SetJournalDir,
  } from "../../../wailsjs/go/main/App";
  import JournalDirModal from "./JournalDirModal.svelte";

  const TOAST_ID = "journal-dir";

  let modalOpen = false;

  onMount(async () => {
    const configured = await GetJournalDirStatus();
    if (!configured) {
      modalOpen = true;
    }
  });

  async function handleConfirm(path: string) {
    try {
      await SetJournalDir(path);
      modalOpen = false;
      toasts.set(TOAST_ID, {
        title: "Journal Directory",
        message: "Journal directory configured successfully.",
        level: "success",
        dismissable: true,
      });
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      toasts.set(TOAST_ID, {
        title: "Journal Directory",
        message: `Failed to set journal directory: ${msg}`,
        level: "danger",
        dismissable: true,
      });
    }
  }
</script>

<JournalDirModal
  bind:open={modalOpen}
  onConfirm={handleConfirm}
/>
