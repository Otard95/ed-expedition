<script lang="ts">
  import Router from 'svelte-spa-router'
  import ExpeditionIndex from './views/ExpeditionIndex.svelte'
  import ExpeditionEdit from './views/ExpeditionEdit.svelte'
  import ExpeditionActive from './views/ExpeditionActive.svelte'
  import ExpeditionView from './views/ExpeditionView.svelte'
  import ToastTest from './views/ToastTest.svelte'
  import ToastContainer from './features/toasts/ToastContainer.svelte'
  import FuelAlertHandler from './features/fuel/FuelAlertHandler.svelte'

  const routes = {
    '/': ExpeditionIndex,
    '/active': ExpeditionActive,
    '/expeditions/:id/view': ExpeditionView,
    '/expeditions/:id': ExpeditionEdit,
    '/test/toasts': ToastTest,
  }
</script>

<main>
  <Router {routes} />
  <ToastContainer />
  <FuelAlertHandler />
</main>

<style>
  main {
    padding: 2rem;
  }
</style>
