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
	let currentJobId: string | null = null;
	let jobStatusCallback: ((status: any) => void) | null = null;

	GetGalaxyState().then((state) => set(state));

	return {
		subscribe,

		getJobId: () => currentJobId,

		markReady() {
			set(GalaxyStatus.READY);
		},

		onJobStatus(callback: (status: any) => void) {
			jobStatusCallback = callback;
		},

		async accept() {
			const jobId = await AcceptGalaxy();
			currentJobId = jobId;
			set(GalaxyStatus.IN_PROGRESS);
			if (jobId) {
				EventsOn(`job:${jobId}`, (status: any) => {
					jobStatusCallback?.(status);
				});
			}
		},

		async decline() {
			await DeclineGalaxy();
			set(GalaxyStatus.UNAVAILABLE);
		},

		async continue() {
			const jobId = await ContinueGalaxyBuild();
			currentJobId = jobId;
			set(GalaxyStatus.IN_PROGRESS);
			if (jobId) {
				EventsOn(`job:${jobId}`, (status: any) => {
					jobStatusCallback?.(status);
				});
			}
		},
	};
}

export const galaxy = createGalaxyStore();
