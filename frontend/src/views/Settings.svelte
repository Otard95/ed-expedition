<script lang="ts">
  import { onMount } from "svelte";
  import Button from "../components/Button.svelte";
  import ButtonLink from "../components/ButtonLink.svelte";
  import FormField from "../components/FormField.svelte";
  import Arrow from "../components/icons/Arrow.svelte";
  import { GetSettingsConfig, UpdateSetting } from "../../wailsjs/go/main/App";
  import type { form } from "../../wailsjs/go/models";
  import { toasts } from "../lib/stores/toast";
  import { settings } from "../lib/stores/settings";

  let fields: form.InputFieldConfig[] = [];
  let values: Record<string, string> = {};
  let savedValues: Record<string, string> = {};
  let loading = true;
  let saving = false;
  let error: string | null = null;
  let formKey = 0;

  $: dirty = Object.keys(values).some((k) => values[k] !== savedValues[k]);

  onMount(async () => {
    await loadSettings();
  });

  async function loadSettings() {
    try {
      loading = true;
      error = null;
      fields = await GetSettingsConfig();
      values = {};
      for (const field of fields) {
        values[field.name] = field.default;
      }
      savedValues = { ...values };
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    } finally {
      loading = false;
    }
  }

  function handleReset() {
    values = { ...savedValues };
    formKey++;
  }

  async function handleSave() {
    saving = true;
    const changed = Object.keys(values).filter((k) => values[k] !== savedValues[k]);
    try {
      for (const key of changed) {
        await UpdateSetting(key, values[key]);
      }
      savedValues = { ...values };
      await settings.load();
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err);
      toasts.set("settings-save-error", {
        message: `Failed to save settings: ${msg}`,
        level: "danger",
        persistent: false,
        dismissable: true,
      });
      await loadSettings();
    } finally {
      saving = false;
    }
  }

  $: sections = groupBySection(fields);

  function groupBySection(
    fields: form.InputFieldConfig[],
  ): { name: string; fields: form.InputFieldConfig[] }[] {
    const map = new Map<string, form.InputFieldConfig[]>();
    for (const field of fields) {
      const section = field.section || "";
      if (!map.has(section)) {
        map.set(section, []);
      }
      map.get(section)!.push(field);
    }
    return Array.from(map.entries()).map(([name, fields]) => ({
      name,
      fields,
    }));
  }
</script>

<div class="settings flex-col flex-gap-lg">
  <div class="header flex-between">
    <div class="title-group">
      <ButtonLink href="#/" variant="secondary" size="small">
        <Arrow direction="left" size="0.75rem" /> Back
      </ButtonLink>
      <h1 class="text-uppercase-tracked">Settings</h1>
    </div>
  </div>

  {#if loading}
    <p class="text-secondary">Loading settings...</p>
  {:else if error}
    <p class="text-danger">Error: {error}</p>
  {:else}
    <div class="settings-form flex-col flex-gap-lg">
      {#each sections as section}
        <div class="section flex-col flex-gap-md">
          {#if section.name}
            <h2 class="section-title">{section.name}</h2>
          {/if}
          {#each section.fields as field (field.name)}
            {#key formKey}
              <FormField {field} bind:value={values[field.name]} showInfoInline />
            {/key}
          {/each}
        </div>
      {/each}

      {#if dirty}
        <div class="save-actions">
          <Button variant="secondary" onClick={handleReset} disabled={saving}>
            Reset
          </Button>
          <Button variant="primary" onClick={handleSave} disabled={saving}>
            {saving ? "Saving..." : "Save"}
          </Button>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .title-group {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  h1 {
    margin: 0;
    font-size: 2rem;
    font-weight: 600;
    color: var(--ed-orange);
  }

  .settings-form {
    max-width: 600px;
  }

  .save-actions {
    display: flex;
    gap: 0.5rem;
    justify-content: flex-end;
  }

  .section-title {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--ed-orange-dim);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    padding-bottom: 0.375rem;
    border-bottom: 1px solid var(--ed-border);
  }
</style>
