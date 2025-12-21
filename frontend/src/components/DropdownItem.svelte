<script lang="ts">
  import { getContext } from 'svelte'
  import type { Writable } from 'svelte/store'

  export let variant: 'default' | 'danger' = 'default'
  export let onClick: (() => void) | undefined = undefined

  const context = getContext<{ isOpen: Writable<boolean> }>('dropdown')
  const { isOpen } = context

  function handleClick() {
    if (onClick) {
      onClick()
    }
    isOpen.set(false)
  }
</script>

<button class="item {variant}" on:click={handleClick}>
  <slot />
</button>

<style>
  .item {
    width: 100%;
    background: transparent;
    border: none;
    color: var(--ed-text-primary);
    text-align: left;
    padding: 0.75rem 1rem;
    cursor: pointer;
    font-size: 0.875rem;
    transition: background-color 0.15s ease, color 0.15s ease;
    border-bottom: 1px solid var(--ed-border);
  }

  .item:last-child {
    border-bottom: none;
  }

  .item:hover {
    background: rgb(from var(--ed-orange) r g b / 0.1);
    color: var(--ed-orange);
  }

  .item.danger {
    color: var(--ed-danger);
  }

  .item.danger:hover {
    background: rgb(from var(--ed-danger) r g b / 0.1);
    color: var(--ed-danger);
    filter: brightness(1.2);
  }
</style>
