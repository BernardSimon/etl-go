<template>
  <div class="login-page">
    <section class="login-hero">
      <div class="login-hero__brand">
        <div class="login-hero__mark">ETL</div>
        <span class="login-hero__name">ETL-GO</span>
      </div>
      <h1>{{ t("login.hero.title") }}</h1>
    </section>

    <section class="login-panel app-ghost">
      <div class="login-panel__topbar">
        <a-segmented v-model:value="preferencesStore.theme" :options="themeOptions" size="small" />
        <a-radio-group
          v-model:value="userStore.language"
          button-style="solid"
          size="small"
          @change="changeLanguage"
        >
          <a-radio-button value="zh">{{ $t("layout.language.zh") }}</a-radio-button>
          <a-radio-button value="en">{{ $t("layout.language.en") }}</a-radio-button>
        </a-radio-group>
      </div>

      <div class="login-panel__body">
        <h2 class="login-panel__title">
          {{ step === "totp" ? $t("login.totp.title") : $t("login.title") }}
        </h2>

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
              <template #prefix><UserOutlined /></template>
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
              <template #prefix><LockOutlined /></template>
            </a-input-password>
          </a-form-item>

          <a-alert v-if="loginError" :message="loginError" type="error" show-icon class="login-alert" />

          <a-button type="primary" size="large" class="login-button" html-type="submit" :loading="loading" block>
            {{ $t("login.btn") }}
          </a-button>
        </a-form>

        <a-form
          v-else
          ref="totpFormRef"
          :model="totpForm"
          :rules="getTotpRules()"
          class="login-form"
          @finish="handleVerifyTotp"
        >
          <a-form-item name="code">
            <a-input
              v-model:value="totpForm.code"
              :placeholder="$t('login.totp.placeholder')"
              size="large"
              allow-clear
              :maxlength="6"
              @pressEnter="handleVerifyTotp"
            >
              <template #prefix><SafetyOutlined /></template>
            </a-input>
          </a-form-item>

          <a-alert v-if="loginError" :message="loginError" type="error" show-icon class="login-alert" />

          <div class="login-form__actions">
            <a-button type="primary" size="large" class="login-button" html-type="submit" :loading="loading" block>
              {{ $t("login.totp.verify") }}
            </a-button>
            <a-button size="large" class="login-button" :disabled="loading" block @click="backToCredentials">
              {{ $t("login.totp.back") }}
            </a-button>
          </div>
        </a-form>
      </div>

      <div class="login-panel__footer">
        <a href="https://github.com/BernardSimon/etl-go" target="_blank" rel="noreferrer">
          {{ $t("login.copyright") }}
        </a>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { message } from "ant-design-vue";
import type { Rule } from "ant-design-vue/es/form";
import { LockOutlined, SafetyOutlined, UserOutlined } from "@ant-design/icons-vue";
import { useI18n } from "vue-i18n";
import type { LoginRequest } from "../types";
import { usePreferencesStore } from "../stores/preferences";
import { useUserStore } from "../stores/user";

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const preferencesStore = usePreferencesStore();

const step = ref<"credentials" | "totp">("credentials");
const preAuthToken = ref("");
const loading = ref(false);
const loginError = ref("");

const loginForm = reactive<LoginRequest>({ username: "", password: "" });
const totpForm = reactive({ code: "" });

const themeOptions = computed(() => [
  { label: t("layout.theme.light"), value: "light" },
  { label: t("layout.theme.dark"), value: "dark" },
]);

const getLoginRules = (): Record<string, Rule[]> => ({
  username: [{ required: true, message: t("login.account.alert"), trigger: "blur" }],
  password: [{ required: true, message: t("login.password.alert"), trigger: "blur" }],
});

const getTotpRules = (): Record<string, Rule[]> => ({
  code: [{ required: true, message: t("login.totp.alert"), trigger: "blur" }],
});

const handleLogin = async () => {
  loginError.value = "";
  loading.value = true;
  try {
    const res: any = await userStore.login({ username: loginForm.username, password: loginForm.password });
    if (res?.code === 0) {
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
  } finally {
    loading.value = false;
  }
};

const handleVerifyTotp = async () => {
  loginError.value = "";
  loading.value = true;
  try {
    const res: any = await userStore.verifyTwoFactor(preAuthToken.value, totpForm.code);
    if (res?.code === 0) {
      message.success(t("login.success"));
      const redirect = typeof route.query.redirect === "string" ? route.query.redirect : "/";
      await router.push(redirect);
    }
  } catch (error: any) {
    loginError.value = error?.message || t("login.error");
  } finally {
    loading.value = false;
  }
};

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
.login-page {
  min-height: 100vh;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(360px, 440px);
  gap: 16px;
  padding: 16px;
}

// ── Hero ─────────────────────────────────────────────────────────────────────

.login-hero {
  border-radius: var(--app-radius-xl);
  padding: 36px 40px;
  display: flex;
  flex-direction: column;
  gap: 32px;
  background:
    radial-gradient(circle at 10% 20%, rgba(47, 111, 237, 0.18), transparent 40%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.7), rgba(255, 255, 255, 0.28));
  border: 1px solid var(--app-border);
  box-shadow: var(--app-shadow-lg);
}

:global(:root[data-theme="dark"]) .login-hero {
  background:
    radial-gradient(circle at 10% 20%, rgba(110, 163, 255, 0.2), transparent 40%),
    linear-gradient(145deg, rgba(10, 19, 34, 0.82), rgba(10, 19, 34, 0.4));
}

.login-hero__brand {
  display: flex;
  align-items: center;
  gap: 12px;
}

.login-hero__mark {
  width: 44px;
  height: 44px;
  border-radius: var(--app-radius-md);
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, var(--app-primary), #79b2ff);
  color: white;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.1em;
}

.login-hero__name {
  font-size: 17px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.login-hero h1 {
  font-size: clamp(32px, 5vw, 52px);
  line-height: 1.1;
  letter-spacing: -0.05em;
  max-width: 480px;
}

// ── Panel ─────────────────────────────────────────────────────────────────────

.login-panel {
  border-radius: var(--app-radius-xl);
  padding: 24px;
  display: flex;
  flex-direction: column;
}

.login-panel__topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.login-panel__body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 32px 0 24px;
}

.login-panel__title {
  font-size: 28px;
  letter-spacing: -0.04em;
  margin-bottom: 28px;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.login-form__actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.login-alert {
  margin-bottom: 8px;
}

.login-button {
  height: 44px;
}

.login-panel__footer {
  padding-top: 16px;
  border-top: 1px solid var(--app-border);
  color: var(--app-text-faint);
  font-size: 12px;
}

// ── Responsive ───────────────────────────────────────────────────────────────

@media (max-width: 1023px) {
  .login-page {
    grid-template-columns: 1fr;
  }

  .login-hero {
    min-height: 200px;
  }

  .login-hero h1 {
    font-size: 28px;
  }
}

@media (max-width: 600px) {
  .login-page {
    padding: 12px;
    gap: 12px;
  }

  .login-panel__topbar {
    flex-direction: column;
    align-items: stretch;
  }
}
</style>
