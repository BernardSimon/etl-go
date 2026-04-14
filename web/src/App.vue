<template>
  <a-config-provider :locale="locale" :theme="themeConfig">
    <router-view />
  </a-config-provider>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { theme } from "ant-design-vue";
import zhCN from "ant-design-vue/es/locale/zh_CN";
import enUS from "ant-design-vue/es/locale/en_US";
import { useUserStore } from "./stores/user.ts";
import { usePreferencesStore } from "./stores/preferences";
import { loadLanguageAsync } from "./i18n.ts";

const supportedLocales = {
  zh: zhCN,
  en: enUS,
} as const;

type SupportedLocale = typeof zhCN | typeof enUS;

const locale = ref<SupportedLocale>();
const userStore = useUserStore();
const preferencesStore = usePreferencesStore();

// 设置语言并同步到 store 与 i18n
const setLanguage = async (lang: keyof typeof supportedLocales) => {
  userStore.changeLanguage(lang);
  locale.value = supportedLocales[lang];
  await loadLanguageAsync(lang);
};

// 监听store中language的变化
watch(
  () => userStore.language,
  async (newLang) => {
    if (newLang && newLang in supportedLocales) {
      locale.value = supportedLocales[newLang as keyof typeof supportedLocales];
      await loadLanguageAsync(newLang);
    }
  },
  { immediate: true }
);

onMounted(() => {
  preferencesStore.applyTheme(preferencesStore.theme);

  // 优先使用store中的语言设置
  if (userStore.language && userStore.language in supportedLocales) {
    setLanguage(userStore.language as keyof typeof supportedLocales);
    return;
  }

  // 如果store中没有设置，则获取浏览器语言
  const browserLanguage = navigator.language.split('-')[0];

  // 类型守卫：判断是否为支持的语言
  if (browserLanguage in supportedLocales) {
    setLanguage(browserLanguage as keyof typeof supportedLocales);
  } else {
    setLanguage("en");
  }
});

const themeConfig = computed(() => ({
  algorithm: preferencesStore.isDark ? theme.darkAlgorithm : theme.defaultAlgorithm,
  token: {
    colorPrimary: "#2f6fed",
    colorInfo: "#2f6fed",
    colorSuccess: "#2d9d78",
    colorWarning: "#d7922a",
    colorError: "#d75c57",
    borderRadius: 16,
    fontFamily: "'IBM Plex Sans', 'Segoe UI', sans-serif",
  },
}));
</script>
