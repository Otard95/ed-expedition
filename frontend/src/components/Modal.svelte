<script lang="ts">
  import { onMount, onDestroy } from "svelte";

  export let open: boolean = false;
  export let title: string = "";
  export let onRequestClose: (() => void) | undefined = undefined;
  export let showCloseButton: boolean = true;

  let className: string = "";
  export { className as class };

  function handleBackdropClick(event: MouseEvent) {
    if (onRequestClose && event.target === event.currentTarget) {
      onRequestClose();
    }
  }

  function handleEscapeKey(event: KeyboardEvent) {
    if (onRequestClose && event.key === "Escape") {
      onRequestClose();
    }
  }

  $: if (open && onRequestClose) {
    window.addEventListener("keydown", handleEscapeKey);
  } else {
    window.removeEventListener("keydown", handleEscapeKey);
  }

  onDestroy(() => {
    window.removeEventListener("keydown", handleEscapeKey);
  });
</script>

{#if open}
  <div class="modal-backdrop" on:click={handleBackdropClick}>
    <div class="modal-content {className}" on:click|stopPropagation>
      {#if title || (showCloseButton && onRequestClose)}
        <div class="modal-header">
          {#if title}
            <h2 class="modal-title">{title}</h2>
          {/if}
          {#if showCloseButton && onRequestClose}
            <button class="close-button" on:click={onRequestClose}>×</button>
          {/if}
        </div>
      {/if}
      <div class="modal-body">
        <slot />
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 2rem;
  }

  .modal-content {
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border);
    border-radius: 4px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid var(--ed-border);
  }

  .modal-title {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--ed-orange);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .close-button {
    background: none;
    border: none;
    color: var(--ed-text-secondary);
    font-size: 2rem;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 2px;
    transition: all 0.15s ease;
  }

  .close-button:hover {
    background: var(--ed-bg-tertiary);
    color: var(--ed-orange);
  }

  .modal-body {
    padding: 1.5rem;
  }
</style>
