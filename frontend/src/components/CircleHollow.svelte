<script lang="ts">
  export let color: string = "currentColor";
  export let size: string = "1rem";
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 24 24"
  fill="none"
  style="vertical-align: middle; display: inline-block;"
  xmlns="http://www.w3.org/2000/svg"
>
  <circle
    cx="12"
    cy="12"
    r="10"
    stroke={color}
    stroke-width="2"
    fill="none"
  />
</svg>
