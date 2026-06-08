import { models } from "../../../wailsjs/go/models";

export function injectionLevel(boost: models.FSDBoost): 1 | 2 | 3 | null {
  switch (boost) {
    case models.FSDBoost.INJECTION_BASIC:    return 1;
    case models.FSDBoost.INJECTION_STANDARD: return 2;
    case models.FSDBoost.INJECTION_PREMIUM:  return 3;
    default:                                 return null;
  }
}
