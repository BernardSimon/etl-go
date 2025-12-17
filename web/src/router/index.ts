import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import { useUserStore } from "../stores/user";
import { publicRoutes } from "./modules/public";
import { layoutRoute } from "./modules/layout";
import { errorRoutes } from "./modules/error";

// 路由配置（分层组织）
const routes: RouteRecordRaw[] = [...publicRoutes, layoutRoute, ...errorRoutes];

// 创建路由实例
const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

// 路由守卫 - Token 验证
router.beforeEach((to, _from, next) => {
  const userStore = useUserStore();

  // 判断是否需要登录
  if (to.meta.requiresAuth) {
    // 需要登录验证
    if (userStore.token) {
      // 已登录，放行
      next();
    } else {
      // 未登录，跳转到登录页
      // message.info("请先登录");
      next({
        path: "/login",
        query: { redirect: to.fullPath }, // 保存目标路由，登录后跳转
      });
    }
  } else {
    // 不需要登录验证
    if (to.path === "/login" && userStore.token) {
      // 已登录状态访问登录页，跳转到首页
      next("/");
    } else {
      next();
    }
  }
});

export default router;
