export interface ApiResponse<T = unknown> {
  code: number;
  data: T;
  message: string;
}

export interface PageResult<T> {
  list: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface PageParams {
  page: number;
  page_size: number;
  keyword?: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export interface UserInfo {
  id: number;
  username: string;
  nickname: string;
  avatar: string;
  email: string;
  phone: string;
  status: number;
  roles: RoleInfo[];
  created_at: string;
}

export interface RoleInfo {
  id: number;
  name: string;
  code: string;
  remark: string;
  status: number;
  created_at: string;
}
