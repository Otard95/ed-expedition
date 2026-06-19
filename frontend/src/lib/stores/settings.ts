import { writable } from "svelte/store";
import { GetSettingsConfig } from "../../../wailsjs/go/main/App";

export interface AppSettings {
  journalDir: string;
  galaxyDecision: string;
  debug: boolean;
}

const defaults: AppSettings = {
  journalDir: "",
  galaxyDecision: "not_asked",
  debug: false,
};

function createSettingsStore() {
  const { subscribe, set } = writable<AppSettings>(defaults);

  async function load() {
    const fields = await GetSettingsConfig();
    const raw: Record<string, string> = {};
    for (const field of fields) {
      raw[field.name] = field.default;
    }
    set({
      journalDir: raw["journal_dir"] ?? "",
      galaxyDecision: raw["galaxy_decision"] ?? "not_asked",
      debug: raw["debug"] === "1",
    });
  }

  return {
    subscribe,
    load,
  };
}

export const settings = createSettingsStore();
