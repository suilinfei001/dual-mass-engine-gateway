<template>
  <div class="allocation-history-page">
    <div class="page-header">
      <div class="header-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </div>
      <div class="header-content">
        <h1>分配历史</h1>
        <p>查看资源分配记录和使用情况</p>
      </div>
      <button class="btn btn-secondary" @click="exportHistory">
        <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
        </svg>
        导出记录
      </button>
    </div>

    <div class="stats-summary">
      <div class="stat-card">
        <div class="stat-icon stat-icon-purple">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ stats.total }}</span>
          <span class="stat-label">总记录</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon stat-icon-green">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ stats.active }}</span>
          <span class="stat-label">使用中</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon stat-icon-blue">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ stats.released }}</span>
          <span class="stat-label">已释放</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon stat-icon-orange">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ stats.expired }}</span>
          <span class="stat-label">已过期</span>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon stat-icon-cyan">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="stat-info">
          <span class="stat-value">{{ stats.avgDuration }}</span>
          <span class="stat-label">平均时长</span>
        </div>
      </div>
    </div>

    <div class="filters-card">
      <div class="filters">
        <div class="filter-group">
          <label>状态</label>
          <select v-model="filters.status" class="form-control" @change="fetchHistory">
            <option value="">所有状态</option>
            <option value="active">使用中</option>
            <option value="released">已释放</option>
            <option value="expired">已过期</option>
            <option value="cancelled">已取消</option>
          </select>
        </div>
        <div class="filter-group">
          <label>类别</label>
          <select v-model="filters.category" class="form-control" @change="fetchHistory">
            <option value="">所有类别</option>
            <option v-for="cat in categories" :key="cat.uuid" :value="cat.uuid">
              {{ cat.name }}
            </option>
          </select>
        </div>
        <div class="filter-group">
          <label>用户</label>
          <input
            v-model="filters.user"
            type="text"
            class="form-control"
            placeholder="输入用户名..."
            @input="debounceFetch"
          />
        </div>
        <div class="filter-group date-range">
          <label>时间范围</label>
          <div class="date-inputs">
            <input
              v-model="filters.date_from"
              type="date"
              class="form-control"
              @change="fetchHistory"
            />
            <span class="date-separator">至</span>
            <input
              v-model="filters.date_to"
              type="date"
              class="form-control"
              @change="fetchHistory"
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

      <div v-else-if="allocations.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
        </div>
        <h3>没有找到分配记录</h3>
        <p>尝试调整筛选条件</p>
      </div>

      <table v-else class="data-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>用户</th>
            <th>Testbed</th>
            <th>类别</th>
            <th>状态</th>
            <th>申请时间</th>
            <th>过期时间</th>
            <th>使用时长</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="allocation in allocations" :key="allocation.uuid">
            <td :title="shortUUID(allocation.uuid)">
              <span class="uuid-badge">{{ shortUUID(allocation.uuid) }}</span>
            </td>
            <td :title="allocation.allocated_to">
              <div class="user-cell">
                <div class="user-avatar">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                </div>
                <span>{{ allocation.allocated_to }}</span>
              </div>
            </td>
            <td :title="allocation.testbed_name || '-'">
              <router-link
                v-if="allocation.testbed_uuid"
                :to="`/resource-pool/testbeds/${allocation.testbed_uuid}`"
                class="testbed-link"
              >
                {{ allocation.testbed_name || '-' }}
              </router-link>
              <span v-else>-</span>
            </td>
            <td :title="allocation.category_name || '-'">{{ allocation.category_name || '-' }}</td>
            <td :title="getStatusLabel(allocation.status)">
              <span class="status-badge" :class="getStatusClass(allocation.status)">
                <span class="status-dot"></span>
                {{ getStatusLabel(allocation.status) }}
              </span>
            </td>
            <td :title="formatTime(allocation.created_at)">{{ formatTime(allocation.created_at) }}</td>
            <td :title="formatTime(allocation.expires_at)">{{ formatTime(allocation.expires_at) }}</td>
            <td :title="calculateDuration(allocation)">
              <span class="duration">{{ calculateDuration(allocation) }}</span>
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
import { ref, computed, onMounted, inject } from 'vue'
import { adminAPI } from '../../api/resourcePool'

export default {
  name: 'AllocationHistory',
  setup() {
    const dialog = inject('dialog') || {
      alertError: (msg) => console.error(msg)
    }

    const allocations = ref([])
    const categories = ref([])
    const loading = ref(false)

    const filters = ref({
      status: '',
      category: '',
      user: '',
      date_from: '',
      date_to: ''
    })

    const pagination = ref({
      page: 1,
      page_size: 50,
      total: 0
    })

    const stats = ref({
      total: 0,
      active: 0,
      released: 0,
      expired: 0,
      avgDuration: '-'
    })

    const totalPages = computed(() => {
      return Math.ceil(pagination.value.total / pagination.value.page_size)
    })

    let debounceTimer = null

    const fetchHistory = async () => {
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

        const data = await adminAPI.getAllocationHistory(params)
        allocations.value = data.data || data.allocations || []
        pagination.value.total = data.total || allocations.value.length

        updateStats()
      } catch (error) {
        console.error('Failed to fetch history:', error)
        dialog.alertError('获取分配历史失败: ' + error.message)
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

    const updateStats = () => {
      stats.value.total = allocations.value.length
      stats.value.active = allocations.value.filter(a => a.status === 'active').length
      stats.value.released = allocations.value.filter(a => a.status === 'released').length
      stats.value.expired = allocations.value.filter(a => a.status === 'expired').length

      const durations = allocations.value
        .filter(a => a.created_at && a.released_at)
        .map(a => {
          const created = new Date(a.created_at).getTime()
          const released = new Date(a.released_at).getTime()
          return (released - created) / 1000 / 60
        })

      if (durations.length > 0) {
        const avgMinutes = Math.round(durations.reduce((a, b) => a + b, 0) / durations.length)
        if (avgMinutes >= 60) {
          stats.value.avgDuration = `${Math.round(avgMinutes / 60)}h`
        } else {
          stats.value.avgDuration = `${avgMinutes}m`
        }
      }
    }

    const changePage = (page) => {
      pagination.value.page = page
      fetchHistory()
    }

    const debounceFetch = () => {
      clearTimeout(debounceTimer)
      debounceTimer = setTimeout(() => {
        pagination.value.page = 1
        fetchHistory()
      }, 300)
    }

    const exportHistory = () => {
      const headers = ['UUID', '用户', 'Testbed', '类别', '状态', '申请时间', '过期时间', '释放时间', '使用时长']
      const rows = allocations.value.map(a => [
        a.uuid,
        a.allocated_to,
        a.testbed_name || '',
        a.category_name || '',
        getStatusLabel(a.status),
        formatTime(a.created_at),
        formatTime(a.expires_at),
        formatTime(a.released_at),
        calculateDuration(a)
      ])

      const csvContent = [
        headers.join(','),
        ...rows.map(row => row.map(cell => `"${cell}"`).join(','))
      ].join('\n')

      const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' })
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = `allocation_history_${new Date().toISOString().slice(0, 10)}.csv`
      link.click()
    }

    const shortUUID = (uuid) => {
      if (!uuid) return '-'
      return uuid.slice(0, 8)
    }

    const getStatusClass = (status) => {
      const classes = {
        'active': 'status-active',
        'released': 'status-released',
        'expired': 'status-expired',
        'cancelled': 'status-cancelled'
      }
      return classes[status] || 'status-cancelled'
    }

    const getStatusLabel = (status) => {
      const labels = {
        'active': '使用中',
        'released': '已释放',
        'expired': '已过期',
        'cancelled': '已取消'
      }
      return labels[status] || status
    }

    const formatTime = (time) => {
      if (!time) return '-'
      return new Date(time).toLocaleString('zh-CN')
    }

    const calculateDuration = (allocation) => {
      if (!allocation.created_at) return '-'

      const startTime = new Date(allocation.created_at).getTime()
      const endTime = allocation.released_at
        ? new Date(allocation.released_at).getTime()
        : Date.now()

      const diffMinutes = Math.floor((endTime - startTime) / 1000 / 60)

      if (diffMinutes < 60) {
        return `${diffMinutes} 分钟`
      }

      const hours = Math.floor(diffMinutes / 60)
      const minutes = diffMinutes % 60
      return `${hours}h ${minutes}m`
    }

    onMounted(() => {
      fetchHistory()
      fetchCategories()
    })

    return {
      allocations,
      categories,
      loading,
      filters,
      pagination,
      stats,
      totalPages,
      fetchHistory,
      changePage,
      debounceFetch,
      exportHistory,
      shortUUID,
      getStatusClass,
      getStatusLabel,
      formatTime,
      calculateDuration
    }
  }
}
</script>

<style scoped>
.allocation-history-page {
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

.stats-summary {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon svg {
  width: 22px;
  height: 22px;
}

.stat-icon-purple {
  background: linear-gradient(135deg, #F5F3FF 0%, #EDE9FE 100%);
  color: #7C3AED;
}

.stat-icon-green {
  background: linear-gradient(135deg, #ECFDF5 0%, #D1FAE5 100%);
  color: #059669;
}

.stat-icon-blue {
  background: linear-gradient(135deg, #EFF6FF 0%, #DBEAFE 100%);
  color: #2563EB;
}

.stat-icon-orange {
  background: linear-gradient(135deg, #FFF7ED 0%, #FFEDD5 100%);
  color: #EA580C;
}

.stat-icon-cyan {
  background: linear-gradient(135deg, #ECFEFF 0%, #CFFAFE 100%);
  color: #0891B2;
}

.stat-info {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0F172A;
}

.stat-label {
  font-size: 0.75rem;
  color: #64748B;
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

.filter-group .form-control {
  min-width: 140px;
}

.date-range {
  flex: 1;
  min-width: 280px;
}

.date-inputs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.date-separator {
  color: #94A3B8;
  font-size: 0.875rem;
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

.uuid-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.5rem;
  background: #F1F5F9;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.75rem;
  color: #64748B;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.user-avatar {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, #F5F3FF 0%, #EDE9FE 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-avatar svg {
  width: 14px;
  height: 14px;
  color: #7C3AED;
}

.testbed-link {
  color: #7C3AED;
  font-weight: 500;
  text-decoration: none;
  transition: color 0.2s ease;
}

.testbed-link:hover {
  color: #6D28D9;
  text-decoration: underline;
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

.status-active {
  background: #ECFDF5;
  color: #059669;
}

.status-active .status-dot {
  background: #10B981;
}

.status-released {
  background: #EFF6FF;
  color: #2563EB;
}

.status-released .status-dot {
  background: #3B82F6;
}

.status-expired {
  background: #FEF2F2;
  color: #DC2626;
}

.status-expired .status-dot {
  background: #EF4444;
}

.status-cancelled {
  background: #F1F5F9;
  color: #64748B;
}

.status-cancelled .status-dot {
  background: #94A3B8;
}

.duration {
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.8125rem;
  color: #64748B;
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

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-secondary {
  background: white;
  color: #64748B;
  border: 1px solid #E2E8F0;
}

.btn-secondary:hover {
  background: #F8FAFC;
  border-color: #CBD5E1;
}

.btn-icon {
  width: 18px;
  height: 18px;
}

@media (max-width: 1024px) {
  .stats-summary {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .stats-summary {
    grid-template-columns: repeat(2, 1fr);
  }

  .filters {
    flex-direction: column;
  }

  .filter-group {
    width: 100%;
  }

  .date-range {
    min-width: auto;
  }

  .data-table {
    display: block;
    overflow-x: auto;
  }
}
</style>
