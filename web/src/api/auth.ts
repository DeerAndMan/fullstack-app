import request from "./request";
import type { ApiResponse, TokenPair } from "@/types/api";

export function login(data: { username: string; password: string }) {
  return request.post<ApiResponse<TokenPair>>("/api/v1/auth/login", data);
}

export function register(data: {
  username: string;
  password: string;
  nickname?: string;
}) {
  return request.post<ApiResponse>("/api/v1/auth/register", data);
}

export function refreshToken(data: { refresh_token: string }) {
  return request.post<ApiResponse<TokenPair>>("/api/v1/auth/refresh-token", data);
}
