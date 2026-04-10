import { create } from "zustand";

import { getRoleList } from "@/api/enum";

import type { Role } from "@/types/enum";

interface EnumState {
  roleList: Role[];
  setRoleList: (roles: Role[]) => void;
  fetchRoles: () => Promise<void>;
}

export const useEnumStore = create<EnumState>()((set) => ({
  roleList: [],

  setRoleList: (roles: Role[]) => set({ roleList: roles }),

  fetchRoles: async () => {
    try {
      const data = await getRoleList();
      if (data.code !== 0) {
        set({ roleList: [] });
        return;
      }
      set({ roleList: data.data });
    } catch {
      set({ roleList: [] });
    }
  },
}));
