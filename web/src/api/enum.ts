import request from "./request";
import { apiControl } from "@/api";

import type { PromiseResponseData } from "@/api";
import type { Role } from "@/types/enum";

export const getRoleList = (): PromiseResponseData<Role[]> => request.get(apiControl.enum.role);
