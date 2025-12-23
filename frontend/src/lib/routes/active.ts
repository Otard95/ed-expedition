import type { models } from '../../../wailsjs/go/models';

export class ActiveJump {
  // Common fields
  public get system_name(): string { return this.jump.system_name; }
  public get system_id(): number { return this.jump.system_id; }
  public get distance(): number { return this.jump.distance; };

  // Actual fuel: only from history entries
  public get fuel_in_tank(): number | undefined {
    if (ActiveJump.IsHistory(this.jump)) return this.jump.fuel_in_tank;
    return undefined;
  }

  public get fuel_used(): number | undefined { return this.jump.fuel_used; }

  // Expected fuel: from baked route (history) or from the RouteJump itself (future)
  public get expected_fuel(): number | undefined {
    return this.bakedJump?.fuel_in_tank ?? (ActiveJump.IsHistory(this.jump) ? undefined : this.jump.fuel_in_tank);
  }

  // RouteJump fields
  public get has_neutron(): boolean | undefined {
    if (!ActiveJump.IsHistory(this.jump)) return this.jump.has_neutron;
    if (this.bakedJump) return this.bakedJump.has_neutron;
  }
  public get scoopable(): boolean | undefined {
    if (!ActiveJump.IsHistory(this.jump)) return this.jump.scoopable;
    return this.bakedJump && this.bakedJump.scoopable;
  }
  public get must_refuel(): boolean | undefined {
    if (!ActiveJump.IsHistory(this.jump)) return this.jump.must_refuel;
    return this.bakedJump && this.bakedJump.must_refuel;
  }

  // History only - defaults if base is RouteJump
  public get expected(): boolean {
    if (ActiveJump.IsHistory(this.jump)) return this.jump.expected;
    return true;
  };
  public get synthetic(): boolean {
    if (ActiveJump.IsHistory(this.jump)) return this.jump.synthetic;
    return false;
  };

  public get on_route(): boolean { return this.bakedJump !== undefined }

  private bakedJump?: models.RouteJump

  constructor(
    private readonly jump: models.JumpHistoryEntry | models.RouteJump,
    bakedJumps: models.RouteJump[],
  ) {
    if (ActiveJump.IsHistory(jump) && typeof jump.baked_index === 'number') {
      this.bakedJump = bakedJumps[jump.baked_index]
    }
  }

  private static IsHistory(
    jump: models.JumpHistoryEntry | models.RouteJump
  ): jump is models.JumpHistoryEntry {
    return 'timestamp' in jump;
  }
}
