<template>
  <div class="login-container">

    <div class="login-box">
      <a-radio-group class="language-select" v-model:value="userStore.language" @change="changeLanguage" button-style="solid" size="small">
        <a-radio-button value="zh">中文</a-radio-button>
        <a-radio-button value="en">English</a-radio-button>
      </a-radio-group>
      <div class="login-header">

          <h2>{{ $t('login.title') }}</h2>

      </div>

      <a-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="getLoginRules()"
        class="login-form"
        @finish="handleLogin"
      >
        <template v-if="!showCodeInput">
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
        </template>

        <a-form-item>
          <a-button
            type="primary"
            size="large"
            class="login-button"
            html-type="submit"
            :loading="loading"
            block
          >
            {{
              $t('login.btn')
            }}
          </a-button>

        </a-form-item>
      </a-form>

      <div class="login-footer">
        <a href="https://github.com/BernardSimon/etl-go" target="_blank"  class="text-gray-400">© 2025 ETL-GO</a>

      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { useUserStore } from "../stores/user";
import { message } from "ant-design-vue";
import { UserOutlined, LockOutlined } from "@ant-design/icons-vue";
import type { LoginRequest } from "../types";
import type { Rule } from "ant-design-vue/es/form";
import { useI18n } from "vue-i18n";
const { t } = useI18n();
const router = useRouter();
const userStore = useUserStore();

// 是否显示验证码输入框
const showCodeInput = ref(false);

// 登录表单数据
const loginForm = reactive<LoginRequest>({
  username: "",
  password: "",
});

// 加载状态
const loading = ref(false);

// 表单验证规则
const getLoginRules = (): Record<string, Rule[]> => ({
  username: [
    { required: true, message: t("login.account.alert"), trigger: "blur" },
  ],
  password: [
    { required: true, message: t("login.password.alert"), trigger: "blur" },
  ],
});

// 处理登录 
const handleLogin = async () => {
  loading.value = true;
  try {
    let res:any;
      res = await userStore.login({ username: loginForm.username, password: loginForm.password });
      if (res && res.code === 0) {
        message.success(t("login.success"));
        await router.push("/");
        return;
      }
  } catch (error: any) { 
    console.error(t("login.error"), error);
  } finally {
    loading.value = false;
  }
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
