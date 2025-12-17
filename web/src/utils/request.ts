import axios, {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  AxiosError,
} from "axios";

import { useUserStore } from "../stores/user";
import router from "../router";
import { message } from "ant-design-vue";

// 创建 axios 实例
const service: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL + "/etlApi",
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
    config.headers["Accept-Language"] = userStore.language;

    return config;
  },
  (error: AxiosError) => {
    console.error("request error：", error);
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
          message.error("登录已过期，请重新登录");
          router.push("/login");
        } else {
          // 统一显示后端返回的错误信息
          message.error(res.message || "请求失败");
        }

        return Promise.reject(new Error(res.message || "请求失败"));
      }
    }

    return res;
  },
  (error: AxiosError) => {
    console.error("响应错误：", error);
    const status = (error.response && error.response.status) || undefined;
    const code =
      (error.response && (error.response.data as any)?.code) || undefined;
    
    // 处理认证相关错误
    if (status === 401 || status === 4 || code === 3 || code === 4) {
      const userStore = useUserStore();
      userStore.resetUser();
      router.push("/login");
      message.error("登录已过期，请重新登录");
    } else {
      // 统一显示错误信息
      const msg =
        (error.response && (error.response.data as any)?.message) ||
        error.message ||
        "请求失败";


      console.error("请求错误：", error);
      message.error(msg);
    }

    return Promise.reject(error);
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
