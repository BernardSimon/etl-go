import { request } from "../utils/request";
import type { LoginRequest, LoginResponse, VerifyTwoFactorRequest } from "../types";

/**
 * 登录接口
 */
export const loginApi = (data: LoginRequest) => {
  return request.post<LoginResponse>("/login", data);
};

/**
 * 两步验证接口
 */
export const verify2FAApi = (data: VerifyTwoFactorRequest) => {
  return request.post<LoginResponse>("/verify-2fa", data);
};

/**
 * 刷新 token 接口
 */
export const refreshTokenApi = (refreshToken: string) => {
  return request.post<LoginResponse>(
    "/refresh-token",
    {
      refresh_token: refreshToken,
    },
    {
      skipAuthRefresh: true,
    }
  );
};
