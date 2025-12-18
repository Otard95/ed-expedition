<script lang="ts">
  export let size: string = "1.2rem";
  export let color: string = "currentColor";
</script>

<div style="width: {size}; height: {size}; display: inline-flex;">
  <svg
    width="100%"
    height="100%"
    viewBox="0 0 64 64"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
  >
  <!-- Diamond/Arrow shape -->
  <g transform="matrix(1.18775,0,0,1.1776,-7.32882,-10.4176)">
    <path
      d="M33.112,22.433L45.741,17.338L33.112,61.496L20.483,17.338L33.112,22.433Z"
      fill={color}
    />
  </g>

  <!-- Top arc -->
  <path
    d="M17.634,11.042L32,4L46.342,11.026"
    fill="none"
    stroke={color}
    stroke-width="2.44"
  />
  </svg>
</div>
