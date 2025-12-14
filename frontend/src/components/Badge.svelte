<script lang="ts">
  export let variant: 'solid' | 'outline' | 'success' | 'info' | 'warning' = 'solid'
  export let active: boolean = false
  export let color: string | undefined = undefined
  export let onClick: ((e: MouseEvent) => void) | undefined = undefined

  const variantStyles: Record<typeof variant, { bg: string; border: string; text: string }> = {
    solid: { bg: color || 'var(--ed-orange)', border: 'transparent', text: '#000' },
    outline: { bg: 'transparent', border: color || 'var(--ed-orange)', text: color || 'var(--ed-orange)' },
    success: { bg: 'rgba(34, 197, 94, 0.1)', border: 'rgb(34, 197, 94)', text: 'rgb(34, 197, 94)' },
    info: { bg: 'rgba(59, 130, 246, 0.1)', border: 'var(--ed-info)', text: 'var(--ed-info)' },
    warning: { bg: 'rgba(255, 120, 0, 0.1)', border: 'var(--ed-orange)', text: 'var(--ed-orange)' }
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
    box-shadow: 0 0 8px rgba(255, 120, 0, 0.5);
  }

  .badge.clickable {
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  .badge.clickable:hover {
    filter: brightness(1.15);
  }
</style>
