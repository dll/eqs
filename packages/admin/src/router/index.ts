import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
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
        { path: 'dashboard', name: 'Dashboard', component: () => import('@/views/dashboard/index.vue'), meta: { titleKey: 'dashboard' } },
        { path: 'audit', name: 'Audit', component: () => import('@/views/audit/index.vue'), meta: { titleKey: 'audit' } },
        { path: 'project', name: 'Project', component: () => import('@/views/project/index.vue'), meta: { titleKey: 'project' } },
        { path: 'order', name: 'Order', component: () => import('@/views/order/index.vue'), meta: { titleKey: 'order' } },
        { path: 'settlement', name: 'Settlement', component: () => import('@/views/settlement/index.vue'), meta: { titleKey: 'settlement' } },
        { path: 'credit', name: 'Credit', component: () => import('@/views/credit/index.vue'), meta: { titleKey: 'credit' } },
        { path: 'dispute', name: 'Dispute', component: () => import('@/views/dispute/index.vue'), meta: { titleKey: 'dispute' } },
        { path: 'user', name: 'User', component: () => import('@/views/user/index.vue'), meta: { titleKey: 'user' } },
        { path: 'settings', name: 'Settings', component: () => import('@/views/settings/index.vue'), meta: { titleKey: 'settings' } },
      ],
    },
  ],
})

export default router
