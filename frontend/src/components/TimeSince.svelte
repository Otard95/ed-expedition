<script lang="ts">
  import { onDestroy } from "svelte";
  import { formatDuration } from "../lib/utils/dateFormat";

  export let since: string | Date;
  export let interval: number | undefined = undefined;

  let duration: string;
  let intervalId: number | undefined;

  function updateDuration() {
    const startTime = since instanceof Date ? since : new Date(since);
    const durationMs = Date.now() - startTime.getTime();
    duration = formatDuration(durationMs);
  }

  $: {
    since;
    updateDuration();

    if (intervalId !== undefined) {
      clearInterval(intervalId);
      intervalId = undefined;
    }

    if (interval !== undefined) {
      intervalId = window.setInterval(updateDuration, interval);
    }
  }

  onDestroy(() => {
    if (intervalId !== undefined) {
      clearInterval(intervalId);
    }
  });
</script>

{duration}
