import { writable } from "svelte/store";
import {
	AcceptGalaxy,
	ContinueGalaxyBuild,
	DeclineGalaxy,
	GetGalaxyState,
} from "../../../wailsjs/go/main/App";
import { main } from "../../../wailsjs/go/models";
import { EventsOn } from "../../../wailsjs/runtime";

export type GalaxyStatus = main.GalaxyStatus;
export const GalaxyStatus = main.GalaxyStatus;

function createGalaxyStore() {
	const { subscribe, set } = writable<GalaxyStatus | null>(null);

	GetGalaxyState().then((state) => set(state));

	EventsOn("GalaxyBuildComplete", () => {
		set(GalaxyStatus.READY);
	});

	return {
		subscribe,

		async accept() {
			await AcceptGalaxy();
			set(GalaxyStatus.IN_PROGRESS);
		},

		async decline() {
			await DeclineGalaxy();
			set(GalaxyStatus.UNAVAILABLE);
		},

		continue() {
			ContinueGalaxyBuild();
			set(GalaxyStatus.IN_PROGRESS);
		},
	};
}

export const galaxy = createGalaxyStore();
