<script lang="ts">
  import { models } from "../../../wailsjs/go/models";
  import Card from "../../components/Card.svelte";
  import Tooltip from "../../components/Tooltip.svelte";
  import TimeSince from "../../components/TimeSince.svelte";
  import type { ActiveExpeditionStats as IActiveExpeditionStats } from "../../lib/expedition/active";

  export let expedition: models.Expedition | null = null;
  export let stats: IActiveExpeditionStats | null = null;
  export let compact: boolean = false;
</script>

{#if expedition && stats}
  <Card>
    <div class="stats" class:compact>
      <div class="stat">
        <div class="stat-label-small text-uppercase-tracked">Progress</div>
        <div class="stat-value-compact">
          {stats.progressPercent}%
        </div>
      </div>
      <div class="stat">
        <div class="stat-label-small text-uppercase-tracked">Jumps Left</div>
        <div class="stat-value-compact">
          {stats.jumpsLeft}
        </div>
      </div>
      <div class="stat">
        <div class="stat-label-small text-uppercase-tracked">
          Jumps <Tooltip
            text="On Route / Detour / Total"
            direction="down"
            nowrap
            size="0.75rem"
          />
        </div>
        <div class="stat-value-compact">
          {stats.onRouteCount} <span class="slash text-dim">/</span>
          {stats.detourCount} <span class="slash text-dim">/</span>
          {stats.totalJumps}
        </div>
      </div>
      {#if stats.startDate}
        <div class="stat">
          <div class="stat-label-small text-uppercase-tracked">Started</div>
          <div class="stat-value-compact small">{stats.startDate}</div>
        </div>
      {/if}
      {#if expedition}
        <div class="stat">
          <div class="stat-label-small text-uppercase-tracked">Duration</div>
          <div class="stat-value-compact">
            <TimeSince since={expedition.started_on} interval={10000} />
          </div>
        </div>
      {/if}
      <div class="stat">
        <div class="stat-label-small text-uppercase-tracked">Distance</div>
        <div class="stat-value-compact">
          {stats.totalDistance.toFixed(1)} LY
        </div>
      </div>
    </div>
  </Card>
{/if}

<style>
  .stats {
    padding: 1rem 1.5rem;
    display: flex;
    align-items: stretch;
    gap: 3rem;
    justify-content: center;
    transition: all 0.2s ease;
  }

  .stats.compact {
    padding: 0 1rem;
    gap: 5rem;
  }

  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: space-between;
    gap: 0.25rem;
    transition: gap 0.2s ease;
  }

  .compact .stat {
    gap: 0.15rem;
  }

  .stat-label-small {
    color: hsl(from var(--ed-orange) h s calc(l * 0.7));
    font-size: 0.75rem;
    transition: font-size 0.2s ease;
  }

  .stat-value-compact {
    color: var(--ed-text-primary);
    font-size: 1.5rem;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
    transition: font-size 0.2s ease;
  }

  .stat-value-compact.small {
    font-size: 1rem;
  }

  .compact .stat-value-compact {
    font-size: 1.1rem;
  }

  .compact .stat-value-compact.small {
    font-size: 0.85rem;
  }

  .compact .stat-label-small {
    font-size: 0.65rem;
  }

  .slash {
    font-weight: 400;
  }
</style>
