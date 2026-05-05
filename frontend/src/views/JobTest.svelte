<script lang="ts">
  import { onDestroy, tick } from "svelte";
  import Button from "../components/Button.svelte";
  import NumberInput from "../components/NumberInput.svelte";
  import { push } from "svelte-spa-router";
  import { MockJob } from "../../wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";
  import { toasts } from "../lib/stores/toast";

  let duration = 10;
  let jobId: string | null = null;
  let logLines: string[] = [];
  let logEl: HTMLDivElement;

  function ts(): string {
    const d = new Date();
    return d.toLocaleTimeString("en-GB", { hour12: false }) + "." + String(d.getMilliseconds()).padStart(3, "0");
  }

  function log(msg: string) {
    logLines = [...logLines, `${ts()}  ${msg}`];
    tick().then(() => {
      if (logEl) logEl.scrollTop = logEl.scrollHeight;
    });
  }

  function formatStatus(status: any): string {
    const parts: string[] = [`[${status.status}]`];

    if (status.phase) {
      parts.push(`${status.phase.label} (${status.phase.index + 1}/${status.phase.total})`);
    }

    if (status.progress) {
      const pct = (status.progress.fraction * 100).toFixed(1);
      parts.push(`${pct}%`);
      if (status.progress.label) {
        parts.push(`- ${status.progress.label}`);
      }
    }

    return parts.join("  ");
  }

  function toastFromStatus(status: any) {
    const isDone = status.status === "complete" || status.status === "error";
    const phase = status.phase ? `${status.phase.label}` : "";
    const pct = status.progress ? `${(status.progress.fraction * 100).toFixed(0)}%` : "";
    const message = [phase, pct].filter(Boolean).join(" · ") || status.status;

    toasts.set("mock-job", {
      title: "Mock Job",
      message,
      level: isDone ? (status.status === "complete" ? "success" : "danger") : "info",
      persistent: !isDone,
      dismissable: isDone,
      animate: !isDone,
      progress: status.progress ? {
        fraction: status.progress.fraction,
        phase: !isDone && status.phase ? { index: status.phase.index, total: status.phase.total } : undefined,
      } : undefined,
      ...(isDone ? { timeout: 5000 } : {}),
    });
  }

  async function startMockJob() {
    if (jobId) {
      EventsOff(`job:${jobId}`);
    }
    logLines = [];
    jobId = await MockJob(duration);
    log(`job started: ${jobId}`);
    EventsOn(`job:${jobId}`, (status: any) => {
      log(formatStatus(status));
      toastFromStatus(status);
    });
  }

  onDestroy(() => {
    if (jobId) EventsOff(`job:${jobId}`);
  });
</script>

<div class="job-test flex-col flex-gap-lg">
  <div class="header flex-between">
    <h1 class="text-uppercase-tracked">Job Testing</h1>
    <Button variant="ghost" onClick={() => push("/")}>Back</Button>
  </div>

  <div class="controls flex-row flex-gap-md flex-align-end">
    <NumberInput label="Duration (seconds)" bind:value={duration} min={1} max={300} />
    <Button variant="primary" onClick={startMockJob}>Start Mock Job</Button>
  </div>

  {#if jobId}
    <p class="text-secondary">Job: <code>{jobId}</code></p>
  {/if}

  <div class="log" bind:this={logEl}>
    {#each logLines as line}
      <div class="log-line">{line}</div>
    {/each}
  </div>
</div>

<style>
  h1 {
    margin: 0;
    font-size: 2rem;
    font-weight: 600;
    color: var(--ed-orange);
  }

  .log {
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    padding: 0.5rem;
    height: 15rem;
    overflow-y: auto;
    font-family: monospace;
    font-size: 0.75rem;
  }

  .log-line {
    color: var(--ed-text-secondary);
    white-space: nowrap;
    line-height: 1.5;
    text-align: left;
  }
</style>
