<script lang="ts">
  import { models } from "../../wailsjs/go/models";
  import ExpeditionStatusBadge from "../components/ExpeditionStatusBadge.svelte";
  import Card from "../components/Card.svelte";
  import Button from "../components/Button.svelte";
  import RouteEditTable from "../features/routes/RouteEditTable.svelte";
  import LinksSection from "../features/links/LinksSection.svelte";
  import { EditViewLink, EditViewRoute } from "../lib/routes/edit";
  import { routeExpansion } from "../lib/stores/routeExpansion";

  // Mock expedition data for now
  const expedition_start = { route_id: "route-1", jump_index: 0 };
  const expedition = new models.Expedition({
    id: "123e4567-e89b-12d3-a456-426614174000",
    name: "",
    created_at: new Date().toISOString(),
    last_updated: new Date().toISOString(),
    status: "planned",
    start: expedition_start,
    routes: [],
    links: [],
    jump_history: [],
    current_baked_index: 0,
  });

  const rawLinks = [
    new models.Link({
      id: "link-1",
      from: { route_id: "route-1", jump_index: 2 },
      to: { route_id: "route-2", jump_index: 3 },
    }),
    new models.Link({
      id: "link-2",
      from: { route_id: "route-2", jump_index: 7 },
      to: { route_id: "route-1", jump_index: 4 },
    }),
  ];

  const rawRoutes = [
    models.Route.createFrom({
      id: "route-1",
      name: "Sol → Colonia",
      plotter: "spansh",
      plotter_parameters: {},
      plotter_metadata: {},
      jumps: [
        {
          system_name: "Sol",
          system_id: 10477373803,
          scoopable: true,
          distance: 0,
          fuel_used: 0,
        },
        {
          system_name: "Alpha Centauri",
          system_id: 5031721931,
          scoopable: true,
          distance: 4.38,
          fuel_used: 0.42,
        },
        {
          system_name: "Colonia",
          system_id: 3238296097,
          scoopable: true,
          distance: 22000.5,
          fuel_used: 2.5,
        },
        {
          system_name: "Eol Prou VY-R d4-223",
          system_id: 2109876543,
          scoopable: true,
          distance: 54.2,
          fuel_used: 2.2,
        },
        {
          system_name: "Eol Prou PC-V d2-178",
          system_id: 1098765432,
          scoopable: false,
          distance: 46.8,
          fuel_used: 1.9,
        },
        {
          system_name: "Eol Prou MH-V e2-987",
          system_id: 9871234560,
          scoopable: true,
          distance: 55.1,
          fuel_used: 2.3,
        },
        {
          system_name: "Eol Prou ZX-T d3-456",
          system_id: 8761234509,
          scoopable: true,
          distance: 51.7,
          fuel_used: 2.1,
        },
      ],
      created_at: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000).toISOString(),
    }),
    models.Route.createFrom({
      id: "route-2",
      name: "Colonia → Eol Prou LW-L c8-306",
      plotter: "spansh",
      plotter_parameters: {},
      plotter_metadata: {},
      jumps: [
        {
          system_name: "Boewnst KS-S c17-890",
          system_id: 7651234098,
          scoopable: true,
          distance: 0,
          fuel_used: 0,
        },
        {
          system_name: "Boewnst AA-Z d13-567",
          system_id: 6541234087,
          scoopable: false,
          distance: 48.3,
          fuel_used: 1.9,
        },
        {
          system_name: "Eol Prou YZ-P d5-234",
          system_id: 5431234076,
          scoopable: true,
          distance: 52.6,
          fuel_used: 2.1,
        },
        {
          system_name: "Colonia",
          system_id: 3238296097,
          scoopable: true,
          distance: 49.8,
          fuel_used: 2.0,
        },
        {
          system_name: "Eol Prou RS-T d3-94",
          system_id: 9876543210,
          scoopable: true,
          distance: 52.3,
          fuel_used: 2.1,
        },
        {
          system_name: "Eol Prou IW-W e1-123",
          system_id: 8765432109,
          scoopable: false,
          distance: 45.2,
          fuel_used: 1.8,
        },
        {
          system_name: "Eol Prou KW-L c8-45",
          system_id: 7654321098,
          scoopable: true,
          distance: 53.7,
          fuel_used: 2.2,
        },
        {
          system_name: "Eol Prou NX-U d2-67",
          system_id: 6543210987,
          scoopable: true,
          distance: 49.1,
          fuel_used: 2.0,
        },
        {
          system_name: "Eol Prou QZ-Y c15-89",
          system_id: 5432109876,
          scoopable: false,
          distance: 51.4,
          fuel_used: 2.1,
        },
        {
          system_name: "Eol Prou VW-E d11-34",
          system_id: 4321098765,
          scoopable: true,
          distance: 47.8,
          fuel_used: 1.9,
        },
        {
          system_name: "Eol Prou YC-M d7-56",
          system_id: 3210987654,
          scoopable: false,
          distance: 50.3,
          fuel_used: 2.0,
        },
        {
          system_name: "Eol Prou LW-L c8-306",
          system_id: 1234567890,
          scoopable: false,
          distance: 48.7,
          fuel_used: 1.9,
        },
      ],
      created_at: new Date(Date.now() - 9 * 24 * 60 * 60 * 1000).toISOString(),
    }),
  ];

  const links = rawLinks.map((l) => new EditViewLink(l, rawRoutes));

  const routeIdToIdx = rawRoutes.reduce((acc, r, i) => {
    acc[r.id] = i;
    return acc;
  }, {});

  const routes = rawRoutes.map(
    (r) => new EditViewRoute(r, expedition_start, rawLinks, routeIdToIdx),
  );

  function scrollToJump(routeId: string, jumpIndex: number, event: MouseEvent) {
    // Signal the target route to expand if needed
    routeExpansion.expandRoute(routeId);

    // Wait a tick for the route to expand, then scroll
    setTimeout(() => {
      const element = document.getElementById(`jump-${routeId}-${jumpIndex}`);
      if (element) {
        // Get the parent row element
        const row = element.closest("tr");
        if (row) {
          row.scrollIntoView({ behavior: "smooth", block: "center" });
          // Add both highlight and blink to the target row
          row.classList.add("highlight", "blink");
          setTimeout(() => row.classList.remove("blink"), 1000);
          setTimeout(() => row.classList.remove("highlight"), 2000);
        }
      }
    }, 100);
  }

  let expeditionName = expedition.name;
</script>

<div class="expedition-edit">
  <div class="header">
    <div class="title-section">
      <input
        type="text"
        class="name-input"
        bind:value={expeditionName}
        placeholder="Unnamed Expedition"
      />
      <ExpeditionStatusBadge status={expedition.status} />
    </div>
  </div>

  <div class="sections">
    <div class="section">
      <div class="section-header">
        <h2>Routes</h2>
        <Button variant="primary" size="small">Add Route</Button>
      </div>
      {#if routes.length === 0}
        <Card>
          <p class="empty-message">
            No routes added yet. Add a route to begin planning your expedition.
          </p>
        </Card>
      {:else}
        <div class="routes-list">
          {#each routes as route, idx}
            <RouteEditTable {route} {idx} onGotoJump={scrollToJump} />
          {/each}
        </div>
      {/if}
    </div>

    <LinksSection {links} onGotoJump={scrollToJump} />
  </div>
</div>

<style>
  .expedition-edit {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .title-section {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex: 1;
  }

  .name-input {
    flex: 1;
    max-width: 600px;
    background: var(--ed-bg-secondary);
    border: 1px solid var(--ed-border);
    border-radius: 2px;
    padding: 0.5rem 0.75rem;
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--ed-text-primary);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .name-input:focus {
    outline: none;
    border-color: var(--ed-orange);
  }

  .name-input::placeholder {
    color: var(--ed-text-dim);
  }

  .sections {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .routes-list {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  :global(.jump-index) {
    color: var(--ed-text-dim);
    font-variant-numeric: tabular-nums;
  }

  :global(.numeric) {
    font-variant-numeric: tabular-nums;
  }

  :global(.scoopable) {
    font-size: 1.25rem;
    color: var(--ed-text-dim);
  }

  :global(.scoopable.yes) {
    color: var(--ed-orange);
  }

  :global(.links-cell) {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  :global(tr.highlight) {
    background: rgba(255, 120, 0, 0.3) !important;
    transition: background-color 0.3s ease;
  }

  @keyframes blink {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.3;
    }
  }

  :global(tr.blink) {
    animation: blink 0.5s ease-in-out 2;
  }
</style>
