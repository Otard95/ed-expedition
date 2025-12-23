import { models } from "../../../wailsjs/go/models";
import { formatDuration } from "../utils/dateFormat";

export interface CompletionStats {
  totalJumps: number;
  onRouteJumps: number;
  detourJumps: number;
  totalDistance: number;
  longestJump: number;
  averageJump: number;
  routeAccuracy: number;
  duration: string;
  startDate: string;
  endDate: string;
}

export function calculateCompletionStats(exp: models.Expedition): CompletionStats {
  const totalJumps = exp.jump_history.length;
  let onRouteJumps = 0;
  let totalDistance = 0;
  let longestJump = 0;

  for (const jump of exp.jump_history) {
    if (jump.baked_index !== undefined) onRouteJumps++;
    const distance = jump.distance || 0;
    totalDistance += distance;
    if (distance > longestJump) longestJump = distance;
  }

  const detourJumps = totalJumps - onRouteJumps;
  const averageJump = totalJumps > 0 ? totalDistance / totalJumps : 0;
  const routeAccuracy = totalJumps > 0 ? (onRouteJumps / totalJumps) * 100 : 0;

  let duration = "Unknown";
  let startDate = "Unknown";
  let endDate = "Unknown";

  if (exp.jump_history.length > 0) {
    const firstJump = new Date(exp.jump_history[0].timestamp);
    const lastJump = new Date(exp.jump_history[exp.jump_history.length - 1].timestamp);
    const durationMs = lastJump.getTime() - firstJump.getTime();

    duration = formatDuration(durationMs);

    startDate = firstJump.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
    endDate = lastJump.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  return {
    totalJumps,
    onRouteJumps,
    detourJumps,
    totalDistance,
    longestJump,
    averageJump,
    routeAccuracy,
    duration,
    startDate,
    endDate,
  };
}
