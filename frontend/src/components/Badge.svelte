<script lang="ts">
  type ExpeditionStatus = 'planned' | 'active' | 'completed' | 'ended'

  export let status: ExpeditionStatus

  const colors: Record<ExpeditionStatus, string> = {
    planned: 'var(--ed-status-planned)',
    active: 'var(--ed-status-active)',
    completed: 'var(--ed-status-completed)',
    ended: 'var(--ed-status-ended)'
  }
</script>

<span
  class="badge"
  class:active={status === 'active'}
  style="background-color: {colors[status]}"
>
  {status.toUpperCase()}
</span>

<style>
  .badge {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    border-radius: 2px;
    font-size: 0.75rem;
    font-weight: 600;
    color: #000;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .badge.active {
    box-shadow: 0 0 8px rgba(255, 120, 0, 0.5);
  }
</style>
