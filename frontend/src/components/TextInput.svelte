<script lang="ts">
  import Tooltip from "./Tooltip.svelte";

  export let value: string = "";
  export let label: string = "";
  export let placeholder: string = "";
  export let disabled: boolean = false;
  export let info: string | undefined = undefined;

  let className: string = "";
  export { className as class };
</script>

<div class="text-input {className}">
  {#if label}
    <label class="form-field">
      <span class="label-text">
        {label}
        {#if info}
          <Tooltip text={info} />
        {/if}
      </span>
      <input
        type="text"
        bind:value
        {placeholder}
        {disabled}
      />
    </label>
  {:else}
    <input
      type="text"
      bind:value
      {placeholder}
      {disabled}
    />
  {/if}
</div>

<style>
  .text-input {
    display: flex;
    flex-direction: column;
  }

  .label-text {
    color: var(--ed-orange-dim);
    font-size: 0.875rem;
    font-weight: 500;
  }

  input {
    background: var(--ed-bg-primary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    padding: 0.5rem 0.75rem;
    color: var(--ed-text-primary);
    font-size: 1rem;
    width: 100%;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: var(--ed-orange);
  }

  input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  input::placeholder {
    color: var(--ed-text-dim);
  }
</style>
