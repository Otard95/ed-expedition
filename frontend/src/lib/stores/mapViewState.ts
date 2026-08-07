const STORAGE_KEY_PREFIX = "expedition_edit_view.map_state.";

export interface MapCameraState {
  targetX: number;
  targetY: number;
  targetZ: number;
  posX: number;
  posY: number;
  posZ: number;
}

export interface MapViewState {
  markers: { name: string; x: number; y: number; z: number }[];
  showStarField: boolean;
  showRegionLines: boolean;
  camera: MapCameraState | null;
}

const DEFAULTS: MapViewState = {
  markers: [],
  showStarField: true,
  showRegionLines: true,
  camera: null,
};

export function createMapViewStore(expeditionId: string) {
  const storageKey = `${STORAGE_KEY_PREFIX}${expeditionId}`;

  const load = (): MapViewState => {
    try {
      const stored = localStorage.getItem(storageKey);
      return stored ? { ...DEFAULTS, ...JSON.parse(stored) } : { ...DEFAULTS };
    } catch {
      return { ...DEFAULTS };
    }
  };

  const save = (state: MapViewState) => {
    try {
      localStorage.setItem(storageKey, JSON.stringify(state));
    } catch (err) {
      console.error("Failed to save map view state:", err);
    }
  };

  let state = load();

  return {
    get: () => state,
    save: (patch: Partial<MapViewState>) => {
      state = { ...state, ...patch };
      save(state);
    },
  };
}
