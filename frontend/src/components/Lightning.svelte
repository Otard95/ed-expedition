<script lang="ts">
  export let size: string = "1rem";
  export let color: string = "currentColor";
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
>
  <path
    d="M13 2L3 14h8l-1 8 10-12h-8l1-8z"
    fill={color}
    stroke={color}
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  />
</svg>

<style>
  svg {
    display: inline-block;
    vertical-align: middle;
  }
</style>
