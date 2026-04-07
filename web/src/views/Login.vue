<template>
  <div class="login-container">

    <div class="login-box">
      <a-radio-group class="language-select" v-model:value="userStore.language" @change="changeLanguage" button-style="solid" size="small">
        <a-radio-button value="zh">{{ $t('layout.language.zh') }}</a-radio-button>
        <a-radio-button value="en">{{ $t('layout.language.en') }}</a-radio-button>
      </a-radio-group>
      <div class="login-header">
        <h2>{{ step === 'totp' ? $t('login.totp.title') : $t('login.title') }}</h2>
      </div>

      <!-- 第一步：用户名 + 密码 -->
      <a-form
        v-if="step === 'credentials'"
        ref="loginFormRef"
        :model="loginForm"
        :rules="getLoginRules()"
        class="login-form"
        @finish="handleLogin"
      >
          <a-form-item name="username">
            <a-input
              v-model:value="loginForm.username"
              :placeholder="$t('login.account.placeholder')"
              size="large"
              allow-clear
            >
              <template #prefix>
                <UserOutlined />
              </template>
            </a-input>
          </a-form-item>

          <a-form-item name="password">
            <a-input-password
              v-model:value="loginForm.password"
              :placeholder="$t('login.password.placeholder')"
              size="large"
              allow-clear
              @pressEnter="handleLogin"
            >
              <template #prefix>
                <LockOutlined />
              </template>
            </a-input-password>
          </a-form-item>

        <a-form-item>
          <a-alert
            v-if="loginError"
            :message="loginError"
            type="error"
            show-icon
            style="margin-bottom: 16px"
          />
          <a-button
            type="primary"
            size="large"
            class="login-button"
            html-type="submit"
            :loading="loading"
            block
          >
            {{ $t('login.btn') }}
          </a-button>
        </a-form-item>
      </a-form>

      <!-- 第二步：TOTP 验证码 -->
      <a-form
        v-else-if="step === 'totp'"
        ref="totpFormRef"
        :model="totpForm"
        :rules="getTotpRules()"
        class="login-form"
        @finish="handleVerifyTotp"
      >
        <p class="totp-description">{{ $t('login.totp.description') }}</p>

        <a-form-item name="code">
          <a-input
            v-model:value="totpForm.code"
            :placeholder="$t('login.totp.placeholder')"
            size="large"
            allow-clear
            maxlength="6"
            @pressEnter="handleVerifyTotp"
          >
            <template #prefix>
              <SafetyOutlined />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item>
          <a-alert
            v-if="loginError"
            :message="loginError"
            type="error"
            show-icon
            style="margin-bottom: 16px"
          />
          <a-button
            type="primary"
            size="large"
            class="login-button"
            html-type="submit"
            :loading="loading"
            block
          >
            {{ $t('login.totp.verify') }}
          </a-button>
          <a-button
            size="large"
            class="login-button"
            style="margin-top: 8px"
            :disabled="loading"
            block
            @click="backToCredentials"
          >
            {{ $t('login.totp.back') }}
          </a-button>
        </a-form-item>
      </a-form>

      <div class="login-footer">
        <a href="https://github.com/BernardSimon/etl-go" target="_blank"  class="text-gray-400">{{ $t('login.copyright') }}</a>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUserStore } from "../stores/user";
import { message } from "ant-design-vue";
import { UserOutlined, LockOutlined, SafetyOutlined } from "@ant-design/icons-vue";
import type { LoginRequest } from "../types";
import type { Rule } from "ant-design-vue/es/form";
import { useI18n } from "vue-i18n";
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

// 步骤：credentials | totp
const step = ref<"credentials" | "totp">("credentials");
const preAuthToken = ref("");

// 登录表单数据
const loginForm = reactive<LoginRequest>({
  username: "",
  password: "",
});

// TOTP 表单数据
const totpForm = reactive({ code: "" });

// 加载状态
const loading = ref(false);
const loginError = ref("");

// 表单验证规则
const getLoginRules = (): Record<string, Rule[]> => ({
  username: [
    { required: true, message: t("login.account.alert"), trigger: "blur" },
  ],
  password: [
    { required: true, message: t("login.password.alert"), trigger: "blur" },
  ],
});

const getTotpRules = (): Record<string, Rule[]> => ({
  code: [
    { required: true, message: t("login.totp.alert"), trigger: "blur" },
  ],
});

// 处理登录
const handleLogin = async () => {
  loginError.value = "";
  loading.value = true;
  try {
    const res: any = await userStore.login({ username: loginForm.username, password: loginForm.password });
    if (res && res.code === 0) {
      if (res.data?.requires_2fa) {
        preAuthToken.value = res.data.pre_auth_token;
        step.value = "totp";
        return;
      }
      message.success(t("login.success"));
      const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/";
      await router.push(redirect);
    }
  } catch (error: any) {
    loginError.value = error?.message || t("login.error");
    console.error(t("login.error"), error);
  } finally {
    loading.value = false;
  }
};

// 处理 TOTP 验证
const handleVerifyTotp = async () => {
  loginError.value = "";
  loading.value = true;
  try {
    const res: any = await userStore.verifyTwoFactor(preAuthToken.value, totpForm.code);
    if (res && res.code === 0) {
      message.success(t("login.success"));
      const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/";
      await router.push(redirect);
    }
  } catch (error: any) {
    loginError.value = error?.message || t("login.error");
    console.error(t("login.error"), error);
  } finally {
    loading.value = false;
  }
};

// 返回第一步
const backToCredentials = () => {
  step.value = "credentials";
  preAuthToken.value = "";
  totpForm.code = "";
  loginError.value = "";
};

const changeLanguage = (e: any) => {
  userStore.changeLanguage(e.target.value);
};
</script>

<style scoped lang="scss">
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100vh;
  background-size: cover;
  background-position: center;
  background-blend-mode: overlay;
  background-color: white;

  .login-box {
    width: 400px;
    padding: 20px 40px 40px;
    background-color: #fff;
    border-radius: 10px;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
    .language-select {
      margin-bottom: 20px;
      display: flex;
      justify-content: end;
    }
    .login-header {
      text-align: center;
      margin-bottom: 30px;

      h2 {
        font-size: 28px;
        color: #333;
        margin-bottom: 10px;
      }

      p {
        font-size: 14px;
        color: #999;
      }
    }

    .login-button {
      width: 100%;
      margin-top: 10px;
    }

    .totp-description {
      font-size: 14px;
      color: #666;
      text-align: center;
      margin-bottom: 24px;
    }

    .login-footer {
      margin-top: 30px;
      text-align: center;

      p {
        font-size: 12px;
        color: #999;
      }
    }
  }
}


</style>
