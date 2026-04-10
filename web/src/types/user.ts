import { RoleKey } from "./enum";

export interface AddAccount {
  age: number;
  description: string;
  email: string;
  name: string;
  password: string;
}

export interface Account {
  age: number;
  created_at: string;
  description: string;
  email: string;
  id: number;
  name: string;
  password: string;
  updated_at: string;
  addTime: string;
  avatar: string;
  roleKey: RoleKey;
  role: Role | null;
}

export interface Role {
  role_id: number; // 角色ID
  role_name: string; // 角色名称
  role_key: string; // 角色权限标识
  sort: number; // 角色显示顺序
  role_status: number; // 角色启用状态 (int8 映射为 number)
  create_by: string; // 创建人
  create_time: string; // 创建时间 (time.Time 映射为字符串)
  update_by: string; // 更新人
  update_time: string; // 更新时间 (time.Time 映射为字符串)
  remark?: string | null; // 角色备注 (指针类型映射为可选且可为空)
  del_flag: number; // 逻辑删除标志 (tinyint 映射为 number)
}

export interface MenuRole {
  id: number; // 主键ID
  menu_id: number; // 路由ID
  role_id: number; // 角色ID
  create_time: string; // 创建时间 (通常为 ISO 字符串格式)
  update_time: string; // 最后修改的时间戳 (通常为 ISO 字符串格式)
  create_user: number; // 创建用户
  update_user: number; // 操作用户
  is_delete: number; // 是否被删除 (0 或 1，对应 tinyint(1))
}
