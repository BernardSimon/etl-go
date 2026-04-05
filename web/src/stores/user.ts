import { defineStore } from "pinia";
import { ref } from "vue";
import { loginApi, refreshTokenApi } from "../api/auth";
import {
  getToken,
  getRefreshToken,
  getLanguage,
  setToken,
  setRefreshToken,
  setLanguage,
  clearStorage,
} from "../utils/storage";

export const useUserStore = defineStore("user", () => {
  // 状态
  const token = ref<string>(getToken() || "");
  const refreshToken = ref<string>(getRefreshToken() || "");

  const language = ref<string>(getLanguage() || "");

  /**
   * 登录
   */
  const login = async (data: { username: string; password: string }) => {
    const res = await loginApi(data as any);
    // 设置store 与本地token
    token.value = res.data?.token || "";
    refreshToken.value = res.data?.refresh_token || "";
    setToken(res.data?.token || "");
    setRefreshToken(res.data?.refresh_token || "");
    return res;
  };

  /**
   * 刷新 access token
   */
  const refreshAccessToken = async () => {
    if (!refreshToken.value) {
      throw new Error("missing refresh token");
    }

    const res = await refreshTokenApi(refreshToken.value);
    token.value = res.data?.token || "";
    refreshToken.value = res.data?.refresh_token || "";
    setToken(token.value);
    setRefreshToken(refreshToken.value);
    return res;
  };

  /**
   * 切换语言
   */
  const changeLanguage = (lang: string) => {
    language.value = lang;
    setLanguage(lang);
  };

  /**
   * 登出 - 直接清除本地数据
   */
  const logout = () => {
    token.value = "";
    refreshToken.value = "";

    clearStorage();
  };

  /**
   * 重置用户信息
   */
  const resetUser = () => {
    token.value = "";
    refreshToken.value = "";

    clearStorage();
  };

  return {
    token,
    refreshToken,
    language,
    login,
    refreshAccessToken,
    logout,
    resetUser,
    changeLanguage,
  };
});
