<script lang="ts">
  export let size: string = "12px";
  export let color: string = "currentColor";
  export let direction: 'up' | 'right' | 'down' | 'left' = 'down';

  const rotations: Record<typeof direction, number> = {
    up: 180,
    right: -90,
    down: 0,
    left: 90
  };

  $: rotation = rotations[direction];
</script>

<svg
  viewBox="0 0 12 12"
  class="chevron"
  style="width: {size}; height: {size}; transform: rotate({rotation}deg);"
>
  <path
    d="M 2 4 L 6 8 L 10 4"
    fill="none"
    stroke={color}
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  />
</svg>

<style>
  .chevron {
    transition: transform 0.2s ease;
  }
</style>
