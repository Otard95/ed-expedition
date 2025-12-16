import type { models } from '../../../wailsjs/go/models';

interface EditViewRouteJumpLink {
  linkModel: models.Link,
  direction: 'in' | 'out',
  other: { id: string, i: number, label: string }
}
interface EditViewRouteJumpOptions {
  link?: EditViewRouteJumpLink,
  start: boolean,
  reachable: boolean,
}
export class EditViewRouteJump {
  public get system_name(): string { return this.jump.system_name; }
  public get system_id(): number { return this.jump.system_id; }
  public get scoopable(): boolean { return this.jump.scoopable; }
  public get must_refuel(): boolean { return this.jump.must_refuel; }
  public get distance(): number { return this.jump.distance; }
  public get fuel_in_tank(): number | undefined { return this.jump.fuel_in_tank; }
  public get fuel_used(): number | undefined { return this.jump.fuel_used; }
  public get overcharge(): boolean | undefined { return this.jump.overcharge; }
  public get position(): models.Position | undefined { return this.jump.position; }

  public get link(): EditViewRouteJumpLink | undefined { return this.options.link; }
  public set link(val: EditViewRouteJumpLink) { this.options.link = val; }

  public get start(): boolean { return this.options.start; }
  public set start(val: boolean) { this.options.start = val; }

  public get reachable(): boolean { return this.options.reachable; }
  public set reachable(val: boolean) { this.options.reachable = val; }

  constructor(
    private readonly jump: models.RouteJump,
    readonly options: EditViewRouteJumpOptions = { start: false, reachable: false }
  ) { }
}

export class EditViewRoute {
  public get id(): string { return this.route.id; }
  public get name(): string { return this.route.name; }
  public get plotter(): string { return this.route.plotter; }
  public get plotter_parameters(): Record<string, any> { return this.route.plotter_parameters; }
  public get plotter_metadata(): Record<string, any> | undefined { return this.route.plotter_metadata; }
  public get created_at(): any { return this.route.created_at; }

  public jumps: EditViewRouteJump[]

  constructor(
    private readonly route: models.Route,
    expedition_start: models.RoutePosition,
    links: models.Link[],
    routeIdToIdx: Record<string, number>,
  ) {
    const relevanl_links = links.filter(link =>
      link.from.route_id === route.id || link.to.route_id === route.id
    );

    this.jumps = route.jumps.map((jump, i) => {
      const start =
        expedition_start.route_id === route.id
        && expedition_start.jump_index === i;

      const link = EditViewRoute.linkForViewEditRouteJump(
        i,
        route.id,
        relevanl_links,
        routeIdToIdx,
      );

      return new EditViewRouteJump(jump, { link, start, reachable: false })
    });
  }

  private static linkForViewEditRouteJump(
    i: number,
    route_id: string,
    links: models.Link[],
    routeIdToIdx: Record<string, number>,
  ): EditViewRouteJumpOptions['link'] | undefined {
    const link = links.find(l =>
      (l.to.route_id === route_id && l.to.jump_index === i)
      || (l.from.route_id === route_id && l.from.jump_index === i)
    )
    if (!link) return;

    if (link.to.route_id === route_id) return {
      linkModel: link,
      direction: 'in',
      other: {
        i: link.from.jump_index,
        id: link.from.route_id,
        label: String(routeIdToIdx[link.from.route_id]+1)
      }
    }

    return {
      linkModel: link,
      direction: 'out',
      other: {
        i: link.to.jump_index,
        id: link.to.route_id,
        label: String(routeIdToIdx[link.to.route_id]+1)
      }
    }
  }
}

export function calculateReachable(
  start: models.RoutePosition,
  routes: EditViewRoute[]
) {
  const routeMap: Record<string, EditViewRoute> = {};
  routes.forEach(r => routeMap[r.id] = r);

  const visited = [];

  let next: models.RoutePosition | null | undefined = start;
  while (next) {
    const jump = routeMap[next.route_id].jumps[next.jump_index]
    if (visited.includes(jump)) break;
    jump.reachable = true;
    visited.push(jump);

    if (jump.link && jump.link.direction == 'out') {
      next = jump.link.linkModel.to;
    } else if (routeMap[next.route_id].jumps.length > next.jump_index+1) {
      next = { ...next, jump_index: next.jump_index+1 };
    } else {
      next = null;
    }
  }

  return routes
}

export function wouldCycle(
  newLink: models.Link,
  routes: EditViewRoute[]
) {
  const routeMap: Record<string, EditViewRoute> = {};
  routes.forEach(r => routeMap[r.id] = r);

  const visited = [];

  let next: models.RoutePosition | null | undefined = newLink.from;
  while (next) {
    const jump = routeMap[next.route_id].jumps[next.jump_index]
    if (visited.includes(jump)) return true;
    visited.push(jump);

    const isNewFrom = newLink.from.route_id === next.route_id && newLink.from.jump_index === next.jump_index
    if (jump.link && jump.link.direction == 'out' || isNewFrom) {
      next = jump.link ? jump.link.linkModel.to : newLink.to;
    } else if (routeMap[next.route_id].jumps.length > next.jump_index+1) {
      next = { ...next, jump_index: next.jump_index+1 };
    } else {
      next = null;
    }
  }

  return false
}

export class EditViewRoutePosition {
  public get route_id(): string { return this.position.route_id; }
  public get jump_index(): number { return this.position.jump_index; }

  constructor(
    private readonly position: models.RoutePosition,
    public readonly route_name: string,
    public readonly route_idx: number,
    public readonly system_name: string,
  ){}
}

export class EditViewLink {
  public get id(): string { return this.link.id; }
  public get from(): EditViewRoutePosition { return this.from_position; }
  public get to(): EditViewRoutePosition { return this.to_position; }

  private readonly from_position: EditViewRoutePosition
  private readonly to_position: EditViewRoutePosition

  constructor(private readonly link: models.Link, readonly routes: models.Route[]){
    const from = routes.findIndex(r => r.id === link.from.route_id)
    const to = routes.findIndex(r => r.id === link.to.route_id)

    this.from_position = new EditViewRoutePosition(
      link.from,
      routes[from].name,
      from,
      routes[from].jumps[link.from.jump_index].system_name
    )
    this.to_position = new EditViewRoutePosition(
      link.to,
      routes[to].name,
      to,
      routes[to].jumps[link.to.jump_index].system_name
    )
  }
}
