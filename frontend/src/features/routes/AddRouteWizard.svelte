<script lang="ts">
  import { onMount } from "svelte";
  import Button from "../../components/Button.svelte";
  import TextInput from "../../components/TextInput.svelte";
  import PlotterInput from "../../components/PlotterInput.svelte";
  import ConfirmDialog from "../../components/ConfirmDialog.svelte";
  import { GetPlotterOptions, GetPlotterInputConfig, PlotRoute } from "../../../wailsjs/go/main/App";
  import type { plotters, models } from "../../../wailsjs/go/models";

  export let expeditionId: string;
  export let canClose: boolean = true;
  export let initialFrom: string | undefined = undefined;
  export let onComplete: ((route: any) => void) | undefined = undefined;
  export let onCancel: (() => void) | undefined = undefined;

  type WizardStep = "select-plotter" | "configure" | "plotting" | "success";
  let currentStep: WizardStep = "select-plotter";

  let plotterOptions: Record<string, string> = {};
  let selectedPlotterId: string = "";
  let loadingPlotters = true;
  let plotterError: string | null = null;

  // Route configuration
  let fromSystem: string = initialFrom || "";
  let toSystem: string = "";
  let plotterInputConfig: plotters.PlotterInputFieldConfig[] | null = null;
  let inputValues: Record<string, string> = {};

  let plottedRoute: models.Route | null = null;
  let plottingError: string | null = null;
  let showCancelConfirm = false;

  $: canClose = currentStep !== "plotting" && currentStep !== "success";

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
      plottingError = null;

      try {
        plottedRoute = await PlotRoute(
          expeditionId,
          selectedPlotterId,
          fromSystem,
          toSystem,
          inputValues
        );
        currentStep = "success";
      } catch (err) {
        plottingError = err instanceof Error
          ? err.message
          : typeof err === 'string'
            ? err
            : String(err);
        console.error("Failed to plot route:", err);
        currentStep = "configure";
      }
    } else if (currentStep === "success" && onComplete) {
      onComplete(plottedRoute);
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

  function handleCancelPlot() {
    showCancelConfirm = true;
  }

  function confirmCancelPlot() {
    currentStep = "configure";
    showCancelConfirm = false;
  }

  $: showBack = currentStep !== "select-plotter" && currentStep !== "plotting" && currentStep !== "success";
  $: showNext = currentStep !== "success" && currentStep !== "plotting";
  $: showFinish = currentStep === "success";
  $: showCancel = currentStep !== "plotting" && currentStep !== "success";
  $: showCancelPlot = currentStep === "plotting";
  $: nextLabel = currentStep === "configure" ? "Plot Route" : "Next";
  $: canGoNext =
    currentStep === "select-plotter"
      ? selectedPlotterId !== ""
      : currentStep === "configure"
        ? fromSystem !== "" && toSystem !== ""
        : true;
</script>

<div class="wizard stack-lg">
  <div class="wizard-content">
    {#if currentStep === "select-plotter"}
      <div class="step-content stack-md">
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
      <div class="step-content stack-md">
        <h3>Step 2: Configure Route</h3>
        {#if plottingError}
          <p class="error">{plottingError}</p>
        {/if}
        <div class="input-grid stack-md">
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
      <div class="step-content stack-md">
        <h3>Success!</h3>
        {#if plottedRoute}
          <p>Route plotted successfully with {plottedRoute.jumps.length} jumps</p>
          <div class="result-values stack-sm">
            <p><strong>Route Name:</strong> {plottedRoute.name}</p>
            <p><strong>Jumps:</strong> {plottedRoute.jumps.length}</p>
            <p><strong>Plotter:</strong> {plotterOptions[selectedPlotterId]}</p>
          </div>
        {:else}
          <p>Route plotted successfully</p>
        {/if}
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
    {#if showCancelPlot}
      <Button variant="secondary" onClick={handleCancelPlot}>Cancel</Button>
    {/if}
    {#if showNext}
      <Button variant="primary" onClick={handleNext} disabled={!canGoNext}>{nextLabel}</Button>
    {/if}
    {#if showFinish}
      <Button variant="primary" onClick={handleNext}>Finish</Button>
    {/if}
  </div>
</div>

<ConfirmDialog
  bind:open={showCancelConfirm}
  title="Cancel Plotting"
  message="Are you sure you want to cancel plotting? This will abort the current operation."
  confirmLabel="Yes, Cancel"
  cancelLabel="No, Continue"
  confirmVariant="danger"
  onConfirm={confirmCancelPlot}
  onCancel={() => showCancelConfirm = false}
/>

<style>
  .wizard {
    min-width: 500px;
    padding: 1.5rem;
  }

  .wizard-content {
    min-height: 200px;
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

  .step-content label {
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
    margin-top: 1rem;
  }

  .result-values {
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
    justify-content: center;
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
