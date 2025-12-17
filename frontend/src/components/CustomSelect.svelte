<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Chevron from "./Chevron.svelte";
  import Tooltip from "./Tooltip.svelte";

  export let value: string = "";
  export let label: string = "";
  export let disabled: boolean = false;
  export let options: Array<{ value: string; label: string; description?: string }> = [];
  export let info: string | undefined = undefined;

  let className: string = "";
  export { className as class };

  let isOpen = false;
  let dropdownRef: HTMLDivElement;

  $: selectedOption = options.find((opt) => opt.value === value);
  $: displayLabel = selectedOption?.label || "Select...";
  // Tooltip shows the selected option's description (prepended with label for context)
  // if available, otherwise falls back to the field's general info text.
  // This way users see specific info about their current selection rather than generic field help.
  $: selectedTooltip = selectedOption?.description
    ? `${selectedOption.label} — ${selectedOption.description}`
    : info;

  function toggleDropdown() {
    if (!disabled) {
      isOpen = !isOpen;
    }
  }

  function selectOption(optionValue: string) {
    value = optionValue;
    isOpen = false;
  }

  function handleClickOutside(event: MouseEvent) {
    if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
      isOpen = false;
    }
  }

  function handleEscape(event: KeyboardEvent) {
    if (event.key === "Escape" && isOpen) {
      isOpen = false;
    }
  }

  onMount(() => {
    document.addEventListener("click", handleClickOutside);
    document.addEventListener("keydown", handleEscape);
  });

  onDestroy(() => {
    document.removeEventListener("click", handleClickOutside);
    document.removeEventListener("keydown", handleEscape);
  });
</script>

<div class="dropdown-container form-field {className}" bind:this={dropdownRef}>
  {#if label}
    <span class="label-text">
      {label}
      {#if selectedTooltip}
        <Tooltip text={selectedTooltip} />
      {/if}
    </span>
  {/if}
  <div class="dropdown">
    <button
      type="button"
      class="dropdown-trigger"
      class:disabled
      class:open={isOpen}
      on:click={toggleDropdown}
      {disabled}
    >
      <span class="selected-label">{displayLabel}</span>
      <Chevron direction={isOpen ? "up" : "down"} size="16px" color="var(--ed-text-secondary)" />
    </button>
    {#if isOpen}
      <div class="dropdown-menu">
        {#each options as option}
          <button
            type="button"
            class="dropdown-option"
            class:selected={option.value === value}
            on:click={() => selectOption(option.value)}
          >
            {option.label}
          </button>
        {/each}
      </div>
    {/if}
  </div>
</div>

<style>
  .dropdown-container {
    position: relative;
  }

  .label-text {
    color: var(--ed-orange-dim);
    font-size: 0.875rem;
    font-weight: 500;
  }

  .dropdown {
    width: 100%;
    position: relative;
  }

  .dropdown-trigger {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    background: var(--ed-bg-primary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    padding: 0.5rem 0.75rem;
    color: var(--ed-text-primary);
    font-size: 1rem;
    cursor: pointer;
    box-sizing: border-box;
    transition: border-color 0.15s ease;
  }

  .dropdown-trigger:hover:not(.disabled) {
    border-color: var(--ed-orange);
  }

  .dropdown-trigger.open {
    border-color: var(--ed-orange);
  }

  .dropdown-trigger.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .selected-label {
    flex: 1;
    text-align: left;
  }

  .dropdown-menu {
    position: absolute;
    top: calc(100% + 0.25rem);
    left: 0;
    right: 0;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-orange);
    border-radius: 2px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.5);
    z-index: 1000;
    max-height: 200px;
    overflow-y: auto;
  }

  .dropdown-option {
    width: 100%;
    display: block;
    background: none;
    border: none;
    padding: 0.75rem 1rem;
    color: var(--ed-text-primary);
    font-size: 1rem;
    text-align: left;
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  .dropdown-option:hover {
    background: var(--ed-bg-tertiary);
  }

  .dropdown-option.selected {
    background: var(--ed-orange);
    color: #000;
    font-weight: 600;
  }

  .dropdown-option.selected:hover {
    background: var(--ed-orange-bright);
  }
</style>
