<script lang="ts">
  import { toasts } from "../../lib/stores/toast";
  import Toast from "../../components/Toast.svelte";

  $: temporary = $toasts.filter(t => !t.persistent);
  $: persistent = $toasts.filter(t => t.persistent);
</script>

<div class="toast-container flex-col flex-gap-sm">
  {#each temporary as toast (toast.id)}
    <Toast
      id={toast.id}
      message={toast.message}
      level={toast.level}
      dismissable={toast.dismissable}
      action={toast.action}
      title={toast.title}
      animate={toast.animate}
    />
  {/each}
  {#each persistent as toast (toast.id)}
    <Toast
      id={toast.id}
      message={toast.message}
      level={toast.level}
      dismissable={toast.dismissable}
      action={toast.action}
      title={toast.title}
      animate={toast.animate}
    />
  {/each}
</div>

<style>
  .toast-container {
    position: fixed;
    bottom: 1rem;
    right: 1rem;
    z-index: 1000;
  }
</style>
