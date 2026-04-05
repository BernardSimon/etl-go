import axios, {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  AxiosError,
} from "axios";

import { useUserStore } from "../stores/user";
import router from "../router";
import { message } from "ant-design-vue";
import i18n from "../i18n";
import type { ApiErrorData } from "../types";

export class ApiRequestError extends Error {
  code?: number;
  details?: ApiErrorData;

  constructor(message: string, code?: number, details?: ApiErrorData) {
    super(message);
    this.name = "ApiRequestError";
    this.code = code;
    this.details = details;
  }
}

// 创建 axios 实例
const service: AxiosInstance = axios.create({
  baseURL: `${String(import.meta.env.VITE_API_BASE_URL || "").replace(/\/$/, "")}/api/v1`,
  timeout: 30000,
  headers: {
    "Content-Type": "application/json",
  },
});

// 请求拦截器
service.interceptors.request.use(
  (config) => {
    // 从 store 中获取 token
    const userStore = useUserStore();
    const token = userStore.token;

    // 如果存在 token，添加到请求头
    if (token) {
      config.headers.Authorization = `${token}`;
      // config.headers.Token = `${token}`;
    }
    if (config.data instanceof FormData && config.headers) {
      delete (config.headers as Record<string, unknown>)["Content-Type"];
    }
    config.headers["Accept-Language"] = userStore.language;

    return config;
  },
  (error: AxiosError) => {
    console.error("Request error:", error);
    return Promise.reject(error);
  }
);

// 响应拦截器
service.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data;

    // 根据业务需求判断响应状态
    if (res && typeof res === "object" && "code" in res) {
      const success = res.code === 0 || res.code === 200;
      if (!success) {
        // 处理认证过期，跳转到登录页
        if (res.code === 3 || res.code === 4) {
          const userStore = useUserStore();
          userStore.resetUser();
          message.error(i18n.global.t("request.authExpired"));
          router.push("/login");
        } else {
          // 统一显示后端返回的错误信息
          message.error(res.message || i18n.global.t("request.failed"));
        }

        return Promise.reject(
          new ApiRequestError(
            res.message || i18n.global.t("request.failed"),
            res.code,
            res.data
          )
        );
      }
    }

    return res;
  },
  (error: AxiosError) => {
    console.error("Response error:", error);
    const status = (error.response && error.response.status) || undefined;
    const code =
      (error.response && (error.response.data as any)?.code) || undefined;
    
    // 处理认证相关错误
    if (status === 401 || status === 4 || code === 3 || code === 4) {
      const userStore = useUserStore();
      userStore.resetUser();
      router.push({
        path: "/login",
        query: { redirect: router.currentRoute.value.fullPath },
      });
      message.error(i18n.global.t("request.authExpired"));
    } else {
      // 统一显示错误信息
      const isTimeout = error.code === "ECONNABORTED" || error.message?.toLowerCase().includes("timeout");
      const isNetworkError = !error.response;
      const msg = isTimeout
        ? i18n.global.t("request.timeout")
        : isNetworkError
          ? i18n.global.t("request.networkError")
          : (error.response && (error.response.data as any)?.message) ||
            error.message ||
            i18n.global.t("request.failed");


      console.error("Request error:", error);
      message.error(msg);
    }

    return Promise.reject(
      new ApiRequestError(
        String(
          (error.response && (error.response.data as any)?.message) ||
          error.message ||
          i18n.global.t("request.failed")
        ),
        code,
        (error.response && (error.response.data as any)?.data) || undefined
      )
    );
  }
);

// 导出请求方法
export default service;

// 封装常用的请求方法
export const request = {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return service.get(url, config);
  },

  post<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<T> {
    return service.post(url, data, config);
  },

  put<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<T> {
    return service.put(url, data, config);
  },

  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T> {
    return service.delete(url, config);
  },
};
