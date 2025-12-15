<script lang="ts">
  export let direction: "left" | "right" | "up" | "down" = "right";
  export let color: string = "currentColor";
  export let size: string = "0.875rem";

  const rotations: Record<typeof direction, number> = {
    left: 0,
    right: 180,
    up: 90,
    down: -90,
  };

  $: rotation = rotations[direction];
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 1250 1250"
  style="transform: rotate({rotation}deg); vertical-align: middle; display: inline-block; position: relative; top: -1px;"
  xmlns="http://www.w3.org/2000/svg"
>
  <g transform="matrix(0.914463,0,0,1.23355,222.345,178.964)">
    <path
      d="M437.13,465.008L437.13,657.076L-184.681,361.586L437.13,66.096L437.13,258.165L1065.32,258.165L1065.32,465.008L437.13,465.008Z"
      fill={color}
      stroke={color}
      stroke-width="39.91"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
  </g>
</svg>
