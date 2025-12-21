<script lang="ts">
  export let variant: 'solid' | 'outline' | 'success' | 'info' | 'warning' = 'solid'
  export let active: boolean = false
  export let color: string | undefined = undefined
  export let onClick: ((e: MouseEvent) => void) | undefined = undefined

  const variantStyles: Record<typeof variant, { bg: string; border: string; text: string }> = {
    solid: { bg: color || 'var(--ed-orange)', border: 'transparent', text: 'var(--ed-text-on-orange)' },
    outline: { bg: 'transparent', border: color || 'var(--ed-orange)', text: color || 'var(--ed-orange)' },
    success: { bg: 'rgb(from var(--ed-success) r g b / 0.1)', border: 'var(--ed-success)', text: 'var(--ed-success)' },
    info: { bg: 'rgb(from var(--ed-info) r g b / 0.1)', border: 'var(--ed-info)', text: 'var(--ed-info)' },
    warning: { bg: 'rgb(from var(--ed-orange) r g b / 0.1)', border: 'var(--ed-orange)', text: 'var(--ed-orange)' }
  }

  $: styles = variantStyles[variant]
</script>

{#if onClick}
  <button
    class="badge"
    class:active
    class:clickable={true}
    style="background-color: {styles.bg}; border-color: {styles.border}; color: {styles.text};"
    on:click={onClick}
  >
    <slot />
  </button>
{:else}
  <span
    class="badge"
    class:active
    style="background-color: {styles.bg}; border-color: {styles.border}; color: {styles.text};"
  >
    <slot />
  </span>
{/if}

<style>
  .badge {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border: 1px solid transparent;
    border-radius: 2px;
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    white-space: nowrap;
  }

  .badge.active {
    box-shadow: 0 0 8px rgb(from var(--ed-orange) r g b / 0.5);
  }

  .badge.clickable {
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  .badge.clickable:hover {
    filter: brightness(1.15);
  }
</style>
