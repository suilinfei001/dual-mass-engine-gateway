/**
 * Resource Pool Mock API
 * 模拟 Resource Pool 后端 API，用于前端开发和测试
 */

// 模拟数据存储
const mockData = {
  // 类别
  categories: [
    {
      uuid: 'cat-mysql-001',
      name: 'MySQL 测试环境',
      description: '用于 MySQL 相关测试的数据库环境',
      created_at: '2024-01-01T00:00:00Z',
      available_count: 5,
      allocated_count: 3,
      total_count: 8
    },
    {
      uuid: 'cat-pgsql-001',
      name: 'PostgreSQL 测试环境',
      description: '用于 PostgreSQL 相关测试的数据库环境',
      created_at: '2024-01-01T00:00:00Z',
      available_count: 3,
      allocated_count: 2,
      total_count: 5
    },
    {
      uuid: 'cat-redis-001',
      name: 'Redis 缓存环境',
      description: '用于 Redis 缓存测试的环境',
      created_at: '2024-01-01T00:00:00Z',
      available_count: 4,
      allocated_count: 1,
      total_count: 5
    }
  ],

  // 配额策略
  quotaPolicies: [
    {
      uuid: 'quota-mysql-001',
      category_uuid: 'cat-mysql-001',
      min_instances: 1,
      max_instances: 10,
      max_lifetime_seconds: 3600,
      auto_replenish: true,
      replenish_threshold: 3,
      priority: 100
    },
    {
      uuid: 'quota-pgsql-001',
      category_uuid: 'cat-pgsql-001',
      min_instances: 1,
      max_instances: 5,
      max_lifetime_seconds: 7200,
      auto_replenish: false,
      replenish_threshold: 2,
      priority: 100
    },
    {
      uuid: 'quota-redis-001',
      category_uuid: 'cat-redis-001',
      min_instances: 1,
      max_instances: 5,
      max_lifetime_seconds: 1800,
      auto_replenish: true,
      replenish_threshold: 2,
      priority: 50
    }
  ],

  // 资源实例
  resourceInstances: [
    {
      uuid: 'res-vm-001',
      name: 'vm-testbed-01',
      resource_type: 'virtual_machine',
      host: '192.168.1.101',
      ssh_port: 22,
      snapshot_id: 'snapshot-mysql-v1.0',
      status: 'available',
      created_by: 'admin',
      created_at: '2024-01-01T00:00:00Z',
      testbed_count: 2
    },
    {
      uuid: 'res-vm-002',
      name: 'vm-testbed-02',
      resource_type: 'virtual_machine',
      host: '192.168.1.102',
      ssh_port: 22,
      snapshot_id: 'snapshot-mysql-v1.0',
      status: 'available',
      created_by: 'admin',
      created_at: '2024-01-01T00:00:00Z',
      testbed_count: 1
    },
    {
      uuid: 'res-vm-003',
      name: 'vm-testbed-03',
      resource_type: 'virtual_machine',
      host: '192.168.1.103',
      ssh_port: 22,
      snapshot_id: 'snapshot-pgsql-v1.0',
      status: 'in_use',
      created_by: 'admin',
      created_at: '2024-01-01T00:00:00Z',
      testbed_count: 1
    },
    {
      uuid: 'res-vm-004',
      name: 'vm-testbed-04',
      resource_type: 'virtual_machine',
      host: '192.168.1.104',
      ssh_port: 22,
      snapshot_id: 'snapshot-redis-v1.0',
      status: 'available',
      created_by: 'admin',
      created_at: '2024-01-01T00:00:00Z',
      testbed_count: 1
    }
  ],

  // Testbed
  testbeds: [
    {
      uuid: 'tb-mysql-001',
      name: 'mysql-testbed-01',
      category_uuid: 'cat-mysql-001',
      category_name: 'MySQL 测试环境',
      resource_instance_uuid: 'res-vm-001',
      host: '192.168.1.101',
      ssh_port: 22,
      db_port: 3306,
      db_user: 'root',
      db_password: 'Test@123456',
      status: 'available',
      allocated_to: null,
      created_at: '2024-01-01T00:00:00Z',
      last_health_check_at: new Date().toISOString(),
      resource_instance: {
        uuid: 'res-vm-001',
        name: 'vm-testbed-01',
        resource_type: 'virtual_machine',
        host: '192.168.1.101',
        snapshot_id: 'snapshot-mysql-v1.0',
        status: 'available'
      }
    },
    {
      uuid: 'tb-mysql-002',
      name: 'mysql-testbed-02',
      category_uuid: 'cat-mysql-001',
      category_name: 'MySQL 测试环境',
      resource_instance_uuid: 'res-vm-001',
      host: '192.168.1.101',
      ssh_port: 22,
      db_port: 3307,
      db_user: 'root',
      db_password: 'Test@123456',
      status: 'in_use',
      allocated_to: 'user1',
      created_at: '2024-01-01T00:00:00Z',
      last_health_check_at: new Date().toISOString(),
      resource_instance: {
        uuid: 'res-vm-001',
        name: 'vm-testbed-01',
        resource_type: 'virtual_machine',
        host: '192.168.1.101',
        snapshot_id: 'snapshot-mysql-v1.0',
        status: 'in_use'
      }
    },
    {
      uuid: 'tb-pgsql-001',
      name: 'pgsql-testbed-01',
      category_uuid: 'cat-pgsql-001',
      category_name: 'PostgreSQL 测试环境',
      resource_instance_uuid: 'res-vm-003',
      host: '192.168.1.103',
      ssh_port: 22,
      db_port: 5432,
      db_user: 'postgres',
      db_password: 'Pg@Test@123',
      status: 'available',
      allocated_to: null,
      created_at: '2024-01-01T00:00:00Z',
      last_health_check_at: new Date().toISOString(),
      resource_instance: {
        uuid: 'res-vm-003',
        name: 'vm-testbed-03',
        resource_type: 'virtual_machine',
        host: '192.168.1.103',
        snapshot_id: 'snapshot-pgsql-v1.0',
        status: 'in_use'
      }
    },
    {
      uuid: 'tb-redis-001',
      name: 'redis-testbed-01',
      category_uuid: 'cat-redis-001',
      category_name: 'Redis 缓存环境',
      resource_instance_uuid: 'res-vm-004',
      host: '192.168.1.104',
      ssh_port: 22,
      db_port: 6379,
      db_user: 'root',
      db_password: 'Redis@Test@123',
      status: 'available',
      allocated_to: null,
      created_at: '2024-01-01T00:00:00Z',
      last_health_check_at: new Date().toISOString(),
      resource_instance: {
        uuid: 'res-vm-004',
        name: 'vm-testbed-04',
        resource_type: 'virtual_machine',
        host: '192.168.1.104',
        snapshot_id: 'snapshot-redis-v1.0',
        status: 'available'
      }
    }
  ],

  // 分配记录
  allocations: [
    {
      uuid: 'alloc-001',
      testbed_uuid: 'tb-mysql-002',
      testbed_name: 'mysql-testbed-02',
      category_uuid: 'cat-mysql-001',
      category_name: 'MySQL 测试环境',
      allocated_to: 'user1',
      status: 'active',
      purpose: '测试 MySQL 连接池',
      created_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
      expires_at: new Date(Date.now() + 30 * 60 * 1000).toISOString(),
      released_at: null,
      can_extend_until: new Date(Date.now() + 60 * 60 * 1000).toISOString(),
      testbed: {
        uuid: 'tb-mysql-002',
        name: 'mysql-testbed-02',
        host: '192.168.1.101',
        db_port: 3307,
        db_user: 'root',
        db_password: 'Test@123456'
      }
    }
  ],

  // 资源实例任务
  tasks: [
    {
      id: 1,
      uuid: 'task-deploy-001',
      resource_instance_uuid: 'res-vm-001',
      task_type: 'deploy',
      task_type_name: '部署',
      status: 'completed',
      status_name: '已完成',
      trigger_source: 'manual',
      trigger_source_name: '手动触发',
      trigger_user: 'admin',
      category_uuid: 'cat-mysql-001',
      testbed_uuid: 'tb-mysql-001',
      started_at: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
      completed_at: new Date(Date.now() - 59 * 60 * 1000).toISOString(),
      duration_ms: 45000,
      duration_display: '45秒',
      success: true,
      error_code: null,
      error_message: null,
      result_details: {
        testbed_uuid: 'tb-mysql-001',
        mariadb_port: 3306,
        mariadb_user: 'root',
        product_version: 'v1.0.0'
      },
      retry_count: 0,
      max_retries: 3,
      created_at: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 59 * 60 * 1000).toISOString()
    },
    {
      id: 2,
      uuid: 'task-rollback-001',
      resource_instance_uuid: 'res-vm-001',
      task_type: 'rollback',
      task_type_name: '回滚',
      status: 'completed',
      status_name: '已完成',
      trigger_source: 'allocation_release',
      trigger_source_name: '分配释放',
      trigger_user: null,
      category_uuid: 'cat-mysql-001',
      testbed_uuid: 'tb-mysql-001',
      allocation_uuid: 'alloc-001',
      started_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
      completed_at: new Date(Date.now() - 29 * 60 * 1000).toISOString(),
      duration_ms: 15000,
      duration_display: '15秒',
      success: true,
      error_code: null,
      error_message: null,
      result_details: {},
      retry_count: 0,
      max_retries: 3,
      created_at: new Date(Date.now() - 30 * 60 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 29 * 60 * 1000).toISOString()
    },
    {
      id: 3,
      uuid: 'task-deploy-002',
      resource_instance_uuid: 'res-vm-002',
      task_type: 'deploy',
      task_type_name: '部署',
      status: 'running',
      status_name: '执行中',
      trigger_source: 'auto_replenish',
      trigger_source_name: '自动补充',
      trigger_user: null,
      quota_policy_uuid: 'quota-mysql-001',
      category_uuid: 'cat-mysql-001',
      testbed_uuid: null,
      started_at: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
      completed_at: null,
      duration_ms: null,
      duration_display: null,
      success: null,
      error_code: null,
      error_message: null,
      result_details: {},
      retry_count: 0,
      max_retries: 3,
      created_at: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 5 * 60 * 1000).toISOString()
    },
    {
      id: 4,
      uuid: 'task-deploy-003',
      resource_instance_uuid: 'res-vm-003',
      task_type: 'deploy',
      task_type_name: '部署',
      status: 'failed',
      status_name: '失败',
      trigger_source: 'manual',
      trigger_source_name: '手动触发',
      trigger_user: 'admin',
      category_uuid: 'cat-pgsql-001',
      testbed_uuid: null,
      started_at: new Date(Date.now() - 120 * 60 * 1000).toISOString(),
      completed_at: new Date(Date.now() - 115 * 60 * 1000).toISOString(),
      duration_ms: 300000,
      duration_display: '5分钟',
      success: false,
      error_code: 'DEPLOY_TIMEOUT',
      error_message: '部署超时：无法在规定时间内完成部署',
      result_details: {},
      retry_count: 2,
      max_retries: 3,
      created_at: new Date(Date.now() - 120 * 60 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 115 * 60 * 1000).toISOString()
    },
    {
      id: 5,
      uuid: 'task-health-check-001',
      resource_instance_uuid: 'res-vm-001',
      task_type: 'health_check',
      task_type_name: '健康检查',
      status: 'completed',
      status_name: '已完成',
      trigger_source: 'manual',
      trigger_source_name: '手动触发',
      trigger_user: 'admin',
      testbed_uuid: null,
      started_at: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
      completed_at: new Date(Date.now() - 10 * 60 * 1000 + 500).toISOString(),
      duration_ms: 500,
      duration_display: '0秒',
      success: true,
      error_code: null,
      error_message: null,
      result_details: {
        healthy: true,
        response_time_ms: 200
      },
      retry_count: 0,
      max_retries: 3,
      created_at: new Date(Date.now() - 10 * 60 * 1000).toISOString(),
      updated_at: new Date(Date.now() - 10 * 60 * 1000 + 500).toISOString()
    }
  ]
}

// 模拟延迟 (5秒部署时间)
const DEPLOY_DELAY = 5000

/**
 * 模拟部署产品到 ResourceInstance
 * @param {string} resourceInstanceUUID - 资源实例 UUID
 * @param {object} options - 部署选项
 * @returns {Promise} - 返回部署结果
 */
function mockDeployToResourceInstance(resourceInstanceUUID, options = {}) {
  return new Promise((resolve, reject) => {
    console.log(`[Mock Deploy] Starting deployment to ${resourceInstanceUUID}...`)

    // 模拟部署进度
    let progress = 0
    const progressInterval = setInterval(() => {
      progress += 20
      console.log(`[Mock Deploy] Deployment progress: ${progress}%`)
    }, DEPLOY_DELAY / 5)

    // 5秒后返回成功
    setTimeout(() => {
      clearInterval(progressInterval)

      const resourceInstance = mockData.resourceInstances.find(r => r.uuid === resourceInstanceUUID)

      if (!resourceInstance) {
        reject(new Error('Resource instance not found'))
        return
      }

      // 创建新的 Testbed
      const newTestbed = {
        uuid: `tb-mock-${Date.now()}`,
        name: `${resourceInstance.name}-deployed`,
        category_uuid: options.category_uuid || 'cat-mysql-001',
        category_name: options.category_name || 'MySQL 测试环境',
        resource_instance_uuid: resourceInstanceUUID,
        host: resourceInstance.host,
        ssh_port: resourceInstance.ssh_port,
        db_port: options.db_port || 3306,
        db_user: options.db_user || 'root',
        db_password: options.db_password || 'Default@123456',
        status: 'available',
        allocated_to: null,
        created_at: new Date().toISOString(),
        last_health_check_at: new Date().toISOString(),
        resource_instance: {
          uuid: resourceInstance.uuid,
          name: resourceInstance.name,
          resource_type: resourceInstance.resource_type,
          host: resourceInstance.host,
          snapshot_id: resourceInstance.snapshot_id,
          status: resourceInstance.status
        }
      }

      mockData.testbeds.push(newTestbed)

      console.log(`[Mock Deploy] Deployment completed successfully: ${newTestbed.uuid}`)

      resolve({
        success: true,
        data: newTestbed,
        message: 'Deployment completed successfully'
      })
    }, DEPLOY_DELAY)
  })
}

/**
 * 模拟获取所有 Testbed
 */
function mockGetTestbeds(params = {}) {
  let testbeds = [...mockData.testbeds]

  // 过滤
  if (params.status) {
    testbeds = testbeds.filter(t => t.status === params.status)
  }
  if (params.category) {
    testbeds = testbeds.filter(t => t.category_uuid === params.category)
  }
  if (params.search) {
    const search = params.search.toLowerCase()
    testbeds = testbeds.filter(t =>
      t.name.toLowerCase().includes(search) ||
      t.host.toLowerCase().includes(search)
    )
  }

  // 分页
  const page = params.page || 1
  const pageSize = params.page_size || 20
  const start = (page - 1) * pageSize
  const end = start + pageSize
  const paginatedTestbeds = testbeds.slice(start, end)

  return Promise.resolve({
    data: paginatedTestbeds,
    testbeds: paginatedTestbeds,
    total: testbeds.length,
    page: page,
    page_size: pageSize
  })
}

/**
 * 模拟获取单个 Testbed
 */
function mockGetTestbed(uuid) {
  const testbed = mockData.testbeds.find(t => t.uuid === uuid)
  if (!testbed) {
    return Promise.reject(new Error('Testbed not found'))
  }
  return Promise.resolve({ data: testbed, testbed: testbed })
}

/**
 * 模拟获取可用的 Testbed
 */
function mockGetAvailableTestbeds(params = {}) {
  return mockGetTestbeds({ ...params, status: 'available' })
}

/**
 * 模拟申请 Testbed
 */
function mockAcquireTestbed(categoryUUID, options = {}) {
  return new Promise((resolve, reject) => {
    // 查找可用的 Testbed
    const availableTestbed = mockData.testbeds.find(
      t => t.category_uuid === categoryUUID && t.status === 'available'
    )

    if (!availableTestbed) {
      reject(new Error('No available testbed in this category'))
      return
    }

    // 创建分配记录
    const username = options.username || 'test-user'
    const duration = options.duration_seconds || 3600

    const allocation = {
      uuid: `alloc-${Date.now()}`,
      testbed_uuid: availableTestbed.uuid,
      testbed_name: availableTestbed.name,
      category_uuid: categoryUUID,
      category_name: availableTestbed.category_name,
      allocated_to: username,
      status: 'active',
      purpose: options.purpose || '',
      created_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + duration * 1000).toISOString(),
      released_at: null,
      can_extend_until: new Date(Date.now() + duration * 1.5 * 1000).toISOString(),
      testbed: {
        uuid: availableTestbed.uuid,
        name: availableTestbed.name,
        host: availableTestbed.host,
        db_port: availableTestbed.db_port,
        db_user: availableTestbed.db_user,
        db_password: availableTestbed.db_password
      }
    }

    // 更新 Testbed 状态
    availableTestbed.status = 'in_use'
    availableTestbed.allocated_to = username

    mockData.allocations.push(allocation)

    // 返回分配和 testbed 信息
    resolve({
      data: {
        allocation: allocation,
        testbed: availableTestbed
      },
      allocation: allocation,
      testbed: availableTestbed
    })
  })
}

/**
 * 模拟获取我的分配
 */
function mockGetMyAllocations(params = {}) {
  const username = params.username || 'user1'
  let allocations = mockData.allocations.filter(a => a.allocated_to === username)

  // 过滤
  if (params.status) {
    allocations = allocations.filter(a => a.status === params.status)
  }

  return Promise.resolve({
    data: allocations,
    allocations: allocations,
    total: allocations.length
  })
}

/**
 * 模拟释放 Testbed
 */
function mockReleaseTestbed(allocationUUID) {
  return new Promise((resolve, reject) => {
    const allocation = mockData.allocations.find(a => a.uuid === allocationUUID)
    if (!allocation) {
      reject(new Error('Allocation not found'))
      return
    }

    // 更新状态
    allocation.status = 'released'
    allocation.released_at = new Date().toISOString()

    // 释放 Testbed
    const testbed = mockData.testbeds.find(t => t.uuid === allocation.testbed_uuid)
    if (testbed) {
      testbed.status = 'available'
      testbed.allocated_to = null
    }

    resolve({ success: true, message: 'Testbed released successfully' })
  })
}

/**
 * 模拟延期分配
 */
function mockExtendAllocation(allocationUUID, extendSeconds) {
  return new Promise((resolve, reject) => {
    const allocation = mockData.allocations.find(a => a.uuid === allocationUUID)
    if (!allocation) {
      reject(new Error('Allocation not found'))
      return
    }

    // 延期
    const currentExpires = new Date(allocation.expires_at).getTime()
    const newExpires = currentExpires + extendSeconds * 1000
    allocation.expires_at = new Date(newExpires).toISOString()

    resolve({ success: true, message: 'Allocation extended successfully' })
  })
}

/**
 * 模拟获取所有类别
 */
function mockGetCategories() {
  return Promise.resolve({
    data: mockData.categories,
    categories: mockData.categories
  })
}

/**
 * 模拟获取配额策略
 */
function mockGetQuotaPolicy(categoryUUID) {
  const policy = mockData.quotaPolicies.find(p => p.category_uuid === categoryUUID)
  if (!policy) {
    return Promise.reject(new Error('Quota policy not found'))
  }
  return Promise.resolve({
    data: policy,
    quota_policy: policy
  })
}

/**
 * 模拟获取所有配额策略
 */
function mockGetQuotaPolicies() {
  return Promise.resolve({
    data: mockData.quotaPolicies,
    quota_policies: mockData.quotaPolicies
  })
}

/**
 * 模拟获取资源实例列表
 */
function mockGetResourceInstances(params = {}) {
  let instances = [...mockData.resourceInstances]

  // 过滤
  if (params.status) {
    instances = instances.filter(i => i.status === params.status)
  }
  if (params.type) {
    instances = instances.filter(i => i.resource_type === params.type)
  }
  if (params.search) {
    const search = params.search.toLowerCase()
    instances = instances.filter(i =>
      i.name.toLowerCase().includes(search) ||
      i.host.toLowerCase().includes(search)
    )
  }

  // 分页
  const page = params.page || 1
  const pageSize = params.page_size || 20
  const start = (page - 1) * pageSize
  const end = start + pageSize
  const paginated = instances.slice(start, end)

  return Promise.resolve({
    data: paginated,
    resource_instances: paginated,
    total: instances.length
  })
}

/**
 * 模拟创建资源实例
 */
function mockCreateResourceInstance(data) {
  const newInstance = {
    uuid: `res-${Date.now()}`,
    name: data.name,
    resource_type: data.resource_type,
    host: data.host,
    ssh_port: data.ssh_port || 22,
    snapshot_id: data.snapshot_id || '',
    status: 'available',
    created_by: data.created_by || 'admin',
    created_at: new Date().toISOString(),
    testbed_count: 0
  }
  mockData.resourceInstances.push(newInstance)
  return Promise.resolve({ data: newInstance })
}

/**
 * 模拟更新资源实例
 */
function mockUpdateResourceInstance(uuid, data) {
  const instance = mockData.resourceInstances.find(i => i.uuid === uuid)
  if (!instance) {
    return Promise.reject(new Error('Resource instance not found'))
  }
  Object.assign(instance, data)
  return Promise.resolve({ data: instance })
}

/**
 * 模拟删除资源实例
 */
function mockDeleteResourceInstance(uuid) {
  const index = mockData.resourceInstances.findIndex(i => i.uuid === uuid)
  if (index === -1) {
    return Promise.reject(new Error('Resource instance not found'))
  }
  mockData.resourceInstances.splice(index, 1)
  return Promise.resolve({ success: true })
}

/**
 * 模拟创建类别
 */
function mockCreateCategory(data) {
  const newCategory = {
    uuid: `cat-${Date.now()}`,
    name: data.name,
    description: data.description || '',
    created_at: new Date().toISOString(),
    available_count: 0,
    allocated_count: 0,
    total_count: 0
  }
  mockData.categories.push(newCategory)
  return Promise.resolve({ data: newCategory })
}

/**
 * 模拟更新类别
 */
function mockUpdateCategory(uuid, data) {
  const category = mockData.categories.find(c => c.uuid === uuid)
  if (!category) {
    return Promise.reject(new Error('Category not found'))
  }
  Object.assign(category, data)
  return Promise.resolve({ data: category })
}

/**
 * 模拟删除类别
 */
function mockDeleteCategory(uuid) {
  const index = mockData.categories.findIndex(c => c.uuid === uuid)
  if (index === -1) {
    return Promise.reject(new Error('Category not found'))
  }
  mockData.categories.splice(index, 1)
  return Promise.resolve({ success: true })
}

/**
 * 模拟更新配额策略
 */
function mockUpdateQuotaPolicy(uuid, data) {
  const policy = mockData.quotaPolicies.find(p => p.uuid === uuid)
  if (!policy) {
    return Promise.reject(new Error('Quota policy not found'))
  }
  Object.assign(policy, data)
  return Promise.resolve({ data: policy })
}

/**
 * 模拟获取分配历史
 */
function mockGetAllocationHistory(params = {}) {
  let allocations = [...mockData.allocations]

  // 添加一些历史记录
  const historyAllocations = [
    {
      uuid: 'alloc-hist-001',
      testbed_uuid: 'tb-mysql-001',
      testbed_name: 'mysql-testbed-01',
      category_uuid: 'cat-mysql-001',
      category_name: 'MySQL 测试环境',
      allocated_to: 'user1',
      status: 'released',
      purpose: '测试 MySQL 主从复制',
      created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
      expires_at: new Date(Date.now() - 1 * 60 * 60 * 1000).toISOString(),
      released_at: new Date(Date.now() - 1 * 60 * 60 * 1000).toISOString()
    },
    {
      uuid: 'alloc-hist-002',
      testbed_uuid: 'tb-pgsql-001',
      testbed_name: 'pgsql-testbed-01',
      category_uuid: 'cat-pgsql-001',
      category_name: 'PostgreSQL 测试环境',
      allocated_to: 'user2',
      status: 'expired',
      purpose: 'PostgreSQL 性能测试',
      created_at: new Date(Date.now() - 5 * 60 * 60 * 1000).toISOString(),
      expires_at: new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString(),
      released_at: null
    }
  ]

  allocations = [...allocations, ...historyAllocations]

  // 过滤
  if (params.status) {
    allocations = allocations.filter(a => a.status === params.status)
  }
  if (params.user) {
    allocations = allocations.filter(a => a.allocated_to === params.user)
  }

  return Promise.resolve({
    data: allocations,
    allocations: allocations,
    total: allocations.length
  })
}

/**
 * 模拟获取指标
 */
function mockGetMetrics() {
  return Promise.resolve({
    data: {
      total_testbeds: mockData.testbeds.length,
      available_testbeds: mockData.testbeds.filter(t => t.status === 'available').length,
      active_allocations: mockData.allocations.filter(a => a.status === 'active').length,
      total_users: 5
    },
    metrics: {
      total_testbeds: mockData.testbeds.length,
      available_testbeds: mockData.testbeds.filter(t => t.status === 'available').length,
      active_allocations: mockData.allocations.filter(a => a.status === 'active').length,
      total_users: 5
    }
  })
}

/**
 * 模拟获取使用统计
 */
function mockGetUsageStats(params = {}) {
  return Promise.resolve({
    data: {
      categories: mockData.categories.map(cat => ({
        ...cat,
        available: cat.available_count,
        total: cat.total_count
      })),
      users: [
        {
          username: 'user1',
          current_allocations: 1,
          total_allocations: 5,
          total_duration_seconds: 14400,
          last_used_at: new Date().toISOString()
        },
        {
          username: 'user2',
          current_allocations: 0,
          total_allocations: 3,
          total_duration_seconds: 7200,
          last_used_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString()
        }
      ],
      recent_activity: []
    }
  })
}

/**
 * 模拟获取所有分配（管理员）
 */
function mockGetAllAllocations(params = {}) {
  let allocations = [...mockData.allocations]

  // 过滤
  if (params.status) {
    allocations = allocations.filter(a => a.status === params.status)
  }

  // 分页
  const page = params.page || 1
  const pageSize = params.page_size || 50
  const start = (page - 1) * pageSize
  const end = start + pageSize
  const paginated = allocations.slice(start, end)

  return Promise.resolve({
    data: paginated,
    allocations: paginated,
    total: allocations.length
  })
}

// 导出所有 Mock API
export const mockAPI = {
  // 部署相关
  deployToResourceInstance: mockDeployToResourceInstance,

  // Testbed 相关
  getTestbed: mockGetTestbed,
  getTestbeds: mockGetTestbeds,
  getAvailableTestbeds: mockGetAvailableTestbeds,
  updateTestbed: (uuid, data) => {
    const testbed = mockData.testbeds.find(t => t.uuid === uuid)
    if (!testbed) return Promise.reject(new Error('Testbed not found'))
    Object.assign(testbed, data)
    return Promise.resolve({ data: testbed })
  },
  deleteTestbed: (uuid) => {
    const index = mockData.testbeds.findIndex(t => t.uuid === uuid)
    if (index === -1) return Promise.reject(new Error('Testbed not found'))
    mockData.testbeds.splice(index, 1)
    return Promise.resolve({ success: true })
  },

  // 分配相关
  acquireTestbed: mockAcquireTestbed,
  getMyAllocations: mockGetMyAllocations,
  getAllocation: (uuid) => {
    const allocation = mockData.allocations.find(a => a.uuid === uuid)
    if (!allocation) return Promise.reject(new Error('Allocation not found'))
    return Promise.resolve({ data: allocation })
  },
  extendAllocation: mockExtendAllocation,
  releaseTestbed: mockReleaseTestbed,

  // 类别相关
  getCategories: mockGetCategories,
  getCategory: (uuid) => {
    const category = mockData.categories.find(c => c.uuid === uuid)
    if (!category) return Promise.reject(new Error('Category not found'))
    return Promise.resolve({ data: category })
  },
  createCategory: mockCreateCategory,
  updateCategory: mockUpdateCategory,
  deleteCategory: mockDeleteCategory,

  // 配额相关
  getQuotaPolicy: mockGetQuotaPolicy,
  getQuotaPolicies: mockGetQuotaPolicies,
  updateQuotaPolicy: mockUpdateQuotaPolicy,
  createQuotaPolicy: (data) => {
    const newPolicy = {
      uuid: `quota-${Date.now()}`,
      category_uuid: data.category_uuid,
      min_instances: data.min_instances || 0,
      max_instances: data.max_instances || 10,
      max_lifetime_seconds: data.max_lifetime_seconds || 3600,
      auto_replenish: data.auto_replenish || false,
      replenish_threshold: data.replenish_threshold || 5,
      priority: data.priority || 100
    }
    mockData.quotaPolicies.push(newPolicy)
    return Promise.resolve({ data: newPolicy })
  },
  deleteQuotaPolicy: (uuid) => {
    const index = mockData.quotaPolicies.findIndex(p => p.uuid === uuid)
    if (index === -1) return Promise.reject(new Error('Quota policy not found'))
    mockData.quotaPolicies.splice(index, 1)
    return Promise.resolve({ success: true })
  },

  // 资源实例相关
  getResourceInstances: mockGetResourceInstances,
  createResourceInstance: mockCreateResourceInstance,
  updateResourceInstance: mockUpdateResourceInstance,
  deleteResourceInstance: mockDeleteResourceInstance,

  // 统计相关
  getMetrics: mockGetMetrics,
  getUsageStats: mockGetUsageStats,
  getAllocationHistory: mockGetAllocationHistory,
  getAllAllocations: mockGetAllAllocations,

  // 任务相关
  getResourceInstanceTasks: (resourceUUID, params = {}) => {
    const tasks = mockData.tasks.filter(t => t.resource_instance_uuid === resourceUUID)
    const page = params.page || 1
    const pageSize = params.page_size || 20
    const startIdx = (page - 1) * pageSize
    const endIdx = startIdx + pageSize
    const paginatedTasks = tasks.slice(startIdx, endIdx)
    return Promise.resolve({
      data: paginatedTasks,
      tasks: paginatedTasks,
      total: tasks.length,
      page,
      page_size: pageSize
    })
  },

  getTasks: (params = {}) => {
    let tasks = [...mockData.tasks]
    // 按状态过滤
    if (params.status) {
      tasks = tasks.filter(t => t.status === params.status)
    }
    // 按类型过滤
    if (params.task_type) {
      tasks = tasks.filter(t => t.task_type === params.task_type)
    }
    const page = params.page || 1
    const pageSize = params.page_size || 20
    const startIdx = (page - 1) * pageSize
    const endIdx = startIdx + pageSize
    const paginatedTasks = tasks.slice(startIdx, endIdx)
    return Promise.resolve({
      data: paginatedTasks,
      tasks: paginatedTasks,
      total: tasks.length,
      page,
      page_size: pageSize
    })
  },

  getTask: (taskUUID) => {
    const task = mockData.tasks.find(t => t.uuid === taskUUID)
    if (!task) return Promise.reject(new Error('Task not found'))
    return Promise.resolve({
      data: task,
      task
    })
  },

  getTaskStatistics: () => {
    const stats = {
      total: mockData.tasks.length,
      pending: mockData.tasks.filter(t => t.status === 'pending').length,
      running: mockData.tasks.filter(t => t.status === 'running').length,
      completed: mockData.tasks.filter(t => t.status === 'completed').length,
      failed: mockData.tasks.filter(t => t.status === 'failed').length,
      cancelled: mockData.tasks.filter(t => t.status === 'cancelled').length,
      by_type: {
        deploy: mockData.tasks.filter(t => t.task_type === 'deploy').length,
        rollback: mockData.tasks.filter(t => t.task_type === 'rollback').length,
        health_check: mockData.tasks.filter(t => t.task_type === 'health_check').length
      },
      by_trigger: {
        manual: mockData.tasks.filter(t => t.trigger_source === 'manual').length,
        auto_replenish: mockData.tasks.filter(t => t.trigger_source === 'auto_replenish').length,
        allocation_release: mockData.tasks.filter(t => t.trigger_source === 'allocation_release').length
      },
      average_duration_ms: 120000
    }
    return Promise.resolve({
      data: stats,
      stats
    })
  },

  getFailedTasks: (params = {}) => {
    let tasks = mockData.tasks.filter(t => t.status === 'failed')
    const since = params.since
    if (since) {
      const sinceDate = new Date(since)
      tasks = tasks.filter(t => new Date(t.completed_at) >= sinceDate)
    }
    return Promise.resolve({
      data: tasks,
      tasks,
      total: tasks.length
    })
  },

  cleanupOldTasks: (params = {}) => {
    const days = params.days || 30
    const cutoffDate = new Date(Date.now() - days * 24 * 60 * 60 * 1000)
    const beforeCount = mockData.tasks.length
    mockData.tasks = mockData.tasks.filter(t => {
      if (t.status !== 'completed' && t.status !== 'failed' && t.status !== 'cancelled') {
        return true
      }
      return new Date(t.created_at) >= cutoffDate
    })
    const deletedCount = beforeCount - mockData.tasks.length
    return Promise.resolve({
      success: true,
      message: `Deleted ${deletedCount} old tasks (older than ${days} days)`,
      deleted: deletedCount
    })
  },

  restoreSnapshot: (uuid) => {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve({
          success: true,
          message: '快照回滚已启动（模拟）'
        })
      }, 1000)
    })
  }
}

// 检测是否使用 Mock 模式
export const isMockMode = () => {
  return localStorage.getItem('USE_MOCK_API') === 'true' || window.location.search.includes('mock=true')
}

// 导出 Mock 数据
export { mockData }
export default mockAPI
