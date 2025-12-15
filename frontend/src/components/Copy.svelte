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
  <rect
    x="9"
    y="9"
    width="13"
    height="13"
    rx="2"
    ry="2"
    stroke={color}
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    fill="none"
  />
  <path
    d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"
    stroke={color}
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    fill="none"
  />
</svg>
