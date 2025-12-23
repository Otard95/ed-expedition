import type { models } from "../../../wailsjs/go/models";
import { ActiveJump } from "../routes/active";
import { formatDuration } from "../utils/dateFormat";

export interface ActiveExpeditionStats {
  allJumps: ActiveJump[];
  onRouteCount: number;
  detourCount: number;
  totalJumps: number;
  progressPercent: string;
  jumpsLeft: number;
  startDate: string;
  totalDistance: number;
  currentJumpIndex: number;
}

export function computeActiveStats(
  expedition: models.Expedition,
  bakedRoute: models.Route
): ActiveExpeditionStats {
  const allJumps = [
    ...expedition.jump_history,
    ...bakedRoute.jumps.slice(expedition.current_baked_index + 1),
  ].map((jump) => new ActiveJump(jump, bakedRoute.jumps));

  let onRouteCount = 0;
  let detourCount = 0;
  let totalDistance = 0;

  for (const jump of expedition.jump_history) {
    if (jump.baked_index !== undefined) {
      onRouteCount++;
    } else {
      detourCount++;
    }
    totalDistance += jump.distance || 0;
  }

  const totalJumps = expedition.jump_history.length

  // Use max(0, index) because expedition starts at -1 when location unknown at start time
  const effectiveIndex = Math.max(0, expedition.current_baked_index);
  const progressPercent = (
    (effectiveIndex / bakedRoute.jumps.length) *
    100
  ).toFixed(1);
  const jumpsLeft = bakedRoute.jumps.length - effectiveIndex;

  const startedOn = new Date(expedition.started_on);
  const startDate = startedOn.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });

  const currentJumpIndex = Math.max(expedition.jump_history.length - 1, 0);

  return {
    allJumps,
    onRouteCount,
    detourCount,
    totalJumps,
    progressPercent,
    jumpsLeft,
    startDate,
    totalDistance,
    currentJumpIndex,
  };
}
