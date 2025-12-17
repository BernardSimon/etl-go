import { RouteRecordRaw } from "vue-router";
// 布局主路由，包含系统业务页面
export const layoutRoute: RouteRecordRaw = {
  path: "/",
  name: "Layout",
  component: () => import("../../layout/index.vue"),
  // 默认重定向到数据源页面
  redirect: "/datasource",
  meta: {
    requiresAuth: true,
  },
  children: [
    {
      path: "/datasource",
      name: "DataSource",
      component: () => import("../../views/DataSource.vue"),
      meta: {
        title: "router.datasource",
        requiresAuth: true,
      },
    },
    {
      path: "/system-variables",
      name: "SystemVariables",
      component: () => import("../../views/SystemVariables.vue"),
      meta: {
        title: "router.systemVariable",
        requiresAuth: true, // 需要登录权限
      },
    },
    {
      path: "/workflow-management",
      name: "WorkflowManagement",
      component: () => import("../../views/WorkflowManagement.vue"),
      meta: {
        title: "router.task",
        requiresAuth: true, // 需要登录权限
      },
    },
    {
      path: "/run-logs",
      name: "RunLogs",
      component: () => import("../../views/RunLogs.vue"),
      meta: {
        title: "router.runLog",
        requiresAuth: true, // 需要登录权限
      },
    },
    {
      path: "/files",
      name: "files",
      component: () => import("../../views/file.vue"),
      meta: {
        title: "router.file",
        requiresAuth: true, // 需要登录权限
      },
    },
  ],
};
