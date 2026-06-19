<script lang="ts">
  import Modal from "../../components/Modal.svelte";
  import type { models } from "../../../wailsjs/go/models";

  export let open: boolean = false;
  export let expedition: models.Expedition;
</script>

<Modal bind:open title="Expedition Debug Info" onRequestClose={() => open = false}>
  <div class="debug-content">
    <section>
      <h3>Expedition</h3>
      <dl>
        <dt>ID</dt>
        <dd class="mono">{expedition.id}</dd>
        <dt>Status</dt>
        <dd>{expedition.status}</dd>
        <dt>Created</dt>
        <dd>{expedition.created_at}</dd>
        <dt>Updated</dt>
        <dd>{expedition.last_updated}</dd>
        {#if expedition.started_on}
          <dt>Started</dt>
          <dd>{expedition.started_on}</dd>
        {/if}
        {#if expedition.ended_on}
          <dt>Ended</dt>
          <dd>{expedition.ended_on}</dd>
        {/if}
      </dl>
    </section>

    {#if expedition.start}
      <section>
        <h3>Start Position</h3>
        <dl>
          <dt>route_id</dt>
          <dd class="mono">{expedition.start.route_id}</dd>
          <dt>jump_index</dt>
          <dd>{expedition.start.jump_index}</dd>
        </dl>
      </section>
    {/if}

    <section>
      <h3>Routes</h3>
      <div class="id-list">
        {#each expedition.routes as routeId}
          <code>{routeId}</code>
        {/each}
      </div>
    </section>

    {#if expedition.links.length > 0}
      <section>
        <h3>Links ({expedition.links.length})</h3>
        <pre class="mono">{JSON.stringify(expedition.links, null, 2)}</pre>
      </section>
    {/if}

    <section>
      <h3>Baked Route</h3>
      <dl>
        {#if expedition.baked_route_id}
          <dt>baked_route_id</dt>
          <dd class="mono">{expedition.baked_route_id}</dd>
        {/if}
        <dt>current_baked_index</dt>
        <dd>{expedition.current_baked_index}</dd>
        {#if expedition.baked_loop_back_index !== undefined}
          <dt>baked_loop_back_index</dt>
          <dd>{expedition.baked_loop_back_index}</dd>
        {/if}
      </dl>
    </section>

    {#if expedition.jump_history?.length > 0}
      <section>
        <h3>Jump History ({expedition.jump_history.length})</h3>
        <div class="history-list">
          {#each expedition.jump_history.slice(-20) as entry, i}
            <div class="history-entry" class:off-route={!entry.on_route}>
              <span class="history-index">{expedition.jump_history.length - 20 + i + 1}</span>
              <span class="history-system">{entry.system_name}</span>
              <span class="history-flags">
                {#if entry.on_route}✓{:else}✗{/if}
                {#if !entry.expected}<span class="unexpected">unexpected</span>{/if}
              </span>
            </div>
          {/each}
          {#if expedition.jump_history.length > 20}
            <p class="text-dim">...showing last 20 of {expedition.jump_history.length}</p>
          {/if}
        </div>
      </section>
    {/if}
  </div>
</Modal>

<style>
  .debug-content {
    padding: 1.5rem;
    min-width: 400px;
    max-width: 600px;
    max-height: 70vh;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
  }

  section {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  h3 {
    margin: 0;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--ed-orange);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  dl {
    margin: 0;
    display: grid;
    grid-template-columns: auto 1fr;
    gap: 0.25rem 1rem;
  }

  dt {
    color: var(--ed-text-secondary);
    font-size: 0.8125rem;
    text-align: right;
  }

  dd {
    margin: 0;
    color: var(--ed-text-primary);
    font-size: 0.8125rem;
    text-align: left;
  }

  pre {
    margin: 0;
    padding: 0.75rem;
    background: var(--ed-bg-primary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    font-size: 0.75rem;
    color: var(--ed-text-primary);
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-all;
    text-align: left;
  }

  .mono {
    font-family: monospace;
  }

  .id-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .id-list code {
    font-size: 0.75rem;
    color: var(--ed-text-primary);
    font-family: monospace;
  }

  .history-list {
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }

  .history-entry {
    display: flex;
    gap: 0.75rem;
    font-size: 0.75rem;
    font-family: monospace;
    align-items: baseline;
  }

  .history-index {
    color: var(--ed-text-dim);
    min-width: 2rem;
    text-align: right;
  }

  .history-system {
    color: var(--ed-text-primary);
    flex: 1;
  }

  .history-entry.off-route .history-system {
    color: var(--ed-danger);
  }

  .history-flags {
    color: var(--ed-text-secondary);
  }

  .unexpected {
    color: var(--ed-warning);
    margin-left: 0.25rem;
    font-size: 0.6875rem;
  }
</style>
