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
        description: "page.datasource.description",
        requiresAuth: true,
      },
    },
    {
      path: "/system-variables",
      name: "SystemVariables",
      component: () => import("../../views/SystemVariables.vue"),
      meta: {
        title: "router.systemVariable",
        description: "page.variable.description",
        requiresAuth: true, // 需要登录权限
      },
    },
    {
      path: "/workflow-management",
      name: "WorkflowManagement",
      component: () => import("../../views/WorkflowManagement.vue"),
      meta: {
        title: "router.task",
        description: "page.workflow.description",
        requiresAuth: true, // 需要登录权限
      },
    },
    {
      path: "/run-logs",
      name: "RunLogs",
      component: () => import("../../views/RunLogs.vue"),
      meta: {
        title: "router.runLog",
        description: "page.runlog.description",
        requiresAuth: true, // 需要登录权限
      },
    },
    {
      path: "/files",
      name: "files",
      component: () => import("../../views/file.vue"),
      meta: {
        title: "router.file",
        description: "page.file.description",
        requiresAuth: true, // 需要登录权限
      },
    },
  ],
};
