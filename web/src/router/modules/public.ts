import { RouteRecordRaw } from 'vue-router'

// Public (no-auth) routes
export const publicRoutes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../../views/Login.vue'),
    meta: {
      title:"router.login",
      requiresAuth: false,
    },
  },
]


