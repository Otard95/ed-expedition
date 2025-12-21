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
    box-shadow: var(--ed-shadow-sm);
  }

  .card.active {
    border-color: var(--ed-orange);
    box-shadow: 0 0 10px rgb(from var(--ed-orange) r g b / 0.3);
  }
</style>
