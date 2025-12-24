<script lang="ts">
  import type { ToastLevel, ToastAction } from "../lib/stores/toast";
  import { toasts } from "../lib/stores/toast";
  import Card from "./Card.svelte";
  import Button from "./Button.svelte";

  export let id: string;
  export let message: string;
  export let level: ToastLevel = "info";
  export let dismissable: boolean = true;
  export let action: ToastAction | undefined = undefined;
  export let title: string | undefined = undefined;

  const levelColors: Record<ToastLevel, string> = {
    info: 'var(--ed-info)',
    success: 'var(--ed-success)',
    warning: 'var(--ed-warning)',
    danger: 'var(--ed-danger)',
  };

  function dismiss() {
    toasts.dismiss(id);
  }
</script>

<Card class="toast" padding="0.75rem 1rem">
  <div class="level-bar" style="background: {levelColors[level]}"></div>
  <div class="flex-center flex-gap-sm">
    <div class="content">
      {#if title}
        <div class="title" style="color: {levelColors[level]}">{title}</div>
      {/if}
      <span class="message" class:text-secondary={title}>{message}</span>
    </div>
    {#if action}
      <Button variant="secondary" size="small" onClick={action.callback}>{action.cta}</Button>
    {/if}
    {#if dismissable}
      <button class="dismiss" on:click={dismiss}>×</button>
    {/if}
  </div>
</Card>

<style>
  :global(.toast) {
    min-width: 250px;
    max-width: 400px;
    overflow: hidden;
  }

  .level-bar {
    height: 2px;
    margin: -0.75rem -1rem 0.75rem -1rem;
  }

  .content {
    flex: 1;
    text-align: left;
  }

  .title {
    font-weight: 600;
    margin-bottom: 0.25rem;
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
    color: var(--ed-text-secondary);
  }

  .dismiss:hover {
    opacity: 1;
  }
</style>
