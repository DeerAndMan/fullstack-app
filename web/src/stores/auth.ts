import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

import { logout } from "@/api/user";
import { clearAuthCookie } from "@/utils/cookie";

import type { Account, Role } from "@/types/user";
import type { MenuItemType } from "@/types/menu-router";

interface AuthState {
  token: string;
  user: Account | null;
  role: Role | null;
  menuRoles: MenuItemType[] | null;
  loading: boolean;

  setToken: (token: string) => void;
  clearToken: () => void;
  setUser: (user: Account) => void;
  setUserAvatar: (avatar: string) => void;
  setRole: (role: Role) => void;
  setMenuRoles: (menuRoles: MenuItemType[]) => void;
  storeLogout: () => Promise<boolean>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      token: "",
      user: null,
      role: null,
      menuRoles: null,
      loading: false,

      setToken: (token: string) => set({ token }),

      clearToken: () => {
        clearAuthCookie();
        set({ token: "", user: null, role: null, menuRoles: null });
      },

      setUser: (user: Account) => set({ user }),

      setUserAvatar: (avatar: string) => {
        const user = get().user;
        if (!user) return;
        set({ user: { ...user, avatar } });
      },

      setRole: (role: Role) => set({ role }),

      setMenuRoles: (menuRoles: MenuItemType[]) => set({ menuRoles }),

      storeLogout: async () => {
        set({ loading: true });
        try {
          const data = await logout();
          if (data.code === 1) {
            set({ loading: false });
            return false;
          }
          clearAuthCookie();
          set({ token: "", user: null, role: null, menuRoles: null, loading: false });
          return true;
        } catch {
          set({ loading: false });
          return false;
        }
      },
    }),
    {
      name: "auth-storage",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        role: state.role,
        menuRoles: state.menuRoles,
      }),
    },
  ),
);
