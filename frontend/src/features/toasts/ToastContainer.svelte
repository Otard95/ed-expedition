<script lang="ts">
  import { toasts } from "../../lib/stores/toast";
  import Toast from "../../components/Toast.svelte";
</script>

<div class="toast-container flex-col flex-gap-sm">
  {#each $toasts as toast (toast.id)}
    <Toast
      id={toast.id}
      message={toast.message}
      level={toast.level}
      dismissable={toast.dismissable}
      action={toast.action}
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
