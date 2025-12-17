import { writable } from 'svelte/store';

const STORAGE_KEY_PREFIX = 'expedition_edit_view.collapse_state.';

function createRouteCollapseStore(expeditionId: string) {
  const storageKey = `${STORAGE_KEY_PREFIX}${expeditionId}`;

  // Load initial state from localStorage
  const loadState = (): Record<string, boolean> => {
    try {
      const stored = localStorage.getItem(storageKey);
      return stored ? JSON.parse(stored) : {};
    } catch {
      return {};
    }
  };

  const { subscribe, set, update } = writable<Record<string, boolean>>(loadState());

  // Save to localStorage whenever state changes
  const saveState = (state: Record<string, boolean>) => {
    try {
      localStorage.setItem(storageKey, JSON.stringify(state));
    } catch (err) {
      console.error('Failed to save collapse state:', err);
    }
  };

  return {
    subscribe,
    setCollapsed: (routeId: string, collapsed: boolean) => {
      update(state => {
        const newState = { ...state, [routeId]: collapsed };
        saveState(newState);
        return newState;
      });
    },
    getCollapsed: (routeId: string, defaultValue: boolean): boolean => {
      let result = defaultValue;
      subscribe(state => {
        result = state[routeId] ?? defaultValue;
      })();
      return result;
    },
    clear: () => {
      set({});
      try {
        localStorage.removeItem(storageKey);
      } catch (err) {
        console.error('Failed to clear collapse state:', err);
      }
    }
  };
}

export { createRouteCollapseStore };
