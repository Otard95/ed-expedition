<script lang="ts">
  import Table from "../../components/Table.svelte";
  import { ActiveJump } from "../../lib/routes/active";
  import RouteActiveRow from "./RouteActiveRow.svelte";

  export let jumps: ActiveJump[];
  export let currentIndex: number = 0;
</script>

<Table
  columns={[
    { name: "", align: "center" },
    { name: "#", align: "left" },
    { name: "System", align: "left" },
    { name: "Scoopable", align: "center" },
    { name: "Neutron", align: "center" },
    { name: "Distance (LY)", align: "right" },
    {
      name: "Fuel",
      align: "right",
      tooltip:
        "Fuel in tank shown as Actual / Expected. Actual is fuel recorded in your jump history, Expected is fuel calculated by the route plotter.",
    },
  ]}
  data={jumps}
  let:item
  let:index
>
  <RouteActiveRow
    {index}
    jump={item}
    isCurrent={index === currentIndex}
    isNext={index === currentIndex + 1}
    isPrevOnRoute={index > 0 ? jumps[index - 1].on_route : false}
  />
</Table>
