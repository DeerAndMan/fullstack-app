import request from "./request";
import type { ApiResponse, TokenPair } from "@/types/api";

export function login(data: { username: string; password: string }) {
  return request.post<ApiResponse<TokenPair>>("/auth/login", data);
}

export function register(data: {
  username: string;
  password: string;
  nickname?: string;
}) {
  return request.post<ApiResponse>("/auth/register", data);
}

export function refreshToken(data: { refresh_token: string }) {
  return request.post<ApiResponse<TokenPair>>("/auth/refresh-token", data);
}
