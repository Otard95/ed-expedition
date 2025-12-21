<script lang="ts">
  export let variant: "primary" | "secondary" | "danger" = "primary";
  export let size: "small" | "medium" | "large" = "medium";
  export let disabled: boolean = false;
  export let onClick: (() => void) | undefined = undefined;
  let className: string = "";
  export { className as class };

  function handleClick() {
    if (!disabled && onClick) {
      onClick();
    }
  }
</script>

<button
  class="btn {variant} {size} {className}"
  {disabled}
  on:click={handleClick}
>
  <slot />
</button>

<style>
  .btn {
    border: none;
    border-radius: 2px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    cursor: pointer;
    transition:
      background-color 0.15s ease,
      border-color 0.15s ease;
  }

  .btn:disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }

  /* Sizes */
  .btn.small {
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
  }

  .btn.medium {
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
  }

  .btn.large {
    padding: 0.75rem 1.5rem;
    font-size: 1rem;
  }

  /* Primary variant */
  .btn.primary {
    background: var(--ed-orange);
    color: var(--ed-text-on-orange);
  }

  .btn.primary:hover:not(:disabled) {
    background: var(--ed-orange-bright);
  }

  .btn.primary:disabled {
    background: var(--ed-orange-dim);
  }

  /* Secondary variant */
  .btn.secondary {
    background: transparent;
    color: var(--ed-orange);
    border: 1px solid var(--ed-orange);
  }

  .btn.secondary:hover:not(:disabled) {
    background: rgb(from var(--ed-orange) r g b / 0.1);
    border-color: var(--ed-orange-bright);
    color: var(--ed-orange-bright);
  }

  .btn.secondary:disabled {
    border-color: var(--ed-orange-dim);
    color: var(--ed-orange-dim);
  }

  /* Danger variant */
  .btn.danger {
    background: var(--ed-danger);
    color: var(--ed-text-on-danger);
  }

  .btn.danger:hover:not(:disabled) {
    background: var(--ed-danger-hover);
  }

  .btn.danger:disabled {
    background: var(--ed-danger-disabled);
  }
</style>
