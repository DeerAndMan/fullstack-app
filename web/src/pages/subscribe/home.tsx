import OperableDialog from "@/sections/subscribe/home/operable-dialog";
import HomeTable from "@/sections/subscribe/home/table";

export default function index() {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <OperableDialog />
      </div>
      <HomeTable />
    </div>
  );
}
