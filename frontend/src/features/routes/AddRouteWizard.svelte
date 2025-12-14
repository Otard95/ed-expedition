<script lang="ts">
  import { onMount } from "svelte";
  import Button from "../../components/Button.svelte";
  import TextInput from "../../components/TextInput.svelte";
  import PlotterInput from "../../components/PlotterInput.svelte";
  import { GetPlotterOptions, GetPlotterInputConfig } from "../../../wailsjs/go/main/App";
  import type { plotters } from "../../../wailsjs/go/models";

  export let canClose: boolean = true;
  export let onComplete: ((route: any) => void) | undefined = undefined;
  export let onCancel: (() => void) | undefined = undefined;

  type WizardStep = "select-plotter" | "configure" | "plotting" | "success";
  let currentStep: WizardStep = "select-plotter";

  let plotterOptions: Record<string, string> = {};
  let selectedPlotterId: string = "";
  let loadingPlotters = true;
  let plotterError: string | null = null;

  // Route configuration
  let fromSystem: string = "";
  let toSystem: string = "";
  let plotterInputConfig: plotters.PlotterInputFieldConfig[] | null = null;
  let inputValues: Record<string, string> = {};

  $: canClose = currentStep !== "plotting";

  onMount(async () => {
    try {
      plotterOptions = await GetPlotterOptions();
      loadingPlotters = false;
    } catch (err) {
      plotterError = err instanceof Error ? err.message : "Failed to load plotters";
      loadingPlotters = false;
    }
  });

  async function loadPlotterConfig() {
    if (!selectedPlotterId) return;

    try {
      plotterInputConfig = await GetPlotterInputConfig(selectedPlotterId);

      // Pre-populate all input fields with their defaults to ensure inputValues
      // contains all required keys before step 2 renders. PlotterInput components
      // will bind to these values and update them as the user interacts.
      inputValues = {};
      if (plotterInputConfig) {
        for (const field of plotterInputConfig) {
          inputValues[field.name] = field.default;
        }
      }
    } catch (err) {
      console.error("Failed to load plotter config:", err);
      plotterInputConfig = null;
    }
  }

  async function handleNext() {
    if (currentStep === "select-plotter") {
      await loadPlotterConfig();
      currentStep = "configure";
    } else if (currentStep === "configure") {
      currentStep = "plotting";
    } else if (currentStep === "plotting") {
      currentStep = "success";
    } else if (currentStep === "success" && onComplete) {
      onComplete({ id: "mock-route" });
    }
  }

  function handleBack() {
    if (currentStep === "configure") {
      currentStep = "select-plotter";
    } else if (currentStep === "plotting") {
      currentStep = "configure";
    } else if (currentStep === "success") {
      currentStep = "configure";
    }
  }

  function handleCancel() {
    if (onCancel) {
      onCancel();
    }
  }

  $: showBack = currentStep !== "select-plotter" && currentStep !== "plotting";
  $: showNext = currentStep !== "success";
  $: showFinish = currentStep === "success";
  $: showCancel = currentStep !== "plotting";
  $: nextLabel = currentStep === "configure" ? "Plot Route" : "Next";
  $: canGoNext =
    currentStep === "select-plotter"
      ? selectedPlotterId !== ""
      : currentStep === "configure"
        ? fromSystem !== "" && toSystem !== ""
        : true;
</script>

<div class="wizard">
  <div class="wizard-content">
    {#if currentStep === "select-plotter"}
      <div class="step-content">
        <h3>Step 1: Select Plotter</h3>
        {#if loadingPlotters}
          <p class="loading">Loading plotters...</p>
        {:else if plotterError}
          <p class="error">Error: {plotterError}</p>
        {:else}
          <p>Choose a plotter to generate your route:</p>
          <div class="plotter-options">
            {#each Object.entries(plotterOptions) as [id, name]}
              <label class="plotter-option">
                <input
                  type="radio"
                  name="plotter"
                  value={id}
                  bind:group={selectedPlotterId}
                />
                <span class="plotter-name">{name}</span>
              </label>
            {/each}
          </div>
          <p class="disclaimer">More plotters coming soon</p>
        {/if}
      </div>
    {:else if currentStep === "configure"}
      <div class="step-content">
        <h3>Step 2: Configure Route</h3>
        <div class="input-grid">
          <TextInput bind:value={fromSystem} label="From System" placeholder="Sol" />
          <TextInput bind:value={toSystem} label="To System" placeholder="Colonia" />

          {#if plotterInputConfig}
            {#each plotterInputConfig as field}
              <PlotterInput {field} bind:value={inputValues[field.name]} />
            {/each}
          {/if}
        </div>
      </div>
    {:else if currentStep === "plotting"}
      <div class="step-content plotting">
        <h3>Step 3: Plotting Route</h3>
        <div class="loading-spinner"></div>
        <p>Please wait while we plot your route...</p>
      </div>
    {:else if currentStep === "success"}
      <div class="step-content">
        <h3>Success!</h3>
        <p>Mock content: Route plotted successfully</p>
        <div class="result-values">
          <p><strong>From:</strong> {fromSystem}</p>
          <p><strong>To:</strong> {toSystem}</p>
          <p><strong>Plotter:</strong> {plotterOptions[selectedPlotterId]}</p>
          {#if plotterInputConfig}
            {#each plotterInputConfig as field}
              <p><strong>{field.label}:</strong> {inputValues[field.name]}</p>
            {/each}
          {/if}
        </div>
      </div>
    {/if}
  </div>

  <div class="wizard-actions">
    {#if showBack}
      <Button variant="secondary" onClick={handleBack}>Back</Button>
    {/if}
    <div class="spacer"></div>
    {#if showCancel}
      <Button variant="secondary" onClick={handleCancel}>Cancel</Button>
    {/if}
    {#if showNext}
      <Button variant="primary" onClick={handleNext} disabled={!canGoNext}>{nextLabel}</Button>
    {/if}
    {#if showFinish}
      <Button variant="primary" onClick={handleNext}>Finish</Button>
    {/if}
  </div>
</div>

<style>
  .wizard {
    display: flex;
    flex-direction: column;
    gap: 2rem;
    min-width: 500px;
  }

  .wizard-content {
    min-height: 200px;
  }

  .step-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .step-content.plotting {
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 2rem 0;
  }

  .step-content h3 {
    margin: 0;
    color: var(--ed-orange);
    font-size: 1.125rem;
  }

  .step-content p {
    margin: 0;
    color: var(--ed-text-primary);
  }

  .step-content em {
    color: var(--ed-text-secondary);
    font-size: 0.875rem;
  }

  .step-content label {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    color: var(--ed-text-secondary);
    font-size: 0.875rem;
  }

  .step-content input {
    background: var(--ed-bg-primary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    padding: 0.5rem;
    color: var(--ed-text-primary);
    font-size: 1rem;
  }

  .step-content input:focus {
    outline: none;
    border-color: var(--ed-orange);
  }

  .step-content strong {
    color: var(--ed-orange);
  }

  .step-content .loading {
    color: var(--ed-text-secondary);
    font-style: italic;
  }

  .step-content .error {
    color: var(--ed-danger);
  }

  .step-content .disclaimer {
    color: var(--ed-text-dim);
    font-size: 0.875rem;
    font-style: italic;
    margin-top: 0.5rem;
  }

  .input-grid {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-top: 1rem;
  }

  .result-values {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-top: 1rem;
  }

  .plotter-options {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .plotter-option {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem;
    background: var(--ed-bg-primary);
    border: 2px solid var(--ed-border);
    border-radius: 2px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .plotter-option:hover {
    background: var(--ed-bg-tertiary);
    border-color: var(--ed-orange);
  }

  .plotter-option:has(input:checked) {
    background: var(--ed-bg-tertiary);
    border-color: var(--ed-orange);
    box-shadow: 0 0 0 1px var(--ed-orange);
  }

  .plotter-option input[type="radio"] {
    position: absolute;
    opacity: 0;
    width: 0;
    height: 0;
  }

  .plotter-name {
    color: var(--ed-text-primary);
    font-size: 1rem;
  }

  .wizard-actions {
    display: flex;
    gap: 0.75rem;
    padding-top: 1rem;
    border-top: 1px solid var(--ed-border);
  }

  .spacer {
    flex: 1;
  }

  .loading-spinner {
    width: 3rem;
    height: 3rem;
    border: 3px solid var(--ed-border);
    border-top-color: var(--ed-orange);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
