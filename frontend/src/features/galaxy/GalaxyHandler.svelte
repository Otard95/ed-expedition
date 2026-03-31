<script lang="ts">
  import { toasts } from "../../lib/stores/toast";
  import { galaxy, GalaxyStatus } from "../../lib/stores/galaxy";
  import GalaxyPromptModal from "./GalaxyPromptModal.svelte";

  const TOAST_ID = "galaxy-prompt";

  let modalOpen = false;
  let wasInProgress = false;

  $: if ($galaxy) {
    if ($galaxy === GalaxyStatus.PROMPT) {
      showPromptToast();
    } else if ($galaxy === GalaxyStatus.PROMPT_CONTINUE) {
      showContinueToast();
    } else if ($galaxy === GalaxyStatus.IN_PROGRESS) {
      wasInProgress = true;
      showInProgressToast();
    } else if ($galaxy === GalaxyStatus.READY && wasInProgress) {
      showReadyToast();
    }
  }

  function showReadyToast() {
    toasts.set(TOAST_ID, {
      title: "Galaxy Database",
      message: "Ready to use!",
      level: "success",
      dismissable: true,
    });
  }

  function showInProgressToast() {
    toasts.set(TOAST_ID, {
      title: "Galaxy Database",
      message: "Downloading and building database",
      level: "info",
      persistent: true,
      dismissable: true,
      animate: true,
    });
  }

  function showContinueToast() {
    toasts.set(TOAST_ID, {
      title: "Galaxy Database",
      message: "Setup was interrupted. Pick up where you left off?",
      level: "warning",
      persistent: true,
      dismissable: true,
      action: {
        cta: "Continue",
        callback: () => {
          toasts.dismiss(TOAST_ID);
          galaxy.continue();
        },
      },
    });
  }

  function showPromptToast() {
    toasts.set(TOAST_ID, {
      title: "Galaxy Database",
      message: "Enable built-in route plotting with local star data?",
      level: "info",
      persistent: true,
      action: {
        cta: "Learn more",
        callback: () => {
          toasts.dismiss(TOAST_ID);
          modalOpen = true;
        },
      },
    });
  }

  async function handleAccept() {
    try {
      await galaxy.accept();
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      toasts.set(TOAST_ID, {
        title: "Galaxy Database",
        message: `Failed to start setup: ${msg}`,
        level: "danger",
        dismissable: true,
      });
    }
  }

  async function handleDecline() {
    try {
      await galaxy.decline();
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      toasts.set(TOAST_ID, {
        title: "Galaxy Database",
        message: `Something went wrong: ${msg}`,
        level: "danger",
        dismissable: true,
      });
    }
  }
</script>

<GalaxyPromptModal
  bind:open={modalOpen}
  onAccept={handleAccept}
  onDecline={handleDecline}
  onDismiss={() => (modalOpen = false)}
/>
