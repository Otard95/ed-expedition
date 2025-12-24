<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { EventsOn } from "../../../wailsjs/runtime";
  import { toasts } from "../../lib/stores/toast";

  interface FuelAlert {
    level: number;
    message: string;
  }

  const TOAST_ID = "fuel-alert";
  const levelToTitle = [undefined, "Fuel Warning", "Fuel Critical"];
  const levelToToastLevel = ["success", "warning", "danger"] as const;

  let prevLevel: number = -1;
  let cleanupFuel: (() => void) | null = null;
  let cleanupComplete: (() => void) | null = null;

  onMount(() => {
    cleanupFuel = EventsOn("FuelAlert", (alert: FuelAlert) => {
      if (alert.message.length === 0) {
        toasts.dismiss(TOAST_ID);
        return;
      }

      if (!(prevLevel === 0 && alert.level === 0)) {
        toasts.set(TOAST_ID, {
          title: levelToTitle[alert.level],
          message: alert.message,
          level: levelToToastLevel[alert.level],
          persistent: alert.level >= 1,
          dismissable: alert.level === 0,
          animate: alert.level >= 1,
        });
      }

      prevLevel = alert.level;
    });

    cleanupComplete = EventsOn("CompleteExpedition", () => {
      toasts.dismiss(TOAST_ID);
    });
  });

  onDestroy(() => {
    cleanupFuel?.();
    cleanupComplete?.();
  });
</script>
