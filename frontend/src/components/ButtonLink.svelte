<script lang="ts">
  export let href: string
  export let variant: 'primary' | 'secondary' = 'primary'
  export let size: 'small' | 'medium' | 'large' = 'medium'
</script>

<a {href} class="btn {variant} {size}">
  <slot />
</a>

<style>
  .btn {
    display: inline-block;
    text-decoration: none;
    border: none;
    border-radius: 2px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    cursor: pointer;
    transition: background-color 0.15s ease, border-color 0.15s ease;
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

  .btn.primary:hover {
    background: var(--ed-orange-bright);
  }

  /* Secondary variant */
  .btn.secondary {
    background: transparent;
    color: var(--ed-orange);
    border: 1px solid var(--ed-orange);
  }

  .btn.secondary:hover {
    background: rgb(from var(--ed-orange) r g b / 0.1);
    border-color: var(--ed-orange-bright);
    color: var(--ed-orange-bright);
  }
</style>
