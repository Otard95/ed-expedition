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
  export let animate: boolean = false;

  const levelColors: Record<ToastLevel, string> = {
    info: "var(--ed-info)",
    success: "var(--ed-success)",
    warning: "var(--ed-warning)",
    danger: "var(--ed-danger)",
  };

  function dismiss() {
    toasts.dismiss(id);
  }
</script>

<Card class="toast" padding="0.75rem 1rem">
  <div
    class="level-bar"
    class:animate
    style="--level-color: {levelColors[level]}"
  ></div>
  <div class="flex-center flex-gap-sm">
    <div class="content">
      {#if title}
        <div class="title" style="color: {levelColors[level]}">{title}</div>
      {/if}
      <span class="message" class:text-secondary={title}>{message}</span>
    </div>
    {#if action}
      <Button variant="secondary" size="small" onClick={action.callback}
        >{action.cta}</Button
      >
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
    box-sizing: border-box;
  }

  .level-bar {
    height: 2px;
    margin: -0.75rem -1rem 0.75rem -1rem;
    background: var(--level-color);
  }

  .level-bar.animate {
    position: relative;
    width: 400px;
  }

  .level-bar.animate::before,
  .level-bar.animate::after {
    content: "";
    position: absolute;
    top: -38px;
    width: 4px;
    height: 4px;
    border-radius: 50%;
    background: var(--level-color);
    box-shadow:
      0 0 35px 35px var(--level-color),
      60px 0 35px 35px var(--level-color),
      120px 0 35px 35px var(--level-color),
      180px 0 35px 35px var(--level-color),
      240px 0 35px 35px var(--level-color),
      300px 0 35px 35px var(--level-color),
      360px 0 35px 35px var(--level-color),
      420px 0 35px 35px var(--level-color);
    animation: glow-sweep 6s linear infinite;
  }

  .level-bar.animate::after {
    animation-delay: -3s;
  }

  @keyframes glow-sweep {
    0% {
      left: -120%;
    }
    100% {
      left: 120%;
    }
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
