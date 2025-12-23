<script lang="ts">
  import { push } from "svelte-spa-router";
  import { models } from "../../../wailsjs/go/models";
  import { calculateCompletionStats } from "../../lib/expedition/stats";
  import Modal from "../../components/Modal.svelte";
  import Button from "../../components/Button.svelte";

  export let open: boolean;
  export let expedition: models.Expedition | null;

  $: stats = open && expedition ? calculateCompletionStats(expedition) : null;
</script>

<Modal {open} title="Expedition Complete!" showCloseButton={false}>
  <div class="completion-content flex-col flex-gap-md">
    <div class="celebration-text flex-col flex-gap-md">
      <p class="hype">🎉 Outstanding work, Commander! 🎉</p>
      <p class="text-secondary">
        You've successfully completed your expedition! Your flight data has been
        logged and archived in the expedition database.
      </p>
      <p class="expedition-name">
        {expedition?.name || "Unnamed Expedition"}
      </p>
    </div>

    {#if stats}
      <div class="completion-stats-container flex-col flex-gap-md">
        <div class="stats-group">
          <div class="stats-group-title text-uppercase-tracked">Time</div>
          <div class="stats-group-content">
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Started
              </div>
              <div class="completion-stat-value small">
                {stats.startDate}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Ended
              </div>
              <div class="completion-stat-value small">
                {stats.endDate}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Duration
              </div>
              <div class="completion-stat-value">
                {stats.duration}
              </div>
            </div>
          </div>
        </div>

        <div class="stats-group">
          <div class="stats-group-title text-uppercase-tracked">Jumps</div>
          <div class="stats-group-content">
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Total
              </div>
              <div class="completion-stat-value">
                {stats.totalJumps}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                On Route
              </div>
              <div class="completion-stat-value">
                {stats.onRouteJumps}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Detours
              </div>
              <div class="completion-stat-value">
                {stats.detourJumps}
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Accuracy
              </div>
              <div class="completion-stat-value">
                {stats.routeAccuracy.toFixed(1)}%
              </div>
            </div>
          </div>
        </div>

        <div class="stats-group">
          <div class="stats-group-title text-uppercase-tracked">Distance</div>
          <div class="stats-group-content">
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Total
              </div>
              <div class="completion-stat-value">
                {stats.totalDistance.toFixed(2)} LY
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Average
              </div>
              <div class="completion-stat-value">
                {stats.averageJump.toFixed(2)} LY
              </div>
            </div>
            <div class="completion-stat">
              <div class="completion-stat-label text-uppercase-tracked">
                Longest
              </div>
              <div class="completion-stat-value">
                {stats.longestJump.toFixed(2)} LY
              </div>
            </div>
          </div>
        </div>
      </div>
    {/if}

    <div class="action-buttons flex-row">
      <Button
        variant="primary"
        onClick={() => {
          if (expedition) {
            push(`/expeditions/${expedition.id}/view`);
          }
        }}
      >
        View Expedition
      </Button>
      <Button variant="secondary" onClick={() => push("/")}>
        Back to Index
      </Button>
    </div>
  </div>
</Modal>

<style>
  .completion-content {
    max-width: 700px;
    text-align: center;
    padding: 1.5rem;
  }

  .hype {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--ed-orange);
    margin: 0;
  }

  .celebration-text p {
    margin: 0;
    line-height: 1.6;
  }

  .expedition-name {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--ed-text-primary);
    margin-top: 0.5rem;
  }

  .action-buttons {
    margin-top: 2rem;
    display: flex;
    gap: 1rem;
    justify-content: center;
  }

  .stats-group {
    background: var(--ed-bg-tertiary);
    border-radius: 4px;
    border: 1px solid var(--ed-border);
    overflow: hidden;
  }

  .stats-group-title {
    padding: 0.5rem 1rem;
    background: hsl(from var(--ed-bg-tertiary) h s calc(l * 0.9));
    border-bottom: 1px solid var(--ed-border);
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--ed-orange);
  }

  .stats-group-content {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
    gap: 1.5rem;
    padding: 1.5rem;
  }

  .completion-stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
  }

  .completion-stat-label {
    color: hsl(from var(--ed-orange) h s calc(l * 0.7));
    font-size: 0.75rem;
  }

  .completion-stat-value {
    color: var(--ed-text-primary);
    font-size: 1.5rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }

  .completion-stat-value.small {
    font-size: 1rem;
  }
</style>
