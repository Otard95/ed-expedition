<script lang="ts">
  import Card from "../../components/Card.svelte";
  import Button from "../../components/Button.svelte";
  import Arrow from "../../components/Arrow.svelte";
  import { EditViewLink } from "../../lib/routes/edit";

  export let links: EditViewLink[];
  export let onGotoJump: (
    route_id: string,
    jump_index: number,
    event: MouseEvent,
  ) => void;
</script>

<Card>
  <div class="section-header">
    <h2>Links</h2>
    <Button variant="primary" size="small">Add Link</Button>
  </div>
  <hr />
  {#if links.length === 0}
    <p class="empty-message">No links between routes yet.</p>
  {:else}
    <div class="links-list">
      {#each links as link}
        <div class="link-item">
          <div class="link-connection">
            <button
              class="link-endpoint from"
              on:click={(e) => onGotoJump(link.from.route_id, link.from.jump_index, e)}
            >
              <div class="link-route-label">
                Route {link.from.route_idx + 1} | {link.from.route_name}
              </div>
              <div class="link-jump-info">
                <span class="jump-idx">Jump {link.from.jump_index}</span>
                <span class="system-name">{link.from.system_name}</span>
              </div>
            </button>

            <div class="link-arrow">
              <Arrow direction="right" color="var(--ed-orange)" size="1.2rem" />
            </div>

            <button
              class="link-endpoint to"
              on:click={(e) => onGotoJump(link.to.route_id, link.to.jump_index, e)}
            >
              <div class="link-route-label">
                Route {link.to.route_idx + 1} | {link.to.route_name}
              </div>
              <div class="link-jump-info">
                <span class="jump-idx">Jump {link.to.jump_index}</span>
                <span class="system-name">{link.to.system_name}</span>
              </div>
            </button>
          </div>

          <Button variant="secondary" size="small">Remove</Button>
        </div>
      {/each}
    </div>
  {/if}
</Card>

<style>
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  h2 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--ed-orange);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  hr {
    opacity: 0.3;
  }

  .empty-message {
    color: var(--ed-text-secondary);
    font-style: italic;
    margin: 0;
  }

  .links-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .link-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  .link-connection {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex: 1;
  }

  .link-endpoint {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.5rem 0.75rem;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
    flex: 1;
  }

  .link-endpoint:hover {
    border-color: var(--ed-orange);
    background: var(--ed-bg-hover);
  }

  .link-route-label {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--ed-orange);
    text-transform: uppercase;
  }

  .link-jump-info {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
  }

  .link-jump-info .jump-idx {
    font-size: 0.75rem;
    color: var(--ed-text-dim);
    font-variant-numeric: tabular-nums;
  }

  .link-jump-info .system-name {
    font-size: 0.875rem;
    color: var(--ed-text-primary);
  }

  .link-arrow {
    display: flex;
    align-items: center;
  }
</style>
