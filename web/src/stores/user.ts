import { defineStore } from "pinia";
import { ref } from "vue";
import { loginApi } from "../api/auth";
import {
  getToken,
  setToken,
  clearStorage,
} from "../utils/storage";


export const useUserStore = defineStore("user", () => {
  // 状态
  const token = ref<string>(getToken() || "");

  const language = ref<string>("");

  /**
   * 登录
   */
  const login = async (data: { username: string; password: string }) => {
    const res = await loginApi(data as any);
    // 设置store 与本地token
    token.value = res.data?.token;
    setToken(res.data?.token || "");
    return res;
  };

    /**
     * 切换语言
     */
    const changeLanguage = (lang: string) => {
        language.value = lang;
    }

  /**
   * 登出 - 直接清除本地数据
   */
  const logout = () => {
    token.value = "";

    clearStorage();
  };

  /**
   * 重置用户信息
   */
  const resetUser = () => {
    token.value = "";

    clearStorage();
  };

  return {
    token,
      language,
    login,
    logout,
    resetUser,
      changeLanguage,
  };
});
