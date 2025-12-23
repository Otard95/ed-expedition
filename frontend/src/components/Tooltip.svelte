<script lang="ts">
  export let text: string;
  export let direction: 'up' | 'down' | 'left' | 'right' = 'up';
  export let nowrap: boolean = false;
  export let size: string = '1rem';

  let showTooltip = false;
  let triggerElement: HTMLSpanElement;
  const GAP = 8;
  let tooltipPosition = { top: "0px", right: "auto", bottom: "auto", left: "0px" };

  function updatePosition() {
    if (triggerElement) {
      const rect = triggerElement.getBoundingClientRect();

      let top = "auto";
      let right = "auto";
      let bottom = "auto";
      let left = "auto";

      if (direction === 'up') {
        bottom = window.innerHeight - rect.top + GAP + "px";
        left = rect.left + rect.width / 2 + "px";
      } else if (direction === 'down') {
        top = rect.bottom + GAP + "px";
        left = rect.left + rect.width / 2 + "px";
      } else if (direction === 'left') {
        right = window.innerWidth - rect.left + GAP + "px";
        top = rect.top + rect.height / 2 + "px";
      } else if (direction === 'right') {
        left = rect.right + GAP + "px";
        top = rect.top + rect.height / 2 + "px";
      }

      tooltipPosition = { top, right, bottom, left };
    }
  }

  function handleMouseEnter() {
    showTooltip = true;
    updatePosition();
  }

  function handleMouseLeave() {
    showTooltip = false;
  }

  function handleScroll() {
    showTooltip = false;
  }
</script>

<svelte:window on:scroll={handleScroll} />

<span
  class="tooltip-trigger flex-center"
  bind:this={triggerElement}
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
>
  <span class="tooltip-icon flex-center" style="width: {size}; height: {size}; font-size: calc({size} * 0.75);">?</span>
  {#if showTooltip}
    <span
      class="tooltip-content {direction}"
      class:nowrap
      style="top: {tooltipPosition.top}; right: {tooltipPosition.right}; bottom: {tooltipPosition.bottom}; left: {tooltipPosition.left};"
    >
      {text}
    </span>
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
    position: fixed;
    min-width: 200px;
    max-width: 300px;
    padding: 0.75rem;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-orange);
    border-radius: 2px;
    color: var(--ed-text-primary);
    font-size: 0.875rem;
    line-height: 1.4;
    box-shadow: var(--ed-shadow-md);
    z-index: 1000;
    pointer-events: none;
    white-space: normal;
  }

  .tooltip-content.nowrap {
    white-space: nowrap;
    min-width: auto;
    max-width: none;
  }

  /* Direction: up */
  .tooltip-content.up {
    transform: translate(-50%, 0);
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
    transform: translate(-50%, 0);
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
    transform: translate(0, -50%);
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
    transform: translate(0, -50%);
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
