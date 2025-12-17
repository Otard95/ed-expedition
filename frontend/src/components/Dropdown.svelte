<script lang="ts">
  import { setContext } from 'svelte'
  import { writable } from 'svelte/store'

  const isOpen = writable(false)
  setContext('dropdown', { isOpen })

  let dropdownElement: HTMLDivElement

  function toggle() {
    isOpen.update(val => !val)
  }

  function close() {
    isOpen.set(false)
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement
    if (dropdownElement && !dropdownElement.contains(target)) {
      close()
    }
  }

  $: if ($isOpen) {
    setTimeout(() => {
      document.addEventListener('click', handleClickOutside as any)
    }, 0)
  } else {
    document.removeEventListener('click', handleClickOutside as any)
  }
</script>

<div class="dropdown" bind:this={dropdownElement}>
  <button class="toggle flex-center" on:click={toggle}>
    <span class="dots">
      <span class="dot"></span>
      <span class="dot"></span>
      <span class="dot"></span>
    </span>
  </button>
  {#if $isOpen}
    <div class="menu">
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
    transition: color 0.15s ease, border-color 0.15s ease;
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
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: 0.25rem;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border);
    border-radius: 4px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.6);
    min-width: 150px;
    z-index: 1000;
  }
</style>
