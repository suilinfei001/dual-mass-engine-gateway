/**
 * 微服务架构 API 配置
 * 统一管理所有服务的端点
 */

// 根据环境自动选择 API 基础 URL
const isProduction = window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1'
const isDev = !isProduction

// 服务端口配置（内网/外网环境）
export const SERVICES = isProduction
  ? {
      // 生产环境 - 外网服务
      baseURL: 'http://10.4.111.141',
      eventStore: 'http://10.4.111.141:4002',
      // 生产环境 - 内网服务（通过 Nginx 代理）
      taskScheduler: '/api/task-scheduler',
      executorService: '/api/executor-service',
      aiAnalyzer: '/api/ai-analyzer',
      resourceManager: '/api/resource-manager',
      webhookGateway: '/api/webhook-gateway'
    }
  : {
      // 开发环境 - 直接连接各服务
      eventStore: 'http://localhost:4002',
      taskScheduler: 'http://localhost:4003',
      executorService: 'http://localhost:4004',
      aiAnalyzer: 'http://localhost:4005',
      resourceManager: 'http://localhost:4006',
      webhookGateway: 'http://localhost:4001'
    }

// API 路径映射（旧路径 -> 新服务）
export const API_MIGRATION_MAP = {
  // Event Store API (4002)
  '/api/events': SERVICES.eventStore + '/api/events',
  '/api/quality-checks': SERVICES.eventStore + '/api/quality-checks',

  // Task Scheduler API (4003)
  '/api/tasks': SERVICES.taskScheduler + '/api/tasks',
  '/api/task': SERVICES.taskScheduler + '/api/task',

  // Resource Manager API (4006) - 原 /api/resource-pool
  '/api/resource-pool': SERVICES.resourceManager + '/api/resources',
  '/api/resources': SERVICES.resourceManager + '/api/resources',
  '/api/quota-policies': SERVICES.resourceManager + '/api/quota-policies',
  '/api/allocations': SERVICES.resourceManager + '/api/allocations',
  '/api/testbeds': SERVICES.resourceManager + '/api/testbeds',
  '/api/deployments': SERVICES.resourceManager + '/api/deployments',

  // AI Analyzer API (4005)
  '/api/analyze': SERVICES.aiAnalyzer + '/api/analyze',
  '/api/config/ai': SERVICES.aiAnalyzer + '/api/config',
  '/api/config/ai-concurrency': SERVICES.aiAnalyzer + '/api/config/concurrency',
  '/api/config/ai-request-pool-size': SERVICES.aiAnalyzer + '/api/config/pool-size',

  // Executor Service API (4004)
  '/api/execute': SERVICES.executorService + '/api/execute',
  '/api/status': SERVICES.executorService + '/api/status',
  '/api/logs': SERVICES.executorService + '/api/logs',

  // Webhook Gateway API (4001)
  '/api/webhook': SERVICES.webhookGateway + '/api/webhook',
  '/api/hooks': SERVICES.webhookGateway + '/api/hooks'
}

// 获取实际的 API URL（兼容旧路径）
export function getApiUrl(oldPath) {
  // 精确匹配
  if (API_MIGRATION_MAP[oldPath]) {
    return API_MIGRATION_MAP[oldPath]
  }

  // 前缀匹配
  for (const [prefix, target] of Object.entries(API_MIGRATION_MAP)) {
    if (oldPath.startsWith(prefix + '/')) {
      return oldPath.replace(prefix, target.slice(0, -prefix.length))
    }
  }

  // 默认返回原路径
  return oldPath
}

// 兼容旧代码的 API_BASE 常量
export const API_BASE = '/api/resources'

// 各服务模块导出
export const EventStoreAPI = {
  base: () => SERVICES.eventStore,
  events: () => SERVICES.eventStore + '/api/events',
  event: (id) => SERVICES.eventStore + `/api/events/${id}`,
  qualityChecks: (eventId) => SERVICES.eventStore + `/api/events/${eventId}/quality-checks`
}

export const TaskSchedulerAPI = {
  base: () => SERVICES.taskScheduler,
  tasks: () => SERVICES.taskScheduler + '/api/tasks',
  task: (id) => SERVICES.taskScheduler + `/api/tasks/${id}`,
  startTask: (id) => SERVICES.taskScheduler + `/api/tasks/${id}/start`,
  completeTask: (id) => SERVICES.taskScheduler + `/api/tasks/${id}/complete`,
  failTask: (id) => SERVICES.taskScheduler + `/api/tasks/${id}/fail`,
  cancelTask: (id) => SERVICES.taskScheduler + `/api/tasks/${id}/cancel`,
  cancelEvent: (eventId) => SERVICES.taskScheduler + `/api/events/${eventId}/cancel`
}

export const ResourceManagerAPI = {
  base: () => SERVICES.resourceManager,
  resources: () => SERVICES.resourceManager + '/api/resources',
  resource: (id) => SERVICES.resourceManager + `/api/resources/${id}`,
  quotaPolicies: () => SERVICES.resourceManager + '/api/quota-policies',
  allocations: () => SERVICES.resourceManager + '/api/allocations',
  testbeds: () => SERVICES.resourceManager + '/api/testbeds',
  deployments: () => SERVICES.resourceManager + '/api/deployments',
  match: () => SERVICES.resourceManager + '/api/resources/match'
}

export const AIAnalyzerAPI = {
  base: () => SERVICES.aiAnalyzer,
  analyze: () => SERVICES.aiAnalyzer + '/api/analyze',
  batchAnalyze: () => SERVICES.aiAnalyzer + '/api/analyze/batch',
  config: () => SERVICES.aiAnalyzer + '/api/config',
  poolSize: () => SERVICES.aiAnalyzer + '/api/config/pool-size'
}

export const ExecutorServiceAPI = {
  base: () => SERVICES.executorService,
  execute: () => SERVICES.executorService + '/api/execute',
  status: (buildId) => SERVICES.executorService + `/api/status/${buildId}`,
  logs: (buildId) => SERVICES.executorService + `/api/logs/${buildId}`
}

export const WebhookGatewayAPI = {
  base: () => SERVICES.webhookGateway,
  webhook: () => SERVICES.webhookGateway + '/api/webhook',
  hooks: () => SERVICES.webhookGateway + '/api/hooks'
}

// 导出默认配置
export default {
  SERVICES,
  API_MIGRATION_MAP,
  getApiUrl,
  EventStoreAPI,
  TaskSchedulerAPI,
  ResourceManagerAPI,
  AIAnalyzerAPI,
  ExecutorServiceAPI,
  WebhookGatewayAPI
}
