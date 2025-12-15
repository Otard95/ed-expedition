<script lang="ts">
  export let color: string = "currentColor";
  export let size: string = "0.875rem";
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  style="vertical-align: middle; display: inline-block;"
  xmlns="http://www.w3.org/2000/svg"
>
  <polyline
    points="20 6 9 17 4 12"
    stroke={color}
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    fill="none"
  />
</svg>
