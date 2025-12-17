<script lang="ts">
  export let color: string = "currentColor";
  export let size: string = "0.875rem";
</script>

<svg
  viewBox="0 0 24 24"
  fill="none"
  style="width: {size}; height: {size}; vertical-align: middle; display: inline-block;"
  xmlns="http://www.w3.org/2000/svg"
>
  <path
    d="M18 6L6 18"
    stroke={color}
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  />
  <path
    d="M6 6L18 18"
    stroke={color}
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
  />
</svg>
