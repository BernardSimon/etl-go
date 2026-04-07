import { defineStore } from "pinia";
import { ref } from "vue";
import { loginApi, refreshTokenApi, verify2FAApi } from "../api/auth";
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
   * 登录（若开启 2FA 则返回挑战，不存 token）
   */
  const login = async (data: { username: string; password: string }) => {
    const res = await loginApi(data as any);
    if (res.data?.requires_2fa) {
      return res;
    }
    token.value = res.data?.token || "";
    refreshToken.value = res.data?.refresh_token || "";
    setToken(res.data?.token || "");
    setRefreshToken(res.data?.refresh_token || "");
    return res;
  };

  /**
   * 完成两步验证
   */
  const verifyTwoFactor = async (preAuthToken: string, code: string) => {
    const res = await verify2FAApi({ pre_auth_token: preAuthToken, code });
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
    verifyTwoFactor,
    refreshAccessToken,
    logout,
    resetUser,
    changeLanguage,
  };
});
