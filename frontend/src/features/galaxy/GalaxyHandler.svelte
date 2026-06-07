<script lang="ts">
  import { toasts } from "../../lib/stores/toast";
  import { galaxy, GalaxyStatus } from "../../lib/stores/galaxy";
  import GalaxyPromptModal from "./GalaxyPromptModal.svelte";

  const TOAST_ID = "galaxy-prompt";

  let modalOpen = false;
  let wasInProgress = false;

  galaxy.onJobStatus((status) => {
    const isDone = status.status === "complete" || status.status === "error";
    const phase = status.phase ? status.phase.label : "";
    const pct = status.progress ? `${(status.progress.fraction * 100).toFixed(1)}%` : "";
    const message = [phase, pct].filter(Boolean).join(" · ") || status.status;

    toasts.set(TOAST_ID, {
      title: "Galaxy Database",
      message,
      level: isDone ? (status.status === "complete" ? "success" : "danger") : "info",
      persistent: !isDone,
      dismissable: isDone,
      animate: !isDone,
      progress: status.progress ? {
        fraction: status.progress.fraction,
        phase: !isDone && status.phase ? { index: status.phase.index, total: status.phase.total } : undefined,
      } : undefined,
    });

    if (status.status === "complete") {
      galaxy.markReady();
    }
  });

  $: if ($galaxy) {
    if ($galaxy === GalaxyStatus.PROMPT) {
      showPromptToast();
    } else if ($galaxy === GalaxyStatus.PROMPT_CONTINUE) {
      showContinueToast();
    } else if ($galaxy === GalaxyStatus.IN_PROGRESS && !galaxy.getJobId()) {
      showInProgressToast();
    } else if ($galaxy === GalaxyStatus.READY && wasInProgress) {
      wasInProgress = false;
    }
  }

  function showInProgressToast() {
    wasInProgress = true;
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
        callback: async () => {
          toasts.dismiss(TOAST_ID);
          await galaxy.continue();
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
