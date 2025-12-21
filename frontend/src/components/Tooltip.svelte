<script lang="ts">
  export let text: string;
  export let direction: 'up' | 'down' | 'left' | 'right' = 'up';
  export let nowrap: boolean = false;
  export let size: string = '1rem';

  let showTooltip = false;
</script>

<span
  class="tooltip-trigger flex-center"
  on:mouseenter={() => (showTooltip = true)}
  on:mouseleave={() => (showTooltip = false)}
>
  <span class="tooltip-icon flex-center" style="width: {size}; height: {size}; font-size: calc({size} * 0.75);">?</span>
  {#if showTooltip}
    <span class="tooltip-content {direction}" class:nowrap>{text}</span>
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
    border-radius: 50%;
    background: var(--ed-bg-tertiary);
    border: 1px solid var(--ed-border);
    color: var(--ed-text-secondary);
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
    pointer-events: none;
    white-space: normal;
  }

  .tooltip-content.nowrap {
    white-space: nowrap;
    min-width: auto;
    max-width: none;
  }

  /* Direction: up (default) */
  .tooltip-content.up {
    bottom: calc(100% + 0.5rem);
    left: 50%;
    transform: translateX(-50%);
  }

  .tooltip-content.up::after {
    content: "";
    position: absolute;
    top: 100%;
    left: 50%;
    transform: translateX(-50%);
    border: 6px solid transparent;
    border-top-color: var(--ed-orange);
  }

  /* Direction: down */
  .tooltip-content.down {
    top: calc(100% + 0.5rem);
    left: 50%;
    transform: translateX(-50%);
  }

  .tooltip-content.down::after {
    content: "";
    position: absolute;
    bottom: 100%;
    left: 50%;
    transform: translateX(-50%);
    border: 6px solid transparent;
    border-bottom-color: var(--ed-orange);
  }

  /* Direction: left */
  .tooltip-content.left {
    right: calc(100% + 0.5rem);
    top: 50%;
    transform: translateY(-50%);
  }

  .tooltip-content.left::after {
    content: "";
    position: absolute;
    left: 100%;
    top: 50%;
    transform: translateY(-50%);
    border: 6px solid transparent;
    border-left-color: var(--ed-orange);
  }

  /* Direction: right */
  .tooltip-content.right {
    left: calc(100% + 0.5rem);
    top: 50%;
    transform: translateY(-50%);
  }

  .tooltip-content.right::after {
    content: "";
    position: absolute;
    right: 100%;
    top: 50%;
    transform: translateY(-50%);
    border: 6px solid transparent;
    border-right-color: var(--ed-orange);
  }
</style>
