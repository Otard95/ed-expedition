<script lang="ts">
  export let size: string = "1.3rem";
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
    <!-- Large circle (bottom left) -->
    <g transform="matrix(1.00321,0,0,1.00321,0.0425532,-0.177477)">
      <circle cx="18.897" cy="41.993" r="10.965" fill={color} />
    </g>

    <!-- Medium circle (top right) -->
    <g transform="matrix(1,0,0,1,0.106926,0.0595383)">
      <circle cx="47.903" cy="30.95" r="8.01" fill={color} />
    </g>

    <!-- Small circle (top center) -->
    <g transform="matrix(0.959004,0,0,0.959004,0.40678,0.559322)">
      <circle cx="34.508" cy="16.622" r="5.735" fill={color} />
    </g>

    <!-- Connection line (bottom) -->
    <path
      d="M27.684,35.927L40.437,31.028L41.993,34.45L29.55,40.826L27.684,35.927Z"
      fill={color}
    />

    <!-- Connection line (top) -->
    <path
      d="M41.837,26.673L36.005,20.996L38.182,18.586L44.559,24.262L41.837,26.673Z"
      fill={color}
    />
  </svg>
</div>
