<script lang="ts">
  export let value: string = "";
  export let label: string = "";
  export let disabled: boolean = false;
  export let options: Array<{ value: string; label: string }> = [];

  let className: string = "";
  export { className as class };
</script>

<div class="select-container {className}">
  {#if label}
    <label class="form-field">
      <span class="label-text">{label}</span>
      <select bind:value {disabled}>
        {#each options as option}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </label>
  {:else}
    <select bind:value {disabled}>
      {#each options as option}
        <option value={option.value}>{option.label}</option>
      {/each}
    </select>
  {/if}
</div>

<style>
  .select-container {
    display: flex;
    flex-direction: column;
  }

  .label-text {
    color: var(--ed-orange-dim);
    font-size: 0.875rem;
    font-weight: 500;
  }

  select {
    appearance: none;
    -webkit-appearance: none;
    -moz-appearance: none;
    background: var(--ed-bg-primary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    padding: 0.5rem 0.75rem;
    color: var(--ed-text-primary);
    font-size: 1rem;
    cursor: pointer;
    width: 100%;
    box-sizing: border-box;
  }

  select:focus {
    outline: none;
    border-color: var(--ed-orange);
  }

  select:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  option {
    background: var(--ed-bg-primary);
    background-color: var(--ed-bg-primary);
    color: var(--ed-text-primary);
    padding: 0.5rem;
  }

  option:hover {
    background: var(--ed-bg-tertiary);
  }

  option:checked {
    background: var(--ed-orange);
    color: var(--ed-text-on-orange);
  }
</style>
