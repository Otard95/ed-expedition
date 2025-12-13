<script lang="ts">
  export let variant: 'default' | 'active' = 'default'
  export let padding: string = '1rem'
  let className: string = ''
  export { className as class }
</script>

<div class="card {variant} {className}" style="padding: {padding}">
  <slot />
</div>

<style>
  .card {
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border);
    border-radius: 4px;
    color: var(--ed-text-primary);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.4);
  }

  .card.active {
    border-color: var(--ed-orange);
    box-shadow: 0 0 10px rgba(255, 120, 0, 0.3);
  }
</style>
