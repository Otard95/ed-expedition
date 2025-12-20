<script lang="ts">
  import Router from 'svelte-spa-router'
  import ExpeditionIndex from './views/ExpeditionIndex.svelte'
  import ExpeditionEdit from './views/ExpeditionEdit.svelte'
  import ExpeditionActive from './views/ExpeditionActive.svelte'
  import ExpeditionView from './views/ExpeditionView.svelte'

  const routes = {
    '/': ExpeditionIndex,
    '/active': ExpeditionActive,
    '/expeditions/:id/view': ExpeditionView,
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
