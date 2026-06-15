<script lang="ts">
  import Tooltip from "./Tooltip.svelte";
  import Chevron from "./icons/Chevron.svelte";

  export let value: string = "";
  export let label: string = "";
  export let disabled: boolean = false;
  export let options: Array<{
    value: string;
    label: string;
    description?: string;
  }> = [];
  export let info: string | undefined = undefined;
  export let collapsed: boolean = true;

  let className: string = "";
  export { className as class };

  $: selected = new Set(value ? value.split(",") : []);
  $: summary = `${selected.size} / ${options.length}`;

  function toggle(optionValue: string) {
    if (disabled) return;
    const next = new Set(selected);
    if (next.has(optionValue)) {
      next.delete(optionValue);
    } else {
      next.add(optionValue);
    }
    value = options
      .map((o) => o.value)
      .filter((v) => next.has(v))
      .join(",");
  }

  function selectAll() {
    if (disabled) return;
    value = options.map((o) => o.value).join(",");
  }

  function selectNone() {
    if (disabled) return;
    value = "";
  }

  $: allSelected = selected.size === options.length;
  $: noneSelected = selected.size === 0;
</script>

<div class="multiselect-container form-field {className}" class:disabled>
  <button
    type="button"
    class="header"
    on:click={() => (collapsed = !collapsed)}
    {disabled}
  >
    <span class="label-row">
      <Chevron direction={collapsed ? "right" : "down"} size="14px" />
      <span class="label-text">
        {label}
        <span class="summary">{summary}</span>
        {#if info}
          <Tooltip text={info} />
        {/if}
      </span>
    </span>
  </button>

  {#if !collapsed}
    <div class="actions">
      <button
        type="button"
        class="action-link"
        disabled={disabled || allSelected}
        on:click={selectAll}>All</button
      >
      <span class="action-separator">|</span>
      <button
        type="button"
        class="action-link"
        disabled={disabled || noneSelected}
        on:click={selectNone}>None</button
      >
    </div>
    <div class="option-list">
      {#each options as option}
        <label class="option-row" class:checked={selected.has(option.value)}>
          <span class="option-label">{option.label}</span>
          <input
            type="checkbox"
            checked={selected.has(option.value)}
            {disabled}
            on:change={() => toggle(option.value)}
          />
          <span class="checkbox-visual">
            {#if selected.has(option.value)}<span class="checkbox-fill" />{/if}
          </span>
        </label>
      {/each}
    </div>
  {/if}
</div>

<style>
  .multiselect-container {
    display: flex;
    flex-direction: column;
  }

  .multiselect-container.disabled {
    opacity: 0.5;
    pointer-events: none;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    color: var(--ed-text-primary);
  }

  .label-row {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    color: var(--ed-orange-dim);
  }

  .label-text {
    font-size: 0.875rem;
    font-weight: 500;
  }

  .summary {
    color: var(--ed-text-secondary);
    font-size: 0.8125rem;
  }

  .actions {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0 0.8rem;
    margin-top: 0.5rem;
    margin-bottom: 0.25rem;
  }

  .action-link {
    background: none;
    border: none;
    color: var(--ed-text-secondary);
    font-size: 0.75rem;
    cursor: pointer;
    padding: 0;
    text-decoration: underline;
    transition: color 0.15s ease;
  }

  .action-link:hover:not(:disabled) {
    color: var(--ed-orange);
  }

  .action-link:disabled {
    cursor: default;
    text-decoration: none;
    opacity: 0.5;
  }

  .action-separator {
    color: var(--ed-text-dim);
    font-size: 0.75rem;
  }

  .option-list {
    width: 100%;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    gap: 0.3rem;
  }

  .option-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    cursor: pointer;
    padding: 0 0.8rem;
    border-radius: 2px;
    transition: background-color 0.1s ease;
  }

  .option-row:hover .option-label {
    color: var(--ed-orange) !important;
  }

  .option-label {
    color: var(--ed-text-secondary);
    font-size: 0.875rem;
    transition: color 0.1s ease;
  }

  .option-row.checked .option-label {
    color: var(--ed-text-primary);
  }

  .option-row input[type="checkbox"] {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
  }

  .checkbox-visual {
    position: relative;
    width: 1.25rem;
    height: 1.25rem;
    flex-shrink: 0;
  }

  .checkbox-visual::before,
  .checkbox-visual::after {
    content: "";
    position: absolute;
    left: 0;
    right: 0;
    height: 30%;
    border: 1.5px solid var(--ed-text-dim);
    transition: border-color 0.15s ease;
  }

  .checkbox-visual::before {
    top: 0;
    border-bottom: none;
  }

  .checkbox-visual::after {
    bottom: 0;
    border-top: none;
  }

  .option-row.checked .checkbox-visual::before,
  .option-row.checked .checkbox-visual::after {
    border-color: var(--ed-text-secondary);
  }

  .checkbox-fill {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 45%;
    height: 45%;
    background: var(--ed-orange);
  }
</style>
