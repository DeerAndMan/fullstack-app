export enum RoleKey {
  SUPER_ADMIN = "super_admin",
  ADMIN = "admin",
  USER = "user",
}

export type Role = {
  id: number;
  name: string;
  role_key: RoleKey;
};
