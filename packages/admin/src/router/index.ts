import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue'),
    },
    {
      path: '/',
      component: () => import('@/views/layout/index.vue'),
      redirect: '/dashboard',
      children: [
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/index.vue'),
          meta: { title: '数据看板' },
        },
        {
          path: 'audit',
          name: 'Audit',
          component: () => import('@/views/audit/index.vue'),
          meta: { title: '资质审核' },
        },
        {
          path: 'project',
          name: 'Project',
          component: () => import('@/views/project/index.vue'),
          meta: { title: '项目管理' },
        },
        {
          path: 'order',
          name: 'Order',
          component: () => import('@/views/order/index.vue'),
          meta: { title: '订单管理' },
        },
        {
          path: 'settlement',
          name: 'Settlement',
          component: () => import('@/views/settlement/index.vue'),
          meta: { title: '结算中心' },
        },
        {
          path: 'credit',
          name: 'Credit',
          component: () => import('@/views/credit/index.vue'),
          meta: { title: '信用评分' },
        },
        {
          path: 'dispute',
          name: 'Dispute',
          component: () => import('@/views/dispute/index.vue'),
          meta: { title: '纠纷仲裁' },
        },
        {
          path: 'user',
          name: 'User',
          component: () => import('@/views/user/index.vue'),
          meta: { title: '用户管理' },
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/settings/index.vue'),
          meta: { title: '系统配置' },
        },
      ],
    },
  ],
})

export default router
