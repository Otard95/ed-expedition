<script lang="ts">
  import Router from 'svelte-spa-router'
  import ExpeditionIndex from './views/ExpeditionIndex.svelte'
  import ExpeditionEdit from './views/ExpeditionEdit.svelte'

  const routes = {
    '/': ExpeditionIndex,
    '/expeditions/:id': ExpeditionEdit,
  }
</script>

<main>
  <Router {routes} />
</main>

<style>
  main {
    padding: 2rem;
  }
</style>
