<script lang="ts">
  import Tooltip from "./Tooltip.svelte";

  export let value: boolean = false;
  export let label: string = "";
  export let disabled: boolean = false;
  export let info: string | undefined = undefined;

  let className: string = "";
  export { className as class };
</script>

<div class="toggle-container {className}">
  <label class="toggle-label">
    {#if label}
      <span class="label-text">
        {label}
        {#if info}
          <Tooltip text={info} />
        {/if}
      </span>
    {/if}
    <input
      type="checkbox"
      bind:checked={value}
      {disabled}
    />
    <span class="toggle-switch">
      <span class="toggle-slider"></span>
    </span>
  </label>
</div>

<style>
  .toggle-container {
    display: flex;
    flex-direction: column;
  }

  .toggle-label {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    cursor: pointer;
    user-select: none;
  }

  .toggle-label.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  input[type="checkbox"] {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
  }

  .toggle-switch {
    position: relative;
    width: 3rem;
    height: 1.5rem;
    background: var(--ed-bg-tertiary);
    border: 1px solid var(--ed-border);
    border-radius: 1rem;
    transition: all 0.2s ease;
  }

  .toggle-slider {
    position: absolute;
    top: 2px;
    left: 2px;
    width: 1.25rem;
    height: 1.25rem;
    background: var(--ed-text-dim);
    border-radius: 50%;
    transition: all 0.2s ease;
  }

  input:checked + .toggle-switch {
    background: var(--ed-orange);
    border-color: var(--ed-orange);
  }

  input:checked + .toggle-switch .toggle-slider {
    transform: translateX(1.5rem);
    background: white;
  }

  input:focus + .toggle-switch {
    box-shadow: 0 0 0 2px rgb(from var(--ed-orange) r g b / 0.2);
  }

  input:disabled + .toggle-switch {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .label-text {
    color: var(--ed-orange-dim);
    font-size: 0.875rem;
    font-weight: 500;
  }
</style>
