import { request } from "../utils/request";
import type { LoginRequest, LoginResponse } from "../types";

/**
 * 登录接口
 */
export const loginApi = (data: LoginRequest) => {
  return request.post<LoginResponse>("/login", data);
};

/**
 * 验证码登录接口
 */
export const loginWithCode = (data: {
  username: string;
  password: string;
  code: string;
}) => {
  return request.post<LoginResponse>("/loginWithCode", data);
};
