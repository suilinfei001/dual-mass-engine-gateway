import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../views/Home.vue')
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue')
  },
  {
    path: '/resources',
    name: 'Resources',
    component: () => import('../views/Resources.vue')
  },
  {
    path: '/tasks',
    name: 'Tasks',
    component: () => import('../views/Tasks.vue')
  },
  {
    path: '/console',
    name: 'Console',
    component: () => import('../views/Console.vue'),
    meta: { requiresAdmin: true }
  },
  {
    path: '/events',
    name: 'Events',
    component: () => import('../views/Events.vue')
  },
  {
    path: '/documentation',
    name: 'Documentation',
    component: () => import('../views/Documentation.vue')
  },
  {
    path: '/api-reference',
    name: 'ApiReference',
    component: () => import('../views/ApiReference.vue')
  },
  // Resource Pool Routes
  {
    path: '/resource-pool',
    name: 'ResourcePool',
    component: () => import('../views/resource-pool/ResourcePoolHub.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/resource-pool/my-testbeds',
    name: 'MyTestbeds',
    component: () => import('../views/resource-pool/MyTestbedList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/resource-pool/my-resources',
    name: 'MyResourceInstances',
    component: () => import('../views/resource-pool/MyResourceInstanceList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/resource-pool/testbeds',
    name: 'TestbedList',
    component: () => import('../views/resource-pool/TestbedList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/resource-pool/testbeds/:uuid',
    name: 'TestbedDetail',
    component: () => import('../views/resource-pool/TestbedDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/resource-pool/allocations/:uuid',
    name: 'AllocationDetail',
    component: () => import('../views/resource-pool/TestbedDetail.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/resource-pool/resources',
    name: 'ResourceInstances',
    component: () => import('../views/resource-pool/ResourceInstanceList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/resource-pool/categories',
    name: 'Categories',
    component: () => import('../views/resource-pool/CategoryManage.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/resource-pool/quota-policies',
    name: 'QuotaPolicy',
    component: () => import('../views/resource-pool/QuotaPolicy.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/resource-pool/history',
    name: 'AllocationHistory',
    component: () => import('../views/resource-pool/AllocationHistory.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/resource-pool/dashboard',
    name: 'MetricsDashboard',
    component: () => import('../views/resource-pool/MetricsDashboard.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  },
  {
    path: '/resource-pool/pipeline-templates',
    name: 'PipelineTemplates',
    component: () => import('../views/resource-pool/PipelineTemplateManage.vue'),
    meta: { requiresAuth: true, requiresAdmin: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 处理路由加载错误
router.onError((error) => {
  console.error('Router error:', error)
  // 如果是动态导入失败，提示用户刷新页面
  if (error.message && error.message.includes('Failed to fetch dynamically imported module')) {
    console.warn('动态导入失败，可能需要刷新页面')
    // 不自动刷新，避免无限循环，而是给用户提示
  }
})

router.beforeEach(async (to, from, next) => {
  const isLoggedIn = localStorage.getItem('isLoggedIn') === 'true'
  const userRole = localStorage.getItem('userRole')

  if (to.meta.requiresAdmin && userRole !== 'admin') {
    next('/')
    return
  }

  if (to.meta.requiresAuth && !isLoggedIn) {
    next('/login')
    return
  }

  next()
})

export default router
