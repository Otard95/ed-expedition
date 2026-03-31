<script lang="ts">
  import Modal from "../../components/Modal.svelte";
  import Button from "../../components/Button.svelte";

  export let open: boolean = false;
  export let onAccept: () => void;
  export let onDecline: () => void;
  export let onDismiss: () => void;

  function handleAccept() {
    open = false;
    onAccept();
  }

  function handleDecline() {
    open = false;
    onDecline();
  }

  function handleDismiss() {
    open = false;
    onDismiss();
  }
</script>

<Modal {open} title="Galaxy Database">
  <div class="content text-left">
    <p>
      ED Expedition can download a local copy of the galaxy database from
      <strong>Spansh</strong> to enable built-in route plotting and system search.
    </p>

    <div class="details">
      <h3>What this does</h3>
      <ul>
        <li>Downloads ~5 GB of compressed star system data from Spansh</li>
        <li>Processes and indexes ~174 million systems (~18 GB on disk)</li>
        <li>Stored on your machine — no data is sent anywhere</li>
      </ul>

      <h3>Things to know</h3>
      <ul>
        <li>
          This feature is <strong>experimental</strong> — it may not work as intended
        </li>
        <li>
          The initial download and build may take a while depending on you
          internet speed and hardware
        </li>
        <li>
          For now, you should <strong>not close the app</strong> while the build
          is in progress
        </li>
        <li>You can always change this later in settings</li>
      </ul>
    </div>

    <div class="actions">
      <Button variant="secondary" size="small" onClick={handleDecline}
        >Don't ask again</Button
      >
      <div class="actions-right">
        <Button variant="secondary" onClick={handleDismiss}>Not now</Button>
        <Button onClick={handleAccept}>Enable</Button>
      </div>
    </div>
  </div>
</Modal>

<style>
  .content {
    padding: 1.5rem;
    max-width: 480px;
  }

  .content p {
    margin: 0 0 1.25rem;
    color: var(--ed-text-primary);
    line-height: 1.5;
  }

  .content strong {
    color: var(--ed-orange);
  }

  .details {
    margin-bottom: 1.5rem;
  }

  .details h3 {
    margin: 0 0 0.5rem;
    font-size: 0.85rem;
    color: var(--ed-text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .details h3:not(:first-child) {
    margin-top: 1rem;
  }

  .details ul {
    margin: 0;
    padding-left: 1.25rem;
    color: var(--ed-text-secondary);
    line-height: 1.6;
  }

  .actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .actions-right {
    display: flex;
    gap: 0.75rem;
  }
</style>
