<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import Button from "../../components/Button.svelte";
  import AutocompleteInput from "../../components/AutocompleteInput.svelte";
  import PlotterInput from "../../components/PlotterInput.svelte";
  import ConfirmDialog from "../../components/ConfirmDialog.svelte";
  import {
    GetPlotterOptions,
    GetPlotterInputConfig,
    PlotRoute,
    AutocompleteSystems,
    ValidateSystemName,
  } from "../../../wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";
  import type { plotters } from "../../../wailsjs/go/models";
  import ProgressBar from "../../components/ProgressBar.svelte";

  export let expeditionId: string;
  export let canClose: boolean = true;
  export let initialFrom: string | undefined = undefined;
  export let onComplete: (() => void) | undefined = undefined;

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

  let plottingError: string | null = null;
  let showCancelConfirm = false;
  let jobId: string | null = null;
  let jobProgress: number | null = null;
  let jobDeterminate: boolean = false;

  onDestroy(() => {
    if (jobId) EventsOff(`job:${jobId}`);
  });

  async function fetchSystemSuggestions(prefix: string): Promise<string[]> {
    return AutocompleteSystems(prefix);
  }

  async function validateSystem(
    name: string,
  ): Promise<{ valid: boolean; message?: string }> {
    const result = await ValidateSystemName(name);
    return {
      valid: result.valid,
      message: result.valid ? undefined : "Not found in database",
    };
  }

  $: canClose = currentStep !== "plotting" && currentStep !== "success";

  onMount(async () => {
    try {
      plotterOptions = await GetPlotterOptions();
      loadingPlotters = false;
    } catch (err) {
      plotterError =
        err instanceof Error ? err.message : "Failed to load plotters";
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
      jobProgress = null;
      jobDeterminate = false;

      try {
        jobId = await PlotRoute(
          expeditionId,
          selectedPlotterId,
          fromSystem,
          toSystem,
          inputValues,
        );

        EventsOn(`job:${jobId}`, (status: any) => {
          if (status.progress) {
            jobProgress = status.progress.fraction;
            jobDeterminate = status.progress.determinate;
          }

          if (status.status === "complete") {
            EventsOff(`job:${jobId}`);
            currentStep = "success";
          } else if (status.status === "error") {
            EventsOff(`job:${jobId}`);
            plottingError = status.error || "Plotting failed";
            currentStep = "configure";
          }
        });
      } catch (err) {
        plottingError =
          err instanceof Error
            ? err.message
            : typeof err === "string"
              ? err
              : String(err);
        console.error("Failed to plot route:", err);
        currentStep = "configure";
      }
    } else if (currentStep === "success" && onComplete) {
      onComplete();
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

  function handleCancelPlot() {
    showCancelConfirm = true;
  }

  function confirmCancelPlot() {
    currentStep = "configure";
    showCancelConfirm = false;
  }

  $: showBack =
    currentStep !== "select-plotter" &&
    currentStep !== "plotting" &&
    currentStep !== "success";
  $: showNext = currentStep !== "success" && currentStep !== "plotting";
  $: showFinish = currentStep === "success";
  $: showCancelPlot = currentStep === "plotting";
  $: nextLabel = currentStep === "configure" ? "Plot Route" : "Next";
  $: canGoNext =
    currentStep === "select-plotter"
      ? selectedPlotterId !== ""
      : currentStep === "configure"
        ? fromSystem !== "" && toSystem !== ""
        : true;
</script>

<div class="wizard flex-col flex-gap-lg">
  <div class="wizard-content">
    {#if currentStep === "select-plotter"}
      <div class="step-content flex-col flex-gap-md">
        <h3>Step 1: Select Plotter</h3>
        {#if loadingPlotters}
          <p class="loading text-secondary">Loading plotters...</p>
        {:else if plotterError}
          <p class="error text-danger">Error: {plotterError}</p>
        {:else}
          <p>Choose a plotter to generate your route:</p>
          <div class="plotter-options">
            {#each Object.entries(plotterOptions) as [id, name]}
              <label class="plotter-option flex-center">
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
          <p class="disclaimer text-dim">More plotters coming soon</p>
        {/if}
      </div>
    {:else if currentStep === "configure"}
      <div class="step-content flex-col flex-gap-md">
        <h3>Step 2: Configure Route</h3>
        {#if plottingError}
          <p class="error text-danger">{plottingError}</p>
        {/if}
        <div class="input-grid flex-col flex-gap-md">
          <AutocompleteInput
            bind:value={fromSystem}
            label="From System"
            placeholder="Sol"
            fetchSuggestions={fetchSystemSuggestions}
            validate={validateSystem}
            minChars={2}
            debounceMs={150}
          />
          <AutocompleteInput
            bind:value={toSystem}
            label="To System"
            placeholder="Colonia"
            fetchSuggestions={fetchSystemSuggestions}
            validate={validateSystem}
            minChars={2}
            debounceMs={150}
          />

          {#if plotterInputConfig}
            {#each plotterInputConfig as field}
              <PlotterInput {field} bind:value={inputValues[field.name]} />
            {/each}
          {/if}
        </div>
      </div>
    {:else if currentStep === "plotting"}
      <div class="step-content plotting flex-col flex-gap-md">
        <h3>Step 3: Plotting Route</h3>
        {#if jobDeterminate && jobProgress != null}
          <div class="plot-progress">
            <div class="plot-progress-labels">
              <span>{fromSystem}</span>
              <span class="text-secondary"
                >{(jobProgress * 100).toFixed(1)}%</span
              >
              <span>{toSystem}</span>
            </div>
            <div class="plot-progress-bar">
              <ProgressBar fraction={jobProgress} color="var(--ed-orange)" />
            </div>
          </div>
        {:else}
          <div class="loading-spinner"></div>
          <p>Plotting {fromSystem} → {toSystem}...</p>
        {/if}
      </div>
    {:else if currentStep === "success"}
      <div class="step-content flex-col flex-gap-md">
        <h3>Success!</h3>
        <p>Route plotted successfully</p>
        <div class="result-values flex-col flex-gap-sm">
          <p><strong>Route:</strong> {fromSystem} → {toSystem}</p>
          <p><strong>Plotter:</strong> {plotterOptions[selectedPlotterId]}</p>
        </div>
      </div>
    {/if}
  </div>

  <div class="wizard-actions">
    {#if showBack}
      <Button variant="secondary" onClick={handleBack}>Back</Button>
    {/if}
    <div class="spacer"></div>
    {#if showCancelPlot}
      <Button variant="secondary" onClick={handleCancelPlot}>Cancel</Button>
    {/if}
    {#if showNext}
      <Button variant="primary" onClick={handleNext} disabled={!canGoNext}
        >{nextLabel}</Button
      >
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
  onCancel={() => (showCancelConfirm = false)}
/>

<style>
  .wizard {
    width: 100%;
    padding: 1.5rem;
    box-sizing: border-box;
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

  .plot-progress {
    width: 100%;
  }

  .plot-progress-labels {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
    margin-bottom: 0.5rem;
    font-size: 0.85rem;
    color: var(--ed-text-primary);
  }

  .plot-progress-labels span:first-child {
    text-align: left;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .plot-progress-labels span:last-child {
    text-align: right;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .plot-progress-bar {
    position: relative;
    height: 4px;
    background: var(--ed-border);
    border-radius: 2px;
  }

  .step-content h3 {
    margin: 0;
    color: var(--ed-orange);
    font-size: 1.125rem;
  }

  .step-content p {
    margin: 0;
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
    font-style: italic;
  }

  .step-content .error {
    word-break: break-word;
    overflow-wrap: anywhere;
  }

  .step-content .disclaimer {
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
