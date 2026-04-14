import { computed, ref, watch } from "vue";
import { defineStore } from "pinia";
import { getTheme, setTheme } from "../utils/storage";

export type AppTheme = "light" | "dark";

const detectPreferredTheme = (): AppTheme => {
  if (typeof window === "undefined") {
    return "light";
  }
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
};

export const usePreferencesStore = defineStore("preferences", () => {
  const theme = ref<AppTheme>((getTheme() as AppTheme) || detectPreferredTheme());
  const sidebarCollapsed = ref(false);

  const isDark = computed(() => theme.value === "dark");

  const applyTheme = (nextTheme: AppTheme) => {
    theme.value = nextTheme;
    if (typeof document !== "undefined") {
      document.documentElement.dataset.theme = nextTheme;
      document.body.dataset.theme = nextTheme;
    }
    setTheme(nextTheme);
  };

  const toggleTheme = () => {
    applyTheme(theme.value === "dark" ? "light" : "dark");
  };

  const setSidebarCollapsed = (value: boolean) => {
    sidebarCollapsed.value = value;
  };

  watch(
    theme,
    (value) => {
      applyTheme(value);
    },
    { immediate: true }
  );

  return {
    theme,
    isDark,
    sidebarCollapsed,
    applyTheme,
    toggleTheme,
    setSidebarCollapsed,
  };
});
