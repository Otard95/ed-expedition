<script lang="ts">
  export let size: string = "1.2rem";
  export let color: string = "currentColor";
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 32 32"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
>
  <!-- Top line -->
  <line x1="16" y1="2" x2="16" y2="12" stroke={color} stroke-width="3.2" />

  <!-- Bottom line -->
  <line x1="16" y1="20" x2="16" y2="30" stroke={color} stroke-width="3.2" />

  <!-- Left line -->
  <line x1="2" y1="16" x2="12" y2="16" stroke={color} stroke-width="3.2" />

  <!-- Right line -->
  <line x1="20" y1="16" x2="30" y2="16" stroke={color} stroke-width="3.2" />

  <!-- Center circle -->
  <circle
    cx="16"
    cy="16"
    r="10"
    stroke={color}
    stroke-width="3.2"
    fill="none"
  />
</svg>
