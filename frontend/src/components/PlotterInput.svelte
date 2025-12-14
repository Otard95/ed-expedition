<script lang="ts">
  import { onMount } from "svelte";
  import TextInput from "./TextInput.svelte";
  import NumberInput from "./NumberInput.svelte";
  import Toggle from "./Toggle.svelte";
  import CustomSelect from "./CustomSelect.svelte";
  import type { plotters } from "../../wailsjs/go/models";

  export let field: plotters.PlotterInputFieldConfig;
  export let value: string;

  let className: string = "";
  export { className as class };

  $: label = field.label;

  // Determine which component to use
  $: hasOptions = field.options && field.options.length > 0;
  $: isSelect = hasOptions;
  $: isBoolean = field.type === "boolean";
  $: isNumber = field.type === "number" && !hasOptions;
  $: isString = field.type === "string" && !hasOptions;

  // We need local mutable variables because child components use bind:value,
  // but we can't directly bind to reactive derived values from the parent.
  // This creates a two-way sync: parent value → local vars (in onMount),
  // then local vars → parent value (via reactive statements).
  let boolValue: boolean = false;
  let numberValue: number = 0;
  let stringValue: string = "";
  let selectValue: string = "";
  let initialized = false;

  // Initialize local values from parent in onMount to avoid circular dependency.
  // If we used reactive statements ($: localValue = parentValue), Svelte's compiler
  // would detect a cycle: parentValue depends on localValue (in the sync-back statements),
  // and localValue depends on parentValue (in the derive statement).
  onMount(() => {
    boolValue = value === "1";
    numberValue = parseFloat(value) || parseFloat(field.default);
    stringValue = value || field.default;
    selectValue = value || field.default;
    initialized = true;
  });

  // Sync local changes back to parent value, gated by initialized flag.
  // We gate on initialized because reactive statements run immediately on component creation,
  // before onMount, which would update the parent with uninitialized default values (0, "", false).
  $: if (initialized && isBoolean) {
    value = boolValue ? "1" : "0";
  }
  $: if (initialized && isNumber) {
    value = String(numberValue);
  }
  $: if (initialized && isString) {
    value = stringValue;
  }
  $: if (initialized && isSelect) {
    value = selectValue;
  }
</script>

<div class="plotter-input {className}">
  {#if isBoolean}
    <Toggle
      bind:value={boolValue}
      {label}
      info={field.info}
    />
  {:else if isNumber}
    <NumberInput
      bind:value={numberValue}
      {label}
      info={field.info}
    />
  {:else if isSelect}
    <CustomSelect
      bind:value={selectValue}
      {label}
      info={field.info}
      options={field.options.map((opt) => ({
        value: opt.value,
        label: opt.label,
        description: opt.description,
      }))}
    />
  {:else}
    <TextInput
      bind:value={stringValue}
      {label}
      info={field.info}
    />
  {/if}
</div>

<style>
  .plotter-input {
    display: flex;
    flex-direction: column;
  }
</style>
