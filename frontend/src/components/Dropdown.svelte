<script lang="ts">
  import { setContext, onDestroy } from "svelte";
  import { writable } from "svelte/store";

  const isOpen = writable(false);
  setContext("dropdown", { isOpen });

  let toggleElement: HTMLButtonElement;
  const GAP = 4;
  let menuPosition = { top: "0px", right: "0px", bottom: "auto", left: "auto" };
  let centerMenu = false;
  let isHoveringToggle = false;
  let isHoveringMenu = false;
  let closeTimeout: number | null = null;

  function toggle() {
    isOpen.update((val) => !val);
  }

  function close() {
    isOpen.set(false);
  }

  function scheduleClose() {
    if (closeTimeout) clearTimeout(closeTimeout);
    closeTimeout = window.setTimeout(() => {
      if (!isHoveringToggle && !isHoveringMenu) {
        close();
      }
    }, 300);
  }

  function handleToggleMouseEnter() {
    isHoveringToggle = true;
    if (closeTimeout) clearTimeout(closeTimeout);
  }

  function handleToggleMouseLeave() {
    isHoveringToggle = false;
    scheduleClose();
  }

  function handleMenuMouseEnter() {
    isHoveringMenu = true;
    if (closeTimeout) clearTimeout(closeTimeout);
  }

  function handleMenuMouseLeave() {
    isHoveringMenu = false;
    scheduleClose();
  }

  function updatePosition() {
    if (toggleElement) {
      const rect = toggleElement.getBoundingClientRect();
      const center = {
        x: rect.x + rect.width / 2,
        y: rect.y + rect.height / 2,
      };

      let top = "auto";
      let right = "auto";
      let bottom = "auto";
      let left = "auto";
      if (center.x < window.innerWidth / 2) {
        left = rect.right + GAP + "px";
      } else {
        right = window.innerWidth - rect.left + GAP + "px";
      }

      centerMenu = false;
      if (center.y < window.innerHeight * 0.33) {
        top = rect.top + "px";
      } else if (center.y > window.innerHeight * 0.66) {
        bottom = window.innerHeight - rect.bottom + "px";
      } else {
        top = rect.top + rect.height / 2 + "px";
        centerMenu = true;
      }

      menuPosition = { top, right, bottom, left };
    }
  }

  $: if ($isOpen) {
    updatePosition();
  }

  onDestroy(() => {
    if (closeTimeout) clearTimeout(closeTimeout);
  });
</script>

<svelte:window on:scroll={close} />

<div class="dropdown">
  <button
    class="toggle flex-center"
    bind:this={toggleElement}
    on:click={toggle}
    on:mouseenter={handleToggleMouseEnter}
    on:mouseleave={handleToggleMouseLeave}
  >
    <span class="dots">
      <span class="dot"></span>
      <span class="dot"></span>
      <span class="dot"></span>
    </span>
  </button>
  {#if $isOpen}
    <div
      class="menu flex-col"
      class:center-vertical={centerMenu}
      style="top: {menuPosition.top}; right: {menuPosition.right}; bottom: {menuPosition.bottom}; left: {menuPosition.left};"
      on:mouseenter={handleMenuMouseEnter}
      on:mouseleave={handleMenuMouseLeave}
    >
      <slot />
    </div>
  {/if}
</div>

<style>
  .dropdown {
    position: relative;
  }

  .toggle {
    background: transparent;
    border: 1px solid var(--ed-border);
    border-radius: 4px;
    color: var(--ed-text-secondary);
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    font-size: 1.25rem;
    line-height: 1;
    height: 30px;
    transition:
      color 0.15s ease,
      border-color 0.15s ease;
  }

  .toggle:hover {
    color: var(--ed-orange);
    border-color: var(--ed-orange);
  }

  .dots {
    display: flex;
    align-items: center;
    gap: 3px;
  }

  .dot {
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background-color: currentColor;
  }

  .menu {
    position: fixed;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border);
    border-radius: 4px;
    box-shadow: var(--ed-shadow-md);
    min-width: 150px;
    z-index: 1000;
  }
  .menu.center-vertical {
    transform: translateY(-50%);
  }
</style>
