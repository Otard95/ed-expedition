<script lang="ts">
  import { onMount } from 'svelte';

  export let options: IntersectionObserverInit = {};
  let className = '';
  export { className as class };

  let ratio = 1;
  let intersecting = true;
  let container: HTMLElement;

  onMount(() => {
    if (typeof IntersectionObserver === 'undefined') return;

    const observer = new IntersectionObserver((entries) => {
      const entry = entries[0];
      ratio = entry.intersectionRatio;
      intersecting = entry.isIntersecting;
    }, options);

    observer.observe(container);
    return () => observer.unobserve(container);
  });
</script>

<div bind:this={container} class={className}>
  <slot {ratio} {intersecting} />
</div>
