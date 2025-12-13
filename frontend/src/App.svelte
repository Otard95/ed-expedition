<script lang="ts">
  import { onMount } from 'svelte'
  import ExpeditionList from './features/expeditions/ExpeditionList.svelte'
  import { GetExpeditionSummaries } from '../wailsjs/go/main/App'
  import type { models } from '../wailsjs/go/models'

  let expeditions: models.ExpeditionSummary[] = []
  let loading = true
  let error: string | null = null

  onMount(async () => {
    try {
      expeditions = await GetExpeditionSummaries()
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load expeditions'
      console.error('Failed to load expeditions:', e)
    } finally {
      loading = false
    }
  })
</script>

<main>
  <div class="container">
    <h1>ED Expedition Manager</h1>

    {#if loading}
      <p class="loading">Loading expeditions...</p>
    {:else if error}
      <p class="error">Error: {error}</p>
    {:else}
      <ExpeditionList expeditions={expeditions} />
    {/if}
  </div>
</main>

<style>
  main {
    padding: 2rem;
    max-width: 1200px;
    margin: 0 auto;
  }

  .container {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  h1 {
    margin: 0;
    font-size: 2rem;
    font-weight: 600;
    color: var(--ed-orange);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .loading,
  .error {
    text-align: center;
    padding: 2rem;
    font-size: 1.125rem;
  }

  .loading {
    color: var(--ed-text-secondary);
  }

  .error {
    color: var(--ed-danger);
  }
</style>
