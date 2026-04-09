<template>
  <div class="testbed-list-page">
    <div class="page-header">
      <div class="header-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
        </svg>
      </div>
      <div class="header-content">
        <h1>所有 Testbed</h1>
        <p>查看和管理测试环境资源</p>
      </div>
    </div>

    <div class="filters-card">
      <div class="filters">
        <div class="filter-group">
          <label>状态</label>
          <select v-model="filters.status" class="form-control" @change="fetchTestbeds">
            <option value="">所有状态</option>
            <option value="available">可用</option>
            <option value="allocated">已分配</option>
            <option value="in_use">使用中</option>
            <option value="releasing">释放中</option>
            <option value="deleted">已删除</option>
          </select>
        </div>
        <div class="filter-group">
          <label>类别</label>
          <select v-model="filters.category" class="form-control" @change="fetchTestbeds">
            <option value="">所有类别</option>
            <option v-for="cat in categories" :key="cat.uuid" :value="cat.uuid">
              {{ cat.name }}
            </option>
          </select>
        </div>
        <div class="filter-group search-group">
          <label>搜索</label>
          <div class="search-input-wrapper">
            <svg class="search-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="filters.search"
              type="text"
              class="form-control search-input"
              placeholder="搜索名称..."
              @input="debounceFetch"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="table-container">
      <div v-if="loading" class="loading-container">
        <div class="spinner"></div>
        <span>加载中...</span>
      </div>

      <div v-else-if="testbeds.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
          </svg>
        </div>
        <h3>没有找到 Testbed</h3>
        <p>尝试调整筛选条件</p>
      </div>

      <table v-else class="data-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>类别</th>
            <th>服务对象</th>
            <th>状态</th>
            <th>连接信息</th>
            <th>过期时间</th>
            <th>剩余时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="testbed in testbeds" :key="testbed.uuid">
            <td :title="testbed.name">
              <router-link :to="`/resource-pool/testbeds/${testbed.uuid}`" class="testbed-name">
                {{ testbed.name }}
              </router-link>
            </td>
            <td :title="testbed.category_name || '-'">{{ testbed.category_name || '-' }}</td>
            <td :title="formatServiceTarget(testbed.service_target)">
              <span class="service-badge" :class="getServiceTargetClass(testbed.service_target)">
                {{ formatServiceTarget(testbed.service_target) }}
              </span>
            </td>
            <td :title="getStatusLabel(testbed.status)">
              <span class="status-badge" :class="getStatusClass(testbed.status)">
                <span class="status-dot"></span>
                {{ getStatusLabel(testbed.status) }}
              </span>
            </td>
            <td :title="`${testbed.host || '-'}${testbed.ssh_port ? ':' + testbed.ssh_port : ''}`">
              <div class="connection-info">
                <span class="host">{{ testbed.host || '-' }}</span>
                <span v-if="testbed.ssh_port" class="port">:{{ testbed.ssh_port }}</span>
              </div>
            </td>
            <td :title="formatExpiryTime(testbed.expires_at, testbed.status)">{{ formatExpiryTime(testbed.expires_at, testbed.status) }}</td>
            <td :title="formatRemainingTime(testbed.expires_at, testbed.status)">
              <span class="remaining-time" :class="getExpiryClass(testbed.expires_at, testbed.status)">
                {{ formatRemainingTime(testbed.expires_at, testbed.status) }}
              </span>
            </td>
            <td>
              <div class="action-buttons">
                <button class="action-btn action-btn-info" @click="viewTestbed(testbed)" title="查看详情">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </button>
                <button
                  v-if="!isAdmin && testbed.status === 'available' && testbed.service_target !== 'robot'"
                  class="action-btn action-btn-primary"
                  @click="allocateTestbed(testbed)"
                  title="获取"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                  </svg>
                </button>
                <button
                  v-if="isAdmin && testbed.status === 'available'"
                  class="action-btn action-btn-danger"
                  @click="deleteTestbed(testbed)"
                  title="删除"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="pagination.total > 0" class="pagination">
        <button class="pagination-btn" :disabled="pagination.page <= 1" @click="changePage(pagination.page - 1)">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
          上一页
        </button>
        <span class="page-info">第 {{ pagination.page }} 页，共 {{ totalPages }} 页 ({{ pagination.total }} 条)</span>
        <button class="pagination-btn" :disabled="pagination.page >= totalPages" @click="changePage(pagination.page + 1)">
          下一页
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, computed } from 'vue'
import { adminAPI, externalAPI, internalAPI } from '../../api/resourcePool'
import { useDialog } from '../../composables/useDialog'

export default {
  name: 'TestbedList',
  setup() {
    const dialog = useDialog()
    const testbeds = ref([])
    const categories = ref([])
    const loading = ref(false)
    const isAdmin = ref(false)

    const filters = ref({
      status: '',
      category: '',
      search: ''
    })

    const pagination = ref({
      page: 1,
      page_size: 20,
      total: 0
    })

    const totalPages = computed(() => {
      return Math.ceil(pagination.value.total / pagination.value.page_size)
    })

    let debounceTimer = null

    const fetchTestbeds = async () => {
      loading.value = true
      try {
        const params = {
          page: pagination.value.page,
          page_size: pagination.value.page_size,
          ...filters.value
        }
        Object.keys(params).forEach(key => {
          if (params[key] === '' || params[key] === null || params[key] === undefined) {
            delete params[key]
          }
        })

        let data
        if (isAdmin.value) {
          data = await adminAPI.getAllTestbeds(params)
        } else {
          data = await internalAPI.getTestbeds(params)
        }
        let fetchedTestbeds = data.data || data.testbeds || []

        if (!Array.isArray(fetchedTestbeds)) {
          console.warn('API returned non-array data:', data)
          fetchedTestbeds = []
        }

        if (!isAdmin.value) {
          fetchedTestbeds = fetchedTestbeds.filter(t => t.service_target !== 'robot')
        }

        testbeds.value = fetchedTestbeds
        pagination.value.total = data.total || fetchedTestbeds.length
      } catch (error) {
        console.error('Failed to fetch testbeds:', error)
      } finally {
        loading.value = false
      }
    }

    const fetchCategories = async () => {
      try {
        const data = await adminAPI.getCategories()
        categories.value = data.data || data.categories || []
      } catch (error) {
        console.error('Failed to fetch categories:', error)
      }
    }

    const viewTestbed = (testbed) => {
      window.location.href = `/resource-pool/testbeds/${testbed.uuid}`
    }

    const deleteTestbed = async (testbed) => {
      const confirmed = await dialog.confirm(`确定要删除 Testbed "${testbed.name}" 吗？`)
      if (!confirmed) return
      try {
        await adminAPI.updateTestbed(testbed.uuid, { status: 'deleted' })
        fetchTestbeds()
      } catch (error) {
        dialog.alertError('操作失败: ' + error.message)
      }
    }

    const allocateTestbed = async (testbed) => {
      const confirmed = await dialog.confirm(`确定要获取 Testbed "${testbed.name}" 吗？`)
      if (!confirmed) return
      try {
        await externalAPI.acquireTestbed(testbed.category_uuid)
        dialog.alertSuccess('获取成功！请前往"我申请的 Testbed"查看详情。')
        fetchTestbeds()
      } catch (error) {
        dialog.alertError('获取失败: ' + error.message)
      }
    }

    const changePage = (page) => {
      pagination.value.page = page
      fetchTestbeds()
    }

    const debounceFetch = () => {
      clearTimeout(debounceTimer)
      debounceTimer = setTimeout(() => {
        pagination.value.page = 1
        fetchTestbeds()
      }, 300)
    }

    const getStatusClass = (status) => {
      const classes = {
        'available': 'status-available',
        'allocated': 'status-allocated',
        'in_use': 'status-in-use',
        'releasing': 'status-releasing',
        'deleted': 'status-deleted'
      }
      return classes[status] || 'status-releasing'
    }

    const getStatusLabel = (status) => {
      const labels = {
        'available': '可用',
        'allocated': '已分配',
        'in_use': '使用中',
        'releasing': '释放中',
        'deleted': '已删除'
      }
      return labels[status] || status
    }

    const formatServiceTarget = (serviceTarget) => {
      const labels = {
        'robot': 'Robot',
        'normal': '普通用户'
      }
      return labels[serviceTarget] || '-'
    }

    const getServiceTargetClass = (serviceTarget) => {
      return serviceTarget === 'robot' ? 'service-robot' : 'service-normal'
    }

    const formatExpiryTime = (expiresAt, status) => {
      if (!expiresAt) return '-'
      const date = new Date(expiresAt)
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hours = String(date.getHours()).padStart(2, '0')
      const minutes = String(date.getMinutes()).padStart(2, '0')
      return `${month}-${day} ${hours}:${minutes}`
    }

    const formatRemainingTime = (expiresAt, status) => {
      if (!expiresAt) return '-'

      const now = new Date()
      const expiry = new Date(expiresAt)
      const diff = expiry - now

      if (diff <= 0) {
        return '已过期'
      }

      const minutes = Math.floor(diff / 60000)
      const hours = Math.floor(minutes / 60)
      const days = Math.floor(hours / 24)

      if (days > 0) {
        return `${days}天${hours % 24}小时`
      } else if (hours > 0) {
        return `${hours}小时${minutes % 60}分钟`
      } else {
        return `${minutes}分钟`
      }
    }

    const getExpiryClass = (expiresAt, status) => {
      if (!expiresAt) return ''

      const now = new Date()
      const expiry = new Date(expiresAt)
      const diff = expiry - now

      if (diff <= 0) {
        return 'expiry-expired'
      } else if (diff < 30 * 60 * 1000) {
        return 'expiry-warning'
      } else if (diff < 60 * 60 * 1000) {
        return 'expiry-soon'
      }
      return ''
    }

    onMounted(() => {
      isAdmin.value = localStorage.getItem('userRole') === 'admin'
      fetchTestbeds()
      fetchCategories()
    })

    return {
      testbeds, categories, loading, isAdmin, filters, pagination, totalPages,
      fetchTestbeds, viewTestbed, deleteTestbed, allocateTestbed, changePage,
      debounceFetch, getStatusClass, getStatusLabel, formatServiceTarget,
      getServiceTargetClass, formatExpiryTime, formatRemainingTime, getExpiryClass
    }
  }
}
</script>

<style scoped>
.testbed-list-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem 0;
  margin-bottom: 1.5rem;
  border-bottom: 1px solid #E2E8F0;
}

.header-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, #6366F1 0%, #A5B4FC 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.25);
}

.header-icon svg {
  width: 24px;
  height: 24px;
  color: white;
}

.header-content {
  flex: 1;
}

.header-content h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0F172A;
  margin: 0 0 0.25rem;
}

.header-content p {
  font-size: 0.875rem;
  color: #64748B;
  margin: 0;
}

.filters-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  padding: 1.25rem;
  margin-bottom: 1.5rem;
}

.filters {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.filter-group label {
  font-size: 0.75rem;
  font-weight: 500;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.search-group {
  flex: 1;
  min-width: 200px;
}

.search-input-wrapper {
  position: relative;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  width: 18px;
  height: 18px;
  color: #94A3B8;
}

.search-input {
  padding-left: 2.5rem !important;
}

.form-control {
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  background: white;
  transition: all 0.2s ease;
  color: #1E293B;
}

.form-control:focus {
  outline: none;
  border-color: #7C3AED;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
}

.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  overflow: hidden;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 0;
  color: #64748B;
  gap: 1rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #E2E8F0;
  border-top-color: #7C3AED;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 0;
  text-align: center;
}

.empty-icon {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, #F5F3FF 0%, #EDE9FE 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.empty-icon svg {
  width: 32px;
  height: 32px;
  color: #7C3AED;
}

.empty-state h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 0.5rem;
}

.empty-state p {
  font-size: 0.875rem;
  color: #64748B;
  margin: 0;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.data-table th {
  background: #F8FAFC;
  padding: 0.875rem 1rem;
  text-align: left;
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid #E2E8F0;
  white-space: nowrap;
}

.data-table td {
  padding: 1rem;
  border-bottom: 1px solid #F1F5F9;
  font-size: 0.875rem;
  color: #1E293B;
  /* Apply text overflow to all table cells */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Ensure inline elements in table cells also truncate */
.data-table td > span,
.data-table td > a {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

/* Ensure action buttons keep flex layout */
.data-table td > .action-buttons {
  display: flex;
}

.data-table tr:last-child td {
  border-bottom: none;
}

.data-table tr:hover td {
  background: #FAFAFC;
}

.testbed-name {
  color: #7C3AED;
  font-weight: 600;
  text-decoration: none;
  transition: color 0.2s ease;
}

.testbed-name:hover {
  color: #6D28D9;
  text-decoration: underline;
}

.service-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 6px;
}

.service-robot {
  background: #F5F3FF;
  color: #7C3AED;
}

.service-normal {
  background: #EFF6FF;
  color: #2563EB;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 6px;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-available {
  background: #ECFDF5;
  color: #059669;
}

.status-available .status-dot {
  background: #10B981;
}

.status-allocated {
  background: #EFF6FF;
  color: #2563EB;
}

.status-allocated .status-dot {
  background: #3B82F6;
}

.status-in-use {
  background: #FEF3C7;
  color: #D97706;
}

.status-in-use .status-dot {
  background: #F59E0B;
}

.status-releasing {
  background: #F1F5F9;
  color: #64748B;
}

.status-releasing .status-dot {
  background: #94A3B8;
}

.status-deleted {
  background: #FEF2F2;
  color: #DC2626;
}

.status-deleted .status-dot {
  background: #EF4444;
}

.connection-info {
  display: flex;
  align-items: center;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.8125rem;
}

.host {
  font-weight: 500;
  color: #1E293B;
}

.port {
  color: #64748B;
}

.remaining-time {
  font-weight: 500;
}

.expiry-expired {
  color: #DC2626;
}

.expiry-warning {
  color: #EA580C;
}

.expiry-soon {
  color: #D97706;
}

.action-buttons {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  gap: 0.375rem;
  align-items: center;
}

.action-btn {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.action-btn svg {
  width: 16px;
  height: 16px;
}

.action-btn-info {
  background: #EFF6FF;
  color: #2563EB;
}

.action-btn-info:hover {
  background: #DBEAFE;
}

.action-btn-primary {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  color: white;
}

.action-btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.3);
}

.action-btn-danger {
  background: #FEF2F2;
  color: #DC2626;
}

.action-btn-danger:hover {
  background: #FEE2E2;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  border-top: 1px solid #E2E8F0;
}

.pagination-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  background: white;
  color: #64748B;
  cursor: pointer;
  transition: all 0.2s ease;
}

.pagination-btn svg {
  width: 16px;
  height: 16px;
}

.pagination-btn:hover:not(:disabled) {
  background: #F8FAFC;
  border-color: #CBD5E1;
  color: #1E293B;
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.875rem;
  color: #64748B;
}

@media (max-width: 768px) {
  .filters {
    flex-direction: column;
  }

  .filter-group {
    width: 100%;
  }

  .data-table {
    display: block;
    overflow-x: auto;
  }
}
</style>
