<script lang="ts">
  import { onDestroy } from "svelte";
  import Tooltip from "./Tooltip.svelte";
  import { debounce } from "../lib/utils/debounce";

  export let value: string = "";
  export let label: string = "";
  export let placeholder: string = "";
  export let disabled: boolean = false;
  export let info: string | undefined = undefined;

  export let fetchSuggestions: (prefix: string) => Promise<string[]>;
  export let validate:
    | ((value: string) => Promise<{ valid: boolean; message?: string }>)
    | undefined = undefined;
  export let minChars: number = 2;
  export let debounceMs: number = 150;

  let suggestions: string[] = [];
  let showDropdown = false;
  let activeIndex = -1;
  let validation: { valid: boolean; message?: string } | null = null;

  let dropdownMousedown = false;
  let inputEl: HTMLInputElement;
  let listEl: HTMLUListElement;

  async function search(query: string) {
    try {
      suggestions = await fetchSuggestions(query);
      showDropdown = suggestions.length > 0;
      activeIndex = -1;
    } catch {
      suggestions = [];
      showDropdown = false;
    }
  }

  const debouncedSearch = debounce((query: string) => search(query), { ms: debounceMs });
  onDestroy(() => debouncedSearch.cancel());

  function handleInput() {
    validation = null;

    if (value.length < minChars) {
      debouncedSearch.cancel();
      suggestions = [];
      showDropdown = false;
      return;
    }

    debouncedSearch(value);
  }

  function selectItem(name: string) {
    value = name;
    suggestions = [];
    showDropdown = false;
    activeIndex = -1;
    validation = null;
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!showDropdown) return;

    if (e.key === "ArrowDown") {
      e.preventDefault();
      activeIndex = Math.min(activeIndex + 1, suggestions.length - 1);
      scrollActiveIntoView();
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      activeIndex = Math.max(activeIndex - 1, -1);
      scrollActiveIntoView();
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (activeIndex >= 0 && activeIndex < suggestions.length) {
        selectItem(suggestions[activeIndex]);
      }
    } else if (e.key === "Escape") {
      showDropdown = false;
      activeIndex = -1;
    }
  }

  function scrollActiveIntoView() {
    if (!listEl || activeIndex < 0) return;
    const item = listEl.children[activeIndex] as HTMLElement | undefined;
    if (item) {
      item.scrollIntoView({ block: "nearest" });
    }
  }

  function handleDropdownMousedown() {
    dropdownMousedown = true;
  }

  async function handleBlur() {
    // Allow click on dropdown item to register before closing
    if (dropdownMousedown) {
      dropdownMousedown = false;
      inputEl?.focus();
      return;
    }

    showDropdown = false;
    activeIndex = -1;

    if (validate && value.length > 0) {
      try {
        validation = await validate(value);
      } catch {
        validation = null;
      }
    }
  }

  function handleFocus() {
    if (suggestions.length > 0 && value.length >= minChars) {
      showDropdown = true;
    }
  }

  $: hasError = validation !== null && !validation.valid;
</script>

<div class="autocomplete-input flex-col text-left">
  {#if label}
    <label class="form-field">
      <span class="label-text">
        {label}
        {#if info}
          <Tooltip text={info} />
        {/if}
      </span>
      <input
        bind:this={inputEl}
        type="text"
        bind:value
        {placeholder}
        {disabled}
        class:has-error={hasError}
        on:input={handleInput}
        on:keydown={handleKeydown}
        on:blur={handleBlur}
        on:focus={handleFocus}
        autocomplete="off"
        role="combobox"
        aria-expanded={showDropdown}
        aria-autocomplete="list"
        aria-activedescendant={activeIndex >= 0
          ? `ac-option-${activeIndex}`
          : undefined}
      />
    </label>
  {:else}
    <input
      bind:this={inputEl}
      type="text"
      bind:value
      {placeholder}
      {disabled}
      class:has-error={hasError}
      on:input={handleInput}
      on:keydown={handleKeydown}
      on:blur={handleBlur}
      on:focus={handleFocus}
      autocomplete="off"
      role="combobox"
      aria-expanded={showDropdown}
      aria-autocomplete="list"
      aria-activedescendant={activeIndex >= 0
        ? `ac-option-${activeIndex}`
        : undefined}
    />
  {/if}

  {#if showDropdown}
    <ul
      class="dropdown"
      bind:this={listEl}
      role="listbox"
      on:mousedown={handleDropdownMousedown}
    >
      {#each suggestions as name, i}
        <li
          id="ac-option-{i}"
          class="dropdown-item"
          class:active={i === activeIndex}
          role="option"
          aria-selected={i === activeIndex}
          on:click={() => selectItem(name)}
          on:mouseenter={() => (activeIndex = i)}
        >
          {name}
        </li>
      {/each}
    </ul>
  {/if}

  <span class="error-text text-danger" class:visible={hasError && validation?.message}>
    {validation?.message ?? "\u00A0"}
  </span>
</div>

<style>
  .autocomplete-input {
    position: relative;
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

  input.has-error {
    border-color: var(--ed-danger);
  }

  .error-text {
    font-size: 0.75rem;
    margin-top: 0.25rem;
    visibility: hidden;
  }

  .error-text.visible {
    visibility: visible;
  }

  .dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    margin: 0;
    padding: 0;
    list-style: none;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border);
    border-top: none;
    max-height: 200px;
    overflow-y: auto;
    z-index: 100;
  }

  .dropdown-item {
    padding: 0.5rem 0.75rem;
    color: var(--ed-text-primary);
    cursor: pointer;
    font-size: 1rem;
  }

  .dropdown-item.active {
    background: var(--ed-bg-tertiary);
    color: var(--ed-orange);
  }
</style>
