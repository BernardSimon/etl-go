import { request } from "../utils/request";
import type { LoginRequest, LoginResponse } from "../types";

/**
 * 登录接口
 */
export const loginApi = (data: LoginRequest) => {
  return request.post<LoginResponse>("/login", data);
};
