<script lang="ts">
  import Badge from './Badge.svelte'

  type ExpeditionStatus = 'planned' | 'active' | 'completed' | 'ended'

  export let status: ExpeditionStatus

  const statusColors: Record<ExpeditionStatus, string> = {
    planned: 'var(--ed-status-planned)',
    active: 'var(--ed-status-active)',
    completed: 'var(--ed-status-completed)',
    ended: 'var(--ed-status-ended)'
  }
</script>

<Badge variant="solid" color={statusColors[status]} active={status === 'active'}>
  {status}
</Badge>
