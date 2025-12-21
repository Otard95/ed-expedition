<script lang="ts">
  import type { models } from '../../../wailsjs/go/models'
  import ExpeditionCard from './ExpeditionCard.svelte'

  export let expeditions: models.ExpeditionSummary[]
  export let onExpeditionDeleted: ((id: string) => void) | undefined = undefined
</script>

<div class="expedition-list">
  {#if expeditions.length === 0}
    <div class="empty-state">
      <p class="empty-message text-secondary">No expeditions yet. Click "New Expedition" above to get started.</p>
    </div>
  {:else}
    <div class="list stack-md">
      {#each expeditions as expedition (expedition.id)}
        <ExpeditionCard {expedition} onDelete={onExpeditionDeleted} />
      {/each}
    </div>
  {/if}
</div>

<style>
  .expedition-list {
    width: 100%;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1.5rem;
    padding: 4rem 2rem;
    text-align: center;
  }

  .empty-message {
    margin: 0;
    font-size: 1.125rem;
  }

</style>
