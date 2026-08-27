<template>
  <div class="app-layout" :class="{ 'app-layout--dark': preferencesStore.isDark }">
    <aside v-if="!isMobile" class="app-sidebar app-ghost" :class="{ 'app-sidebar--collapsed': collapsed }">
      <div class="app-sidebar__brand">
        <div class="app-sidebar__mark">ETL</div>
        <div v-if="!collapsed" class="app-sidebar__copy">
          <strong>ETL-GO</strong>
        </div>
      </div>

      <nav class="app-sidebar__nav">
        <button
          v-for="item in sidebarItems"
          :key="item.index"
          type="button"
          class="app-nav-item"
          :class="{ 'app-nav-item--active': route.path === item.index }"
          @click="navigateTo(item.index)"
        >
          <component :is="item.icon" class="app-nav-item__icon" />
          <div v-if="!collapsed" class="app-nav-item__copy">
            <span>{{ t(item.title) }}</span>
          </div>
        </button>
      </nav>

      <div class="app-sidebar__footer" :class="{ 'app-sidebar__footer--compact': collapsed }">
        <a-button
          type="text"
          class="app-sidebar__collapse"
          :aria-label="t('layout.toggleSidebar')"
          @click="toggleCollapsed"
        >
          <template #icon>
            <MenuUnfoldOutlined v-if="collapsed" />
            <MenuFoldOutlined v-else />
          </template>
        </a-button>
      </div>
    </aside>

    <a-drawer
      :open="mobileMenuOpen"
      placement="left"
      width="300"
      class="app-mobile-drawer"
      @close="mobileMenuOpen = false"
    >
      <div class="app-mobile-drawer__body">
        <div class="app-sidebar__brand app-sidebar__brand--mobile">
          <div class="app-sidebar__mark">ETL</div>
          <div class="app-sidebar__copy">
            <strong>ETL-GO</strong>
          </div>
        </div>
        <nav class="app-sidebar__nav">
          <button
            v-for="item in sidebarItems"
            :key="item.index"
            type="button"
            class="app-nav-item"
            :class="{ 'app-nav-item--active': route.path === item.index }"
            @click="navigateWithClose(item.index)"
          >
            <component :is="item.icon" class="app-nav-item__icon" />
            <div class="app-nav-item__copy">
              <span>{{ t(item.title) }}</span>
            </div>
          </button>
        </nav>
      </div>
    </a-drawer>

    <main class="app-main">
      <header class="app-topbar app-ghost">
        <div class="app-topbar__left">
          <a-button v-if="isMobile" type="text" class="app-topbar__menu" @click="mobileMenuOpen = true">
            <template #icon>
              <MenuUnfoldOutlined />
            </template>
          </a-button>
          <div v-if="!isMobile" class="app-topbar__headline">
            <span class="app-topbar__label">{{ t("layout.currentPage") }}</span>
            <h2>{{ title }}</h2>
          </div>
        </div>

        <div class="app-topbar__right">
          <a-segmented
            v-model:value="preferencesStore.theme"
            :options="themeOptions"
            size="small"
          />

          <a-dropdown>
            <a-button class="app-toolbar-button" size="small">
              <template #icon><GlobalOutlined /></template>
              {{ currentLanguageLabel }}
            </a-button>
            <template #overlay>
              <a-menu @click="handleLanguageClick">
                <a-menu-item key="zh">{{ t("layout.language.zh") }}</a-menu-item>
                <a-menu-item key="en">{{ t("layout.language.en") }}</a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>

          <a-dropdown>
            <button class="app-user-chip" type="button">
              <span class="app-user-chip__avatar">
                <UserOutlined />
              </span>
              <span class="app-user-chip__name">{{ t("layout.username") }}</span>
            </button>
            <template #overlay>
              <a-menu @click="handleMenuClick">
                <a-menu-item key="refresh">
                  <ReloadOutlined />
                  {{ t("layout.refreshBaseData") }}
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout">
                  <LogoutOutlined />
                  {{ t("layout.logout") }}
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </header>

      <div class="app-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </div>

      <footer class="app-footer">
        <a href="https://github.com/BernardSimon/etl-go" target="_blank" rel="noreferrer">
          Powered by ETL-GO · v0.2.8
        </a>
      </footer>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { message, Modal } from "ant-design-vue";
import {
  GlobalOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  ReloadOutlined,
  UserOutlined,
} from "@ant-design/icons-vue";
import { useI18n } from "vue-i18n";
import { getDataSourceTypeList } from "../api/datasource";
import { getTypeByComponent } from "../api/mission";
import { useResponsive } from "../composables/useResponsive";
import { usePreferencesStore } from "../stores/preferences";
import { useUserStore } from "../stores/user";
import { sidebarItems } from "./sidebarItems";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();
const preferencesStore = usePreferencesStore();
const { isMobile } = useResponsive();
const { t } = useI18n();

const mobileMenuOpen = ref(false);
const title = computed(() => {
  const current = sidebarItems.find((item) => item.index === route.path);
  return current ? t(current.title) : t("layout.home");
});

const collapsed = computed(() => preferencesStore.sidebarCollapsed && !isMobile.value);
const currentLanguageLabel = computed(() =>
  userStore.language === "zh" ? t("layout.language.zh") : t("layout.language.en")
);

const themeOptions = computed(() => [
  { label: t("layout.theme.light"), value: "light" },
  { label: t("layout.theme.dark"), value: "dark" },
]);

const navigateTo = (path: string) => {
  router.push(path);
};

const navigateWithClose = (path: string) => {
  mobileMenuOpen.value = false;
  navigateTo(path);
};

const toggleCollapsed = () => {
  preferencesStore.setSidebarCollapsed(!preferencesStore.sidebarCollapsed);
};

watch(
  isMobile,
  (value) => {
    if (value) {
      mobileMenuOpen.value = false;
    }
  },
  { immediate: true }
);

const handleLanguageClick = async (e: { key: string | number }) => {
  const key = String(e.key);
  if (key === "zh" || key === "en") {
    userStore.changeLanguage(key);
  }
};

const handleMenuClick = async (e: { key: string | number }) => {
  const key = String(e.key);
  if (key === "theme") {
    preferencesStore.toggleTheme();
    return;
  }

  if (key === "refresh") {
    await Promise.all([getDataSourceTypeList(true), getTypeByComponent(true)]);
    message.success(t("layout.refreshBaseData.success"));
    return;
  }

  if (key === "logout") {
    Modal.confirm({
      title: t("layout.logout.title"),
      content: t("layout.logout.content"),
      onOk: async () => {
        userStore.logout();
        await router.push("/login");
      },
    });
  }
};
</script>

<style scoped lang="scss">
.app-layout {
  min-height: 100vh;
  display: flex;
  gap: 20px;
  padding: 20px;
}

.app-sidebar {
  position: sticky;
  top: 20px;
  height: calc(100vh - 40px);
  width: 200px;
  padding: 18px;
  border-radius: var(--app-radius-xl);
  display: flex;
  flex-direction: column;
  gap: 18px;
  transition: width 0.25s ease;
}

.app-sidebar--collapsed {
  width: 100px;
}

.app-sidebar__brand {
  display: flex;
  gap: 14px;
  align-items: center;
  padding: 8px 6px 18px;
}

.app-sidebar__brand--mobile {
  padding-inline: 0;
}

.app-sidebar__mark {
  width: 48px;
  height: 48px;
  border-radius: var(--app-radius-md);
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, var(--app-primary), #79b2ff);
  color: white;
  font-weight: 800;
  letter-spacing: 0.08em;
}

.app-sidebar__copy {
  display: flex;
  flex-direction: column;
}

.app-sidebar__copy strong {
  font-size: 18px;
}

.app-sidebar__nav {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.app-nav-item {
  width: 100%;
  border: 0;
  background: transparent;
  color: var(--app-text);
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px;
  border-radius: var(--app-radius-md);
  transition: all 0.2s ease;
  text-align: left;
  cursor: pointer;
}

.app-nav-item:hover,
.app-nav-item--active {
  background: var(--app-primary-soft);
  color: var(--app-primary);
}

.app-nav-item__icon {
  font-size: 18px;
  flex: 0 0 auto;
}

.app-nav-item__copy {
  min-width: 0;
}

.app-nav-item__copy span {
  font-weight: 700;
}

.app-sidebar__footer {
  margin-top: auto;
  display: flex;
  justify-content: flex-end;
}

.app-sidebar__footer--compact {
  justify-content: center;
}

.app-sidebar__collapse {
  color: var(--app-text) !important;
}

.app-main {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 14px;
  height: calc(100vh - 40px);
  overflow: hidden;
}

.app-topbar {
  border-radius: var(--app-radius-xl);
  padding: 14px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.app-topbar__left,
.app-topbar__right {
  display: flex;
  align-items: center;
  gap: 10px;
}


.app-topbar__headline {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.app-topbar__label {
  font-size: 11px;
  color: var(--app-text-faint);
  letter-spacing: 0.04em;
}

.app-topbar__headline h2 {
  font-size: 20px;
  line-height: 1.1;
  letter-spacing: -0.03em;
}

.app-topbar__menu,
.app-toolbar-button,
.app-user-chip {
  border-radius: 999px;
}

.app-toolbar-button {
  font-size: 12px;
}

.app-user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  border: 1px solid var(--app-border);
  background: var(--app-surface-muted);
  padding: 5px 12px 5px 5px;
  min-width: 0;
  cursor: pointer;
  color: var(--app-text);
  transition: background 0.15s ease;
}

.app-user-chip:hover {
  background: var(--app-primary-soft);
  border-color: var(--app-primary);
  color: var(--app-primary);
}

.app-user-chip__avatar {
  width: 28px;
  height: 28px;
  border-radius: var(--app-radius-sm);
  display: grid;
  place-items: center;
  background: linear-gradient(135deg, var(--app-primary), #79b2ff);
  color: white;
  font-size: 13px;
  flex: 0 0 auto;
}

.app-user-chip__name {
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
}

.app-content {
  min-width: 0;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.app-footer {
  display: flex;
  justify-content: center;
  align-items: center;
  color: var(--app-text-faint);
  font-size: 12px;
  padding: 0 4px 4px;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(8px);
}

.app-mobile-drawer :deep(.ant-drawer-body) {
  padding: 0;
}

.app-mobile-drawer__body {
  padding: 16px;
}

@media (max-width: 767px) {
  .app-layout {
    padding: 12px;
  }

  .app-topbar {
    padding: 12px 16px;
  }

  .app-topbar__right {
    gap: 8px;
  }
}
</style>
