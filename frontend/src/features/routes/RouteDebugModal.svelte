<script lang="ts">
  import Modal from "../../components/Modal.svelte";
  import type { EditViewRoute } from "../../lib/routes/edit";

  export let open: boolean = false;
  export let route: EditViewRoute;
</script>

<Modal bind:open title="Route Debug Info" onRequestClose={() => open = false}>
  <div class="debug-content">
    <section>
      <h3>Route</h3>
      <dl>
        <dt>ID</dt>
        <dd class="mono">{route.id}</dd>
        <dt>Plotter</dt>
        <dd>{route.plotter}</dd>
        <dt>Jumps</dt>
        <dd>{route.jumps.length}</dd>
        <dt>Created</dt>
        <dd>{route.created_at}</dd>
      </dl>
    </section>

    <section>
      <h3>Plotter Parameters</h3>
      <pre class="mono">{JSON.stringify(route.plotter_parameters, null, 2)}</pre>
    </section>

    {#if route.plotter_metadata}
      <section>
        <h3>Plotter Metadata</h3>
        <pre class="mono">{JSON.stringify(route.plotter_metadata, null, 2)}</pre>
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
</style>
