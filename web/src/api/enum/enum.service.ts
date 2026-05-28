import { request } from "../";
import { apiPathsV1 } from "../paths/v1";

import type { PromiseResponseData } from "../";
import type { Role } from "@/types/enum";

export const getRoleList = (): PromiseResponseData<Role[]> => request.get(apiPathsV1.role.all);
