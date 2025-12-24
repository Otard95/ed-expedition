<script lang="ts">
  import type { ToastLevel, ToastAction } from "../lib/stores/toast";
  import { toasts } from "../lib/stores/toast";

  export let id: string;
  export let message: string;
  export let level: ToastLevel = "info";
  export let dismissable: boolean = true;
  export let action: ToastAction | undefined = undefined;

  const levelStyles: Record<ToastLevel, { bg: string; border: string; text: string }> = {
    info: { bg: 'rgb(from var(--ed-info) r g b / 0.15)', border: 'var(--ed-info)', text: 'var(--ed-info)' },
    success: { bg: 'rgb(from var(--ed-success) r g b / 0.15)', border: 'var(--ed-success)', text: 'var(--ed-success)' },
    warning: { bg: 'rgb(from var(--ed-warning) r g b / 0.15)', border: 'var(--ed-warning)', text: 'var(--ed-warning)' },
    danger: { bg: 'rgb(from var(--ed-danger) r g b / 0.15)', border: 'var(--ed-danger)', text: 'var(--ed-danger)' },
  };

  $: styles = levelStyles[level];

  function dismiss() {
    toasts.dismiss(id);
  }
</script>

<div
  class="toast"
  style="background-color: {styles.bg}; border-color: {styles.border}; color: {styles.text};"
>
  <span class="message">{message}</span>
  {#if action}
    <button class="action" on:click={action.callback} style="color: {styles.text};">{action.cta}</button>
  {/if}
  {#if dismissable}
    <button class="dismiss" on:click={dismiss} style="color: {styles.text};">×</button>
  {/if}
</div>

<style>
  .toast {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    border: 1px solid;
    border-radius: 4px;
    box-shadow: var(--ed-shadow-md);
    min-width: 250px;
    max-width: 400px;
  }

  .message {
    flex: 1;
  }

  .action {
    background: none;
    border: 1px solid currentColor;
    border-radius: 2px;
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
    font-weight: 600;
    cursor: pointer;
    opacity: 0.9;
    transition: opacity 0.15s;
  }

  .action:hover {
    opacity: 1;
  }

  .dismiss {
    background: none;
    border: none;
    font-size: 1.25rem;
    line-height: 1;
    cursor: pointer;
    padding: 0;
    opacity: 0.7;
    transition: opacity 0.15s;
  }

  .dismiss:hover {
    opacity: 1;
  }
</style>
