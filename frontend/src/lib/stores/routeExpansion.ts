import { writable } from 'svelte/store';

interface ExpandRouteCommand {
  routeId: string;
  timestamp: number;
}

function createRouteExpansionStore() {
  const { subscribe, set } = writable<ExpandRouteCommand | null>(null);

  return {
    subscribe,
    expandRoute: (routeId: string) => {
      set({ routeId, timestamp: Date.now() });
    },
    clear: () => set(null)
  };
}

export const routeExpansion = createRouteExpansionStore();
