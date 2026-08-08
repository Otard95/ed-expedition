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
    d="M4.21 16.5a9 9 0 1 0 7.79-13.5 9.75 9.75 0 0 0-6.74 2.74L3 8"
    stroke={color}
    stroke-width="2.5"
    stroke-linecap="round"
    stroke-linejoin="round"
  />
  <path
    d="M3 3v5h5"
    stroke={color}
    stroke-width="3.5"
    stroke-linecap="round"
    stroke-linejoin="round"
  />
</svg>
