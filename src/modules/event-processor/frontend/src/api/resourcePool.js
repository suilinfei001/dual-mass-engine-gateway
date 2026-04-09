/**
 * Resource Pool API Service
 * API client for the Resource Pool Management System
 * 支持 Mock 模式用于前端开发测试
 */

const API_BASE = '/api/resource-pool'

// Mock API 检查
let mockAPI = null

async function getMockAPI() {
  if (mockAPI) return mockAPI
  if (isMockMode()) {
    const module = await import('../mock/resourcePoolMock.js')
    mockAPI = module.default
    return mockAPI
  }
  return null
}

function isMockMode() {
  return localStorage.getItem('USE_MOCK_API') === 'true' ||
         window.location.search.includes('mock=true') ||
         window.location.hash.includes('mock=true')
}

/**
 * Helper function to handle API responses
 */
async function handleResponse(response) {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: response.statusText }))
    throw new Error(error.message || `HTTP ${response.status}`)
  }
  return response.json()
}

/**
 * Internal API (no authentication required)
 */
export const internalAPI = {
  /**
   * Get all testbeds
   */
  async getTestbeds(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getTestbeds(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/internal/testbeds${query ? `?${query}` : ''}`)
    return handleResponse(response)
  },

  /**
   * Get testbed by UUID
   */
  async getTestbed(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.getTestbed(uuid)

    const response = await fetch(`${API_BASE}/internal/testbeds/${uuid}`)
    return handleResponse(response)
  },

  /**
   * List available testbeds
   */
  async getAvailableTestbeds(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getAvailableTestbeds(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/internal/testbeds/available${query ? `?${query}` : ''}`)
    return handleResponse(response)
  }
}

/**
 * External API (requires session authentication)
 */
export const externalAPI = {
  /**
   * Acquire a testbed
   */
  async acquireTestbed(categoryUUID, options = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.acquireTestbed(categoryUUID, options)

    const response = await fetch(`${API_BASE}/external/allocations`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ category_uuid: categoryUUID, ...options })
    })
    return handleResponse(response)
  },

  /**
   * Get my allocations
   */
  async getMyAllocations(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getMyAllocations(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/external/allocations${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get allocation by UUID
   */
  async getAllocation(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.getAllocation(uuid)

    const response = await fetch(`${API_BASE}/external/allocations/${uuid}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Extend allocation
   */
  async extendAllocation(uuid, extendSeconds) {
    const mock = await getMockAPI()
    if (mock) return mock.extendAllocation(uuid, extendSeconds)

    const response = await fetch(`${API_BASE}/external/allocations/${uuid}/extend`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ extend_seconds: extendSeconds })
    })
    return handleResponse(response)
  },

  /**
   * Release testbed
   */
  async releaseTestbed(allocationUUID) {
    const mock = await getMockAPI()
    if (mock) return mock.releaseTestbed(allocationUUID)

    const response = await fetch(`${API_BASE}/external/allocations/${allocationUUID}`, {
      method: 'DELETE',
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get all categories
   */
  async getCategories() {
    const mock = await getMockAPI()
    if (mock) return mock.getCategories()

    const response = await fetch(`${API_BASE}/external/categories`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get category by UUID
   */
  async getCategory(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.getCategory(uuid)

    const response = await fetch(`${API_BASE}/external/categories/${uuid}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get quota policy for category
   */
  async getQuotaPolicy(categoryUUID) {
    const mock = await getMockAPI()
    if (mock) return mock.getQuotaPolicy(categoryUUID)

    const response = await fetch(`${API_BASE}/external/categories/${categoryUUID}/quota`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get testbed by UUID (all users can view)
   */
  async getTestbed(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.getTestbed(uuid)

    const response = await fetch(`${API_BASE}/external/testbeds/${uuid}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get my resource instances
   */
  async getMyResourceInstances(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getResourceInstances(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/external/resource-instances/my${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get public resource instances (visible to all users)
   */
  async getPublicResourceInstances(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getResourceInstances(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/external/resource-instances/public${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  }
}

/**
 * Admin API (requires admin session)
 */
export const adminAPI = {
  /**
   * Get all testbeds (admin view)
   */
  async getAllTestbeds(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getTestbeds(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/admin/testbeds${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Create testbed
   */
  async createTestbed(data) {
    const mock = await getMockAPI()
    if (mock) {
      // Mock 创建 testbed 实际上是调用部署接口
      return mock.deployToResourceInstance(data.resource_instance_uuid, data)
    }

    const response = await fetch(`${API_BASE}/admin/testbeds`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * Update testbed
   */
  async updateTestbed(uuid, data) {
    const mock = await getMockAPI()
    if (mock) return mock.updateTestbed(uuid, data)

    const response = await fetch(`${API_BASE}/admin/testbeds/${uuid}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * Delete testbed
   */
  async deleteTestbed(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.deleteTestbed(uuid)

    const response = await fetch(`${API_BASE}/admin/testbeds/${uuid}`, {
      method: 'DELETE',
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get all resource instances
   */
  async getResourceInstances(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getResourceInstances(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/admin/resource-instances${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Create resource instance
   */
  async createResourceInstance(data) {
    const mock = await getMockAPI()
    if (mock) return mock.createResourceInstance(data)

    const response = await fetch(`${API_BASE}/admin/resource-instances`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * Update resource instance
   */
  async updateResourceInstance(uuid, data) {
    const mock = await getMockAPI()
    if (mock) return mock.updateResourceInstance(uuid, data)

    const response = await fetch(`${API_BASE}/admin/resource-instances/${uuid}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * Delete resource instance
   */
  async deleteResourceInstance(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.deleteResourceInstance(uuid)

    const response = await fetch(`${API_BASE}/admin/resource-instances/${uuid}`, {
      method: 'DELETE',
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Check resource instance health
   */
  async checkResourceHealth(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.checkResourceHealth(uuid)

    const response = await fetch(`${API_BASE}/admin/resource-instances/${uuid}/health`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Test resource instance connection before creation
   */
  async testConnection(data) {
    const mock = await getMockAPI()
    if (mock) return mock.testConnection(data)

    const response = await fetch(`${API_BASE}/admin/resource-instances/test-connection`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * Get all categories
   */
  async getCategories() {
    const mock = await getMockAPI()
    if (mock) return mock.getCategories()

    const response = await fetch(`${API_BASE}/admin/categories`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Create category
   */
  async createCategory(data) {
    const mock = await getMockAPI()
    if (mock) return mock.createCategory(data)

    const response = await fetch(`${API_BASE}/admin/categories`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * Update category
   */
  async updateCategory(uuid, data) {
    const mock = await getMockAPI()
    if (mock) return mock.updateCategory(uuid, data)

    const response = await fetch(`${API_BASE}/admin/categories/${uuid}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * Delete category
   */
  async deleteCategory(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.deleteCategory(uuid)

    const response = await fetch(`${API_BASE}/admin/categories/${uuid}`, {
      method: 'DELETE',
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get all quota policies
   */
  async getQuotaPolicies() {
    const mock = await getMockAPI()
    if (mock) return mock.getQuotaPolicies()

    const response = await fetch(`${API_BASE}/admin/quota-policies`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get quota policy by category
   */
  async getQuotaPolicy(categoryUUID) {
    const mock = await getMockAPI()
    if (mock) return mock.getQuotaPolicy(categoryUUID)

    const response = await fetch(`${API_BASE}/admin/quota-policies/by-category/${categoryUUID}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Create quota policy
   */
  async createQuotaPolicy(data) {
    const mock = await getMockAPI()
    if (mock) return mock.createQuotaPolicy(data)

    const response = await fetch(`${API_BASE}/admin/quota-policies`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * Update quota policy
   */
  async updateQuotaPolicy(uuid, data) {
    const mock = await getMockAPI()
    if (mock) return mock.updateQuotaPolicy(uuid, data)

    const response = await fetch(`${API_BASE}/admin/quota-policies/${uuid}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * Delete quota policy
   */
  async deleteQuotaPolicy(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.deleteQuotaPolicy(uuid)

    const response = await fetch(`${API_BASE}/admin/quota-policies/${uuid}`, {
      method: 'DELETE',
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get all allocations
   */
  async getAllAllocations(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getAllAllocations(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/admin/allocations${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get allocation history
   */
  async getAllocationHistory(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getAllocationHistory(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/admin/allocations/history${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get metrics
   */
  async getMetrics() {
    const mock = await getMockAPI()
    if (mock) return mock.getMetrics()

    const response = await fetch(`${API_BASE}/admin/metrics`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Get usage statistics
   */
  async getUsageStats(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getUsageStats(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/admin/metrics/usage${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * Deploy to resource instance (部署产品到 ResourceInstance)
   * 模拟部署过程，5秒后返回结果
   */
  async deployToResourceInstance(resourceInstanceUUID, options = {}) {
    const mock = await getMockAPI()
    if (mock) {
      return mock.deployToResourceInstance(resourceInstanceUUID, options)
    }

    // 实际 API 调用
    const response = await fetch(`${API_BASE}/admin/resource-instances/${resourceInstanceUUID}/deploy`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(options)
    })
    return handleResponse(response)
  },

  // ==================== Task Management API ====================

  /**
   * 获取资源实例的任务列表
   */
  async getResourceInstanceTasks(resourceUUID, params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getResourceInstanceTasks(resourceUUID, params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/admin/resource-instances/${resourceUUID}/tasks${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 获取所有任务列表
   */
  async getTasks(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getTasks(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/admin/tasks${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 获取任务详情
   */
  async getTask(taskUUID) {
    const mock = await getMockAPI()
    if (mock) return mock.getTask(taskUUID)

    const response = await fetch(`${API_BASE}/admin/tasks/${taskUUID}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 获取任务统计信息
   */
  async getTaskStatistics() {
    const mock = await getMockAPI()
    if (mock) return mock.getTaskStatistics()

    const response = await fetch(`${API_BASE}/admin/tasks/statistics`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 获取失败任务列表
   */
  async getFailedTasks(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.getFailedTasks(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/admin/tasks/failed${query ? `?${query}` : ''}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 清理旧任务
   */
  async cleanupOldTasks(params = {}) {
    const mock = await getMockAPI()
    if (mock) return mock.cleanupOldTasks(params)

    const query = new URLSearchParams(params).toString()
    const response = await fetch(`${API_BASE}/admin/tasks/cleanup${query ? `?${query}` : ''}`, {
      method: 'POST',
      credentials: 'include'
    })
    return handleResponse(response)
  },

  // ==================== Pipeline Template Management API ====================

  /**
   * 获取所有部署管道模板
   */
  async getPipelineTemplates() {
    const response = await fetch(`${API_BASE}/admin/pipeline-templates`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 获取单个部署管道模板
   */
  async getPipelineTemplate(id) {
    const response = await fetch(`${API_BASE}/admin/pipeline-templates/${id}`, {
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 创建部署管道模板
   */
  async createPipelineTemplate(data) {
    const response = await fetch(`${API_BASE}/admin/pipeline-templates`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * 更新部署管道模板
   */
  async updatePipelineTemplate(id, data) {
    const response = await fetch(`${API_BASE}/admin/pipeline-templates/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(data)
    })
    return handleResponse(response)
  },

  /**
   * 删除部署管道模板
   */
  async deletePipelineTemplate(id) {
    const response = await fetch(`${API_BASE}/admin/pipeline-templates/${id}`, {
      method: 'DELETE',
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 启用部署管道模板
   */
  async enablePipelineTemplate(id) {
    const response = await fetch(`${API_BASE}/admin/pipeline-templates/${id}/enable`, {
      method: 'POST',
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 禁用部署管道模板
   */
  async disablePipelineTemplate(id) {
    const response = await fetch(`${API_BASE}/admin/pipeline-templates/${id}/disable`, {
      method: 'POST',
      credentials: 'include'
    })
    return handleResponse(response)
  },

  /**
   * 快照回滚 - 将虚拟机回滚到指定快照
   */
  async restoreSnapshot(uuid) {
    const mock = await getMockAPI()
    if (mock) return mock.restoreSnapshot(uuid)

    const response = await fetch(`${API_BASE}/admin/resource-instances/${uuid}/restore-snapshot`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include'
    })
    return handleResponse(response)
  }
}

/**
 * 启用 Mock 模式
 */
export function enableMockMode() {
  localStorage.setItem('USE_MOCK_API', 'true')
  console.log('[Resource Pool API] Mock mode enabled')
  // 刷新页面以应用更改
  window.location.reload()
}

/**
 * 禁用 Mock 模式
 */
export function disableMockMode() {
  localStorage.removeItem('USE_MOCK_API')
  console.log('[Resource Pool API] Mock mode disabled')
  // 刷新页面以应用更改
  window.location.reload()
}

/**
 * 检查是否处于 Mock 模式
 */
export { isMockMode }

export default {
  internal: internalAPI,
  external: externalAPI,
  admin: adminAPI,
  enableMockMode,
  disableMockMode,
  isMockMode
}
