import { RouteRecordRaw } from 'vue-router'

// Error and fallback routes
export const errorRoutes: RouteRecordRaw[] = [
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('../../views/NotFound.vue'),
    meta: {
      title: 'router.404',
      requiresAuth: false,
    },
  },
]


