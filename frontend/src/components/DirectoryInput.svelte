<script lang="ts">
  import TextInput from "./TextInput.svelte";
  import Button from "./Button.svelte";
  import { BrowseDirectory } from "../../wailsjs/go/main/App";

  export let value: string = "";
  export let label: string = "";
  export let info: string | undefined = undefined;

  let className: string = "";
  export { className as class };

  async function handleBrowse() {
    const selected = await BrowseDirectory(label || "Select directory");
    if (selected) {
      value = selected;
    }
  }
</script>

<div class="directory-input {className}">
  <div class="input-row">
    <div class="input-field">
      <TextInput bind:value {label} {info} />
    </div>
    <Button variant="secondary" onClick={handleBrowse}>Browse</Button>
  </div>
</div>

<style>
  .input-row {
    display: flex;
    gap: 0.5rem;
    align-items: flex-end;
  }

  .input-row :global(.btn) {
    font-size: 1rem;
    padding: 0.5rem 0.75rem;
  }

  .input-field {
    flex: 1;
  }
</style>
