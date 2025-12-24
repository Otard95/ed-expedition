<script lang="ts">
  import { onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import ExpeditionList from "../features/expeditions/ExpeditionList.svelte";
  import Button from "../components/Button.svelte";
  import {
    GetExpeditionSummaries,
    CreateExpedition,
  } from "../../wailsjs/go/main/App";
  import type { models } from "../../wailsjs/go/models";
  import { toasts } from "../lib/stores/toast";

  let expeditions: models.ExpeditionSummary[] = [];
  let loading = true;
  let error: string | null = null;
  let creating = false;

  let clickTimes: number[] = [];
  function handleTitleClick() {
    const now = Date.now();
    clickTimes = [...clickTimes.filter(t => now - t < 500), now];
    if (clickTimes.length >= 3) {
      clickTimes = [];
      toasts.set("dev-tools", {
        message: "Dev tools unlocked",
        level: "info",
        persistent: false,
        dismissable: true,
        action: {
          cta: "Toast Test",
          callback: () => {
            push("/test/toasts");
            toasts.dismiss("dev-tools");
          },
        },
      });
    }
  }

  onMount(async () => {
    await loadExpeditions();
  });

  async function loadExpeditions() {
    try {
      loading = true;
      expeditions = await GetExpeditionSummaries();
    } catch (e) {
      error = e instanceof Error ? e.message : "Failed to load expeditions";
      console.error("Failed to load expeditions:", e);
    } finally {
      loading = false;
    }
  }

  async function handleCreateExpedition() {
    if (creating) return;
    creating = true;
    try {
      const expeditionId = await CreateExpedition();
      push(`/expeditions/${expeditionId}`);
    } catch (err) {
      console.error("Failed to create expedition:", err);
      alert("Failed to create expedition");
    } finally {
      creating = false;
    }
  }

  async function handleExpeditionDeleted(id: string) {
    await loadExpeditions();
  }
</script>

<div class="expedition-index flex-col flex-gap-lg">
  <div class="header flex-between">
    <h1 class="text-uppercase-tracked" on:click={handleTitleClick}>ED Expedition Manager</h1>
    <Button
      variant="primary"
      onClick={handleCreateExpedition}
      disabled={creating}
    >
      {creating ? "Creating..." : "New Expedition"}
    </Button>
  </div>

  {#if loading}
    <p class="loading text-secondary">Loading expeditions...</p>
  {:else if error}
    <p class="error text-danger">Error: {error}</p>
  {:else}
    <ExpeditionList
      {expeditions}
      onExpeditionDeleted={handleExpeditionDeleted}
    />
  {/if}
</div>

<style>
  .header {
    gap: 1rem;
  }

  h1 {
    margin: 0;
    font-size: 2rem;
    font-weight: 600;
    color: var(--ed-orange);
  }

  .loading,
  .error {
    text-align: center;
    padding: 2rem;
    font-size: 1.125rem;
  }
</style>
