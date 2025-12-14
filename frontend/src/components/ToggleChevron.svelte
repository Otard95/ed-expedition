<script lang="ts">
  import Chevron from "./Chevron.svelte";

  export let collapsed: boolean = false;
  export let onClick: (e: MouseEvent) => void;

  $: direction = collapsed ? 'right' : 'down';
</script>

<button class="chevron-toggle" on:click={onClick}>
  <Chevron {direction} />
</button>

<style>
  .chevron-toggle {
    background: none;
    border: none;
    padding: 0.25rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--ed-text-secondary);
    transition: color 0.15s ease;
  }

  .chevron-toggle:hover {
    color: var(--ed-orange);
  }
</style>
