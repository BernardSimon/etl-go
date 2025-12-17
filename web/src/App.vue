<template>
  <a-config-provider :locale="locale">
    <router-view />
  </a-config-provider>
</template>



<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import zhCN from 'ant-design-vue/es/locale/zh_CN';
import enUS from 'ant-design-vue/es/locale/en_US';
import {useUserStore} from "./stores/user.ts";
import {loadLanguageAsync} from "./i18n.ts";
const supportedLocales = {
  'zh': zhCN,
  'en': enUS
} as const;

type SupportedLocale = typeof zhCN | typeof enUS;

const locale = ref<SupportedLocale>();
const userStore = useUserStore();

// 设置语言并同步到store
const setLanguage = (lang: keyof typeof supportedLocales) => {
  userStore.changeLanguage( lang);
  // userStore.setLanguage(lang);
};

// 监听store中language的变化
watch(
    () => userStore.language,
    (newLang) => {
      if (newLang && newLang in supportedLocales) {
        locale.value = supportedLocales[newLang as keyof typeof supportedLocales];
        loadLanguageAsync(newLang)
      }
    }
);

onMounted(() => {
  // 优先使用store中的语言设置
  if (userStore.language && userStore.language in supportedLocales) {
    locale.value = supportedLocales[userStore.language as keyof typeof supportedLocales];
    return;
  }

  // 如果store中没有设置，则获取浏览器语言
  const browserLanguage = navigator.language.split('-')[0];

  // 类型守卫：判断是否为支持的语言
  if (browserLanguage in supportedLocales) {
    setLanguage(browserLanguage as keyof typeof supportedLocales);
  } else {
    setLanguage('en');
  }
});
</script>
