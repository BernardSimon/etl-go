<template>
  <a-layout style="min-height: 100vh">
    <a-layout-sider v-model:collapsed="collapsed" :width="200" class="sider">
      <div class="logo">
        <div v-if="!collapsed" class="logo-text">ETL-GO</div>
        <div v-else class="logo-text-mini">ETL</div>
      </div>
      <a-menu
        v-model:selectedKeys="selectedKeys"
        v-model:openKeys="openKeys"
        theme="dark"
        mode="inline"
      >
        <template v-for="item in sidebarItems" :key="item.index">
          <template v-if="!item.children">
            <a-menu-item :key="item.index" @click="navigateTo(item.index)">
              <template #icon>
                <component :is="item.icon" />
              </template>
              <span>{{ $t(item.title) }}</span>
            </a-menu-item>
          </template>
          <template v-else>
            <a-sub-menu :key="item.index">
              <template #title>
                <span>
                  <component :is="item.icon" />
                  <span>{{ $t(item.title) }}</span>
                </span>
              </template>
              <a-menu-item
                v-for="child in item.children"
                :key="child.index"
                @click="navigateTo(child.index)"
              >
                <template #icon>
                  <component :is="child.icon" />
                </template>
                <span>{{ t(child.title) }}</span>
              </a-menu-item>
            </a-sub-menu>
          </template>
        </template>
      </a-menu>
    </a-layout-sider>
    <a-layout>
      <a-layout-header class="layout-header">
        <div class="header-left">
          <a-button
            type="text"
            shape="circle"
            class="trigger-button"
            :title="t('layout.toggleSidebar')"
            :aria-label="t('layout.toggleSidebar')"
            @click="() => (collapsed = !collapsed)"
          >
            <template #icon>
              <menu-unfold-outlined v-if="collapsed" class="trigger" />
              <menu-fold-outlined v-else class="trigger" />
            </template>
          </a-button>

          <!-- 面包屑 -->
          <a-breadcrumb class="breadcrumb">
            <a-breadcrumb-item>{{$t("layout.home")}}</a-breadcrumb-item>
            <a-breadcrumb-item
              v-for="(item, index) in breadcrumbs"
              :key="index"
            >
              {{ $t(item) }}
            </a-breadcrumb-item>
          </a-breadcrumb>
        </div>
        <!-- 右侧操作区：操作手册 + 用户下拉菜单 -->
      <div class="header-right">
        <!-- 切换语言菜单 -->
        <a-dropdown>
          <div
            class="language"
            role="button"
            tabindex="0"
            :aria-label="t('layout.language')"
          >
            <span class="currentLanguage">{{ t("layout.language") }}</span>
          </div>
          <template #overlay>
            <a-menu @click="handleLanguageClick">
              <a-menu-item key="zh">
                {{ t("layout.language.zh") }}
              </a-menu-item>
              <a-menu-item key="en">
                {{ t("layout.language.en") }}
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
          <!-- 用户下拉菜单 -->
          <a-dropdown>
            <div
              class="user-info"
              role="button"
              tabindex="0"
              :aria-label="t('layout.userMenu')"
            >
              <a-avatar :size="32">
                <template #icon><UserOutlined /></template>
              </a-avatar>
              <span class="username">{{ t("layout.username") }}</span>
            </div>
            <template #overlay>
              <a-menu @click="handleMenuClick">
                <a-menu-item key="refresh">
                  <ReloadOutlined />
                  {{ $t("layout.refreshBaseData") }}
                </a-menu-item>
                <a-menu-divider />
                <a-menu-item key="logout">
                  <LogoutOutlined />
                  {{ $t("layout.logout") }}
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <!-- 标签页视图 -->
      <div class="tabs-view-container">
        <a-tabs
          v-model:activeKey="activeTabKey"
          type="editable-card"
          hide-add
          size="small"
          @edit="onTabEdit"
          @change="onTabChange"
          class="layout-tabs"
        >
          <a-tab-pane
            v-for="pane in panes"
            :key="pane.key"
            :tab="$t(pane.title)"
            :closable="pane.closable"
          />
        </a-tabs>
      </div>

      <a-layout-content class="layout-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </a-layout-content>
      <a-layout-footer class="text-center text-xs py-4">
        <a
          class="cursor-pointer text-gray-400"
          href="https://github.com/BernardSimon/etl-go"
          target="_blank"
        >
          Powered by ETL-GO
        </a>
        <span class="text-gray-400"> · v0.2.6</span>
      </a-layout-footer>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUserStore } from "../stores/user";
import { Modal, message } from "ant-design-vue";
import { useI18n } from "vue-i18n"
const { t } = useI18n()
import {
  UserOutlined,
  LogoutOutlined,
  ReloadOutlined,
  MenuUnfoldOutlined,
  MenuFoldOutlined,
} from "@ant-design/icons-vue";
import { getDataSourceTypeList } from "../api/datasource";
import { getTypeByComponent } from "../api/mission";
import { sidebarItems } from "./sidebarItems";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

const collapsed = ref(false);
const selectedKeys = ref<string[]>([]);
const openKeys = ref<string[]>([]);

// 面包屑
const breadcrumbs = computed(() => {
  const matched = findMenuPath(route.path, sidebarItems);
  return matched ? matched.map((m) => m.title) : [];
});

// 标签页
interface TabPane {
  title: string;
  key: string;
  closable?: boolean;
  path: string;
}

const activeTabKey = ref("");
const panes = ref<TabPane[]>([
  { title: "router.datasource", key: "/datasource", path: "/datasource", closable: false }, // 默认标签页
]);

// 导航
const navigateTo = (path: string) => {
  router.push(path);
};

// 查找菜单路径用于面包屑和展开的菜单
const findMenuPath = (
  path: string,
  items: any[],
  parentPath: any[] = []
): any[] | null => {
  for (const item of items) {
    if (item.index === path) {
      return [...parentPath, item];
    }
    if (item.children) {
      const found = findMenuPath(path, item.children, [...parentPath, item]);
      if (found) return found;
    }
  }
  return null;
};

// 添加标签页
const addTab = (path: string) => {
  const title = (route.meta.title as string) || "common.unknown";
  const existing = panes.value.find((p) => p.key === path);
  if (!existing) {
    panes.value.push({
      title,
      key: path,
      path,
      closable: true,
    });
  }
  activeTabKey.value = path;
};

// 删除标签页
const removeTab = (targetKey: string) => {
  let lastIndex = -1;
  panes.value.forEach((pane, i) => {
    if (pane.key === targetKey) {
      lastIndex = i - 1;
    }
  });

  const panesKeep = panes.value.filter((pane) => pane.key !== targetKey);
  if (panesKeep.length && activeTabKey.value === targetKey) {
    if (lastIndex >= 0) {
      activeTabKey.value = panesKeep[lastIndex].key;
    } else {
      activeTabKey.value = panesKeep[0].key;
    }
    router.push(activeTabKey.value);
  }
  panes.value = panesKeep;
};

// 标签页变化
const onTabChange = (key: any) => {
  router.push(key as string);
};

// 标签页编辑
const onTabEdit = (targetKey: any, action: "add" | "remove") => {
  if (action === "remove") {
    removeTab(targetKey as string);
  }
};

// 同步菜单状态与路由
watch(
  () => route.path,
  (newPath) => {
    selectedKeys.value = [newPath];

    // 自动展开子菜单
    const matched = findMenuPath(newPath, sidebarItems);
    if (matched && matched.length > 1) {
      const parent = matched[0];
      if (!openKeys.value.includes(parent.index)) {
        openKeys.value.push(parent.index);
      }
    }

    // 添加标签页
    addTab(newPath);
  },
  { immediate: true }
);

// 用户菜单
const handleMenuClick = async (e: any) => {
  if (e.key === "refresh") {
    await Promise.all([
      getDataSourceTypeList(true),
      getTypeByComponent(true),
    ]);
    message.success(t("layout.refreshBaseData.success"));
  }
  if (e.key === "logout") {
    Modal.confirm({
      title:  t("layout.logout.title"),
      content:  t("layout.logout.content"),
      onOk: async () => {
        await userStore.logout();
        router.push("/login");
      },
    });
  }
};

// 语言切换
const handleLanguageClick = async (e: any) => {
  if (e.key === "zh") {
    userStore.changeLanguage("zh");
  }
  if (e.key === "en") {
    userStore.changeLanguage("en");
  }

};
</script>

<style scoped lang="scss">
.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #002140;
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  overflow: hidden;
  transition: all 0.3s;

  .logo-text {
    white-space: nowrap;
  }

  .logo-text-mini {
    font-size: 18px;
  }
}

.layout-header {
  background: #fff;
  padding: 0 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  z-index: 1;
  height: 64px;
  line-height: 64px;

  .header-left {
    display: flex;
    align-items: center;

    .trigger {
      font-size: 18px;
      transition: color 0.3s;

      &:hover {
        color: #1890ff;
      }
    }

    .breadcrumb {
      display: inline-block;
    }
  }
  .header-right{
    display: flex;
    align-items: center;

    .user-info {
      display: flex;
      align-items: center;
      cursor: pointer;
      margin-left: 20px;

      .username {
        margin-left: 8px;
        color: rgba(0, 0, 0, 0.65);
      }
    }
  }

}

.trigger-button {
  margin-right: 8px;
}

.sider-version {
  text-align: center;
  color: rgba(255, 255, 255, 0.3);
  font-size: 11px;
  padding: 12px 0 8px;
  letter-spacing: 0.5px;

  &--collapsed {
    font-size: 10px;
  }
}

.tabs-view-container {
  padding: 6px 16px 0;
  background: #fff;
  border-top: 1px solid #f0f0f0;

  :deep(.ant-tabs-nav) {
    margin-bottom: 0;
  }
}

.layout-content {
  margin: 16px;
  background: #fff;
  min-height: 280px;
  overflow-y: auto;
  border-radius: 6px;
  // padding: 24px; // 内容填充通常由页面本身或这里处理
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

@media (max-width: 768px) {
  .layout-header {
    height: auto;
    min-height: 64px;
    padding: 8px 12px;
    line-height: 1.4;
    flex-wrap: wrap;
    gap: 8px;

    .header-left,
    .header-right {
      width: 100%;
      justify-content: space-between;
    }

    .header-right {
      gap: 12px;
    }

    .header-left .breadcrumb {
      display: none;
    }
  }

  .tabs-view-container {
    padding: 6px 8px 0;
  }

  .layout-content {
    margin: 8px;
  }
}
</style>
