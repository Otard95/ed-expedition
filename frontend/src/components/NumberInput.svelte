<script lang="ts">
  import Tooltip from "./Tooltip.svelte";

  export let value: number = 0;
  export let label: string = "";
  export let placeholder: string = "";
  export let disabled: boolean = false;
  export let min: number | undefined = undefined;
  export let max: number | undefined = undefined;
  export let step: number | undefined = undefined;
  export let info: string | undefined = undefined;

  let className: string = "";
  export { className as class };
</script>

<div class="number-input {className}">
  {#if label}
    <label>
      <span class="label-text">
        {label}
        {#if info}
          <Tooltip text={info} />
        {/if}
      </span>
      <input
        type="number"
        bind:value
        {placeholder}
        {disabled}
        {min}
        {max}
        {step}
      />
    </label>
  {:else}
    <input
      type="number"
      bind:value
      {placeholder}
      {disabled}
      {min}
      {max}
      {step}
    />
  {/if}
</div>

<style>
  .number-input {
    display: flex;
    flex-direction: column;
  }

  label {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
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

  /* Hide number input spinners */
  input[type="number"]::-webkit-inner-spin-button,
  input[type="number"]::-webkit-outer-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  input[type="number"] {
    -moz-appearance: textfield;
  }
</style>
