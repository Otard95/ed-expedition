<script lang="ts">
  export let size: string = "1.3rem";
  export let color: string = "currentColor";
  export let level: 1 | 2 | 3 = 1;

  const lineGap = 10;

  $: lineCount = level;
  const totalHeight = 36 + 4 + 3 * lineGap;
</script>

<div style="width: {size}; height: {size}; display: inline-flex;">
  <svg
    viewBox="0 0 40 {totalHeight}"
    width="100%"
    height="100%"
    xmlns="http://www.w3.org/2000/svg"
  >
    <polygon
      fill="none"
      stroke={color}
      stroke-width="3"
      points="20,2 38,33 2,33"
    />
    <circle fill={color} cx="20" cy="20" r="4" />
    {#each Array(lineCount) as _, i}
      <rect
        fill={color}
        x="4"
        y={42 + i * lineGap}
        width="32"
        height="4"
        rx="1"
      />
    {/each}
  </svg>
</div>
