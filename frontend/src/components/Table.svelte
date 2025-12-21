<script lang="ts">
  type Column = {
    name: string;
    align?: "left" | "center" | "right";
  };

  export let columns: Column[] = [];
  export let data: any[] = [];
  export let compact: boolean = false;
</script>

<div class="table-container" class:compact>
  <table>
    <thead>
      <tr>
        {#each columns as column}
          <th class="text-uppercase-tracked align-{column.align || 'left'}">{column.name}</th>
        {/each}
      </tr>
    </thead>
    <tbody>
      {#each data as item, index}
        <slot {item} {index} />
      {/each}
    </tbody>
  </table>
</div>

<style>
  .table-container {
    overflow-x: auto;
  }

  table {
    width: 100%;
    border-collapse: collapse;
  }

  thead {
    background: var(--ed-bg-tertiary);
  }

  th {
    padding: 0.75rem 1rem;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--ed-orange);
    border-bottom: 2px solid var(--ed-border-accent);
    white-space: nowrap;
  }

  th.align-left {
    text-align: left;
  }

  th.align-center {
    text-align: center;
  }

  th.align-right {
    text-align: right;
  }

  :global(tbody tr) {
    border-bottom: 1px solid var(--ed-border);
    transition: background-color 0.15s ease;
  }

  :global(tbody tr:hover) {
    background: var(--ed-bg-tertiary);
  }

  :global(tbody tr:last-child) {
    border-bottom: none;
  }

  :global(tbody td) {
    padding: 0.75rem 1rem;
    color: var(--ed-text-primary);
  }

  :global(tbody td.align-left) {
    text-align: left;
  }

  :global(tbody td.align-center) {
    text-align: center;
  }

  :global(tbody td.align-right) {
    text-align: right;
  }

  /* Compact mode - reduced padding */
  .table-container.compact th {
    padding: 0.4rem 0.6rem;
  }

  .table-container.compact :global(tbody td) {
    padding: 0.4rem 0.6rem;
  }
</style>
