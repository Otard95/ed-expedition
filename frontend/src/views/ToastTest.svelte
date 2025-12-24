<script lang="ts">
  import { push } from "svelte-spa-router";
  import { toasts, type ToastLevel } from "../lib/stores/toast";
  import Button from "../components/Button.svelte";

  let counter = 0;

  function addToast(level: ToastLevel, persistent: boolean, dismissable: boolean) {
    const id = `test-${counter++}`;
    toasts.set(id, {
      message: `This is a ${level} toast (${persistent ? 'persistent' : 'temporary'}, ${dismissable ? 'dismissable' : 'not dismissable'})`,
      level,
      persistent,
      dismissable,
    });
  }

  function addNamedToast() {
    toasts.set("named-toast", {
      message: `Named toast updated at ${new Date().toLocaleTimeString()}`,
      level: "info",
      persistent: false,
      dismissable: true,
    });
  }

  function addActionToast() {
    const id = `action-${counter++}`;
    toasts.set(id, {
      message: "This toast has an action",
      level: "warning",
      persistent: true,
      dismissable: false,
      action: {
        cta: "Fix it",
        callback: () => {
          toasts.set(id, {
            message: "Fixed!",
            level: "success",
            persistent: false,
            dismissable: true,
          });
        },
      },
    });
  }

  function addTitleToast(level: ToastLevel) {
    toasts.set(`title-${counter++}`, {
      title: "Fuel Warning",
      message: "You may not have enough fuel to reach the next scoopable star.",
      level,
      persistent: false,
      dismissable: true,
    });
  }

  function addAnimatedToast(level: ToastLevel) {
    toasts.set(`animated-${counter++}`, {
      title: "Alert",
      message: "This toast has an animated top bar.",
      level,
      persistent: true,
      dismissable: true,
      animate: true,
    });
  }

  function clearAll() {
    toasts.clear();
  }
</script>

<div class="flex-col flex-gap-lg">
  <h1>Toast Test</h1>

  <section class="flex-col flex-gap-md">
    <h2>Basic Toasts (dismissable)</h2>
    <div class="flex-gap-sm">
      <Button onClick={() => addToast("info", false, true)}>Info</Button>
      <Button onClick={() => addToast("success", false, true)}>Success</Button>
      <Button onClick={() => addToast("warning", false, true)}>Warning</Button>
      <Button variant="danger" onClick={() => addToast("danger", false, true)}>Danger</Button>
    </div>
  </section>

  <section class="flex-col flex-gap-md">
    <h2>Persistent + Not Dismissable</h2>
    <div class="flex-gap-sm">
      <Button onClick={() => addToast("warning", true, false)}>Warning (sticky)</Button>
      <Button variant="danger" onClick={() => addToast("danger", true, false)}>Danger (sticky)</Button>
    </div>
  </section>

  <section class="flex-col flex-gap-md">
    <h2>Named Toast (updates same toast)</h2>
    <div class="flex-gap-sm">
      <Button onClick={addNamedToast}>Update Named Toast</Button>
    </div>
  </section>

  <section class="flex-col flex-gap-md">
    <h2>Toast with Action</h2>
    <div class="flex-gap-sm">
      <Button onClick={addActionToast}>Add Action Toast</Button>
    </div>
  </section>

  <section class="flex-col flex-gap-md">
    <h2>Toast with Title</h2>
    <div class="flex-gap-sm">
      <Button onClick={() => addTitleToast("info")}>Info</Button>
      <Button onClick={() => addTitleToast("warning")}>Warning</Button>
      <Button variant="danger" onClick={() => addTitleToast("danger")}>Danger</Button>
    </div>
  </section>

  <section class="flex-col flex-gap-md">
    <h2>Animated Top Bar</h2>
    <div class="flex-gap-sm">
      <Button onClick={() => addAnimatedToast("info")}>Info</Button>
      <Button onClick={() => addAnimatedToast("warning")}>Warning</Button>
      <Button variant="danger" onClick={() => addAnimatedToast("danger")}>Danger</Button>
    </div>
  </section>

  <section class="flex-col flex-gap-md">
    <h2>Actions</h2>
    <div class="flex-gap-sm">
      <Button variant="secondary" onClick={clearAll}>Clear All</Button>
      <Button variant="secondary" onClick={() => push("/")}>Back to Index</Button>
    </div>
  </section>
</div>
