<script lang="ts">
  export let text: string;

  let showTooltip = false;
</script>

<span
  class="tooltip-trigger flex-center"
  on:mouseenter={() => (showTooltip = true)}
  on:mouseleave={() => (showTooltip = false)}
>
  <span class="tooltip-icon flex-center">?</span>
  {#if showTooltip}
    <span class="tooltip-content">{text}</span>
  {/if}
</span>

<style>
  .tooltip-trigger {
    position: relative;
    display: inline-flex;
    margin-left: 0.375rem;
    cursor: help;
  }

  .tooltip-icon {
    width: 1rem;
    height: 1rem;
    border-radius: 50%;
    background: var(--ed-bg-tertiary);
    border: 1px solid var(--ed-border);
    color: var(--ed-text-secondary);
    font-size: 0.75rem;
    font-weight: 600;
    transition: all 0.15s ease;
  }

  .tooltip-trigger:hover .tooltip-icon {
    background: var(--ed-orange-dim);
    border-color: var(--ed-orange);
    color: var(--ed-text-primary);
  }

  .tooltip-content {
    position: absolute;
    bottom: calc(100% + 0.5rem);
    left: 50%;
    transform: translateX(-50%);
    min-width: 200px;
    max-width: 300px;
    padding: 0.75rem;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-orange);
    border-radius: 2px;
    color: var(--ed-text-primary);
    font-size: 0.875rem;
    line-height: 1.4;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    z-index: 1000;
    /* Prevent tooltip from capturing mouse events, which would cause flickering
       when the mouse moves over the tooltip itself (triggering mouseleave on the icon) */
    pointer-events: none;
    white-space: normal;
  }

  /* CSS triangle pointing downward, created using border trick.
     Border-top creates the visible triangle, transparent sides create the point shape. */
  .tooltip-content::after {
    content: "";
    position: absolute;
    top: 100%;
    left: 50%;
    transform: translateX(-50%);
    border: 6px solid transparent;
    border-top-color: var(--ed-orange);
  }
</style>
