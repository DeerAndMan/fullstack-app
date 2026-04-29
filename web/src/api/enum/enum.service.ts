import { request, apiControl } from "../";

import type { PromiseResponseData } from "../";
import type { Role } from "@/types/enum";

export const getRoleList = (): PromiseResponseData<Role[]> => request.get(apiControl.role.all);
