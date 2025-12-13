<script lang="ts">
  import type { models } from '../../../wailsjs/go/models'
  import ExpeditionCard from './ExpeditionCard.svelte'
  import Button from '../../components/Button.svelte'

  export let expeditions: models.ExpeditionSummary[]

  function handleCreateExpedition() {
    console.log('Create new expedition')
  }
</script>

<div class="expedition-list">
  {#if expeditions.length === 0}
    <div class="empty-state">
      <p class="empty-message">Nothing here yet. Would you like to create your first expedition?</p>
      <Button variant="primary" size="medium" onClick={handleCreateExpedition}>
        Create Expedition
      </Button>
    </div>
  {:else}
    <div class="list">
      {#each expeditions as expedition (expedition.id)}
        <ExpeditionCard {expedition} />
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
    color: #A0A0A0;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
</style>
