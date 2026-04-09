<template>
  <div class="my-testbed-page">
    <div class="page-header">
      <div class="header-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
        </svg>
      </div>
      <div class="header-content">
        <h1>我申请的 Testbed</h1>
        <p>查看和管理您申请的测试环境</p>
      </div>
    </div>

    <div class="content-container">
      <div class="tabs-container">
        <button class="tab" :class="{ active: activeTab === 'active' }" @click="activeTab = 'active'">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          使用中
          <span class="tab-count">{{ activeCount }}</span>
        </button>
        <button class="tab" :class="{ active: activeTab === 'released' }" @click="activeTab = 'released'">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
          </svg>
          已释放
          <span class="tab-count">{{ releasedCount }}</span>
        </button>
        <button class="tab" :class="{ active: activeTab === 'expired' }" @click="activeTab = 'expired'">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          已过期
          <span class="tab-count">{{ expiredCount }}</span>
        </button>
      </div>

      <div v-if="loading" class="loading-container">
        <div class="spinner"></div>
        <span>加载中...</span>
      </div>

      <div v-else-if="filteredAllocations.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
          </svg>
        </div>
        <h3>{{ emptyTitle }}</h3>
        <p>{{ emptyMessage }}</p>
      </div>

      <div v-else class="allocations-grid">
        <div v-for="item in filteredAllocations" :key="item.uuid" class="allocation-card">
          <div class="card-header">
            <div class="card-title-row">
              <div class="testbed-icon">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
                </svg>
              </div>
              <div class="card-title-info">
                <h3>{{ item.testbed?.name || 'Testbed' }}</h3>
                <span class="category-name">{{ item.category_name || '-' }}</span>
              </div>
            </div>
            <span class="status-badge" :class="getStatusClass(item.status)">
              <span class="status-dot"></span>
              {{ getStatusLabel(item.status) }}
            </span>
          </div>

          <div class="card-body">
            <div class="info-grid">
              <div class="info-item">
                <span class="info-label">服务对象</span>
                <span class="info-value">{{ formatServiceTarget(item.testbed?.service_target) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">申请时间</span>
                <span class="info-value">{{ formatTime(item.created_at) }}</span>
              </div>
              <div class="info-item">
                <span class="info-label">过期时间</span>
                <span class="info-value" :class="{ 'text-danger': isExpired(item) }">
                  {{ formatTime(item.expires_at) }}
                </span>
              </div>
            </div>

            <div v-if="item.status === 'active'" class="time-progress">
              <div class="progress-header">
                <span>使用进度</span>
                <span class="time-remaining">{{ getTimeRemaining(item) }}</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: getTimeProgress(item) + '%' }" :class="getProgressClass(item)"></div>
              </div>
            </div>

            <div v-if="item.testbed && item.status === 'active'" class="connection-section">
              <div class="section-title">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                连接信息
              </div>
              <div class="connection-grid">
                <div class="connection-item">
                  <span class="connection-label">主机</span>
                  <span class="connection-value">{{ item.testbed.host }}</span>
                </div>
                <div class="connection-item">
                  <span class="connection-label">端口</span>
                  <span class="connection-value">{{ item.testbed.db_port }}</span>
                </div>
                <div class="connection-item">
                  <span class="connection-label">用户</span>
                  <span class="connection-value">{{ item.testbed.db_user }}</span>
                </div>
                <div class="connection-item">
                  <span class="connection-label">密码</span>
                  <span class="connection-value password">{{ item.testbed.db_password }}</span>
                </div>
              </div>
            </div>
          </div>

          <div class="card-footer">
            <button v-if="item.status === 'active'" class="action-btn action-btn-extend" @click="extendAllocation(item)">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              延期
            </button>
            <button v-if="item.status === 'active'" class="action-btn action-btn-danger" @click="releaseAllocation(item)">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
              释放
            </button>
            <button class="action-btn action-btn-info" @click="viewDetail(item)">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              详情
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showExtendModal" class="modal-overlay" @click.self="showExtendModal = false">
      <div class="modal modal-small">
        <div class="modal-header">
          <div class="modal-header-content">
            <div class="modal-icon">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3>延期 Testbed</h3>
          </div>
          <button class="modal-close" @click="showExtendModal = false">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <form @submit.prevent="confirmExtend" class="modal-form">
          <div class="extend-info">
            <div class="info-row">
              <span class="info-label">当前过期时间</span>
              <span class="info-value">{{ formatTime(currentAllocation?.expires_at) }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">最大可延期</span>
              <span class="info-value highlight">{{ getMaxExtendTime() }} 分钟</span>
            </div>
          </div>

          <div class="form-group">
            <label for="extendTime">延期时长 (分钟) <span class="required">*</span></label>
            <input
              id="extendTime"
              v-model.number="extendForm.minutes"
              type="number"
              min="1"
              :max="getMaxExtendTime()"
              class="form-control"
              required
            />
          </div>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showExtendModal = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="extending">
              {{ extending ? '延期中...' : '确认延期' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'
import { externalAPI } from '../../api/resourcePool'
import { useDialog } from '../../composables/useDialog'

export default {
  name: 'MyTestbedList',
  setup() {
    const dialog = useDialog()
    const allocations = ref([])
    const categories = ref([])
    const loading = ref(false)
    const extending = ref(false)
    const showExtendModal = ref(false)
    const activeTab = ref('active')

    const extendForm = ref({ minutes: 30 })
    const currentAllocation = ref(null)
    const selectedCategoryQuota = ref(null)

    const isExpiredByTime = (allocation) => {
      if (!allocation.expires_at) return false
      return new Date(allocation.expires_at) < new Date()
    }

    const activeCount = computed(() => {
      return allocations.value.filter(a => a.status === 'active' && !isExpiredByTime(a)).length
    })

    const releasedCount = computed(() => {
      return allocations.value.filter(a => a.status === 'released').length
    })

    const expiredCount = computed(() => {
      return allocations.value.filter(a => a.status === 'expired' || (a.status === 'active' && isExpiredByTime(a))).length
    })

    const filteredAllocations = computed(() => {
      switch (activeTab.value) {
        case 'active':
          return allocations.value.filter(a => a.status === 'active' && !isExpiredByTime(a))
        case 'released':
          return allocations.value.filter(a => a.status === 'released')
        case 'expired':
          return allocations.value.filter(a => a.status === 'expired' || (a.status === 'active' && isExpiredByTime(a)))
        default:
          return allocations.value
      }
    })

    const emptyTitle = computed(() => {
      switch (activeTab.value) {
        case 'active': return '没有使用中的 Testbed'
        case 'released': return '没有已释放记录'
        case 'expired': return '没有过期记录'
        default: return '暂无数据'
      }
    })

    const emptyMessage = computed(() => {
      switch (activeTab.value) {
        case 'active': return '您目前没有使用中的 Testbed，可以前往资源池申请'
        case 'released': return '您还没有释放过任何 Testbed'
        case 'expired': return '您还没有过期记录'
        default: return '暂无数据'
      }
    })

    const fetchAllocations = async () => {
      loading.value = true
      try {
        const data = await externalAPI.getMyAllocations()
        allocations.value = data.data || data.allocations || []
      } catch (error) {
        console.error('Failed to fetch allocations:', error)
        dialog.alertError('获取分配列表失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    const fetchCategories = async () => {
      try {
        const data = await externalAPI.getCategories()
        categories.value = data.data || data.categories || []
      } catch (error) {
        console.error('Failed to fetch categories:', error)
      }
    }

    const releaseAllocation = async (allocation) => {
      const testbedName = allocation.testbed?.name || allocation.testbed_uuid?.substring(0, 8) || 'Testbed'
      const confirmed = await dialog.confirm(`确定要释放 "${testbedName}" 吗？`)
      if (!confirmed) return
      try {
        await externalAPI.releaseTestbed(allocation.uuid)
        dialog.alertSuccess('释放成功')
        fetchAllocations()
      } catch (error) {
        dialog.alertError('释放失败: ' + error.message)
      }
    }

    const extendAllocation = (allocation) => {
      currentAllocation.value = allocation
      extendForm.value.minutes = Math.min(30, getMaxExtendTime())
      showExtendModal.value = true
    }

    const confirmExtend = async () => {
      extending.value = true
      try {
        await externalAPI.extendAllocation(currentAllocation.value.uuid, extendForm.value.minutes * 60)
        dialog.alertSuccess('延期成功')
        showExtendModal.value = false
        fetchAllocations()
      } catch (error) {
        dialog.alertError('延期失败: ' + error.message)
      } finally {
        extending.value = false
      }
    }

    const viewDetail = (allocation) => {
      window.location.href = `/resource-pool/testbeds/${allocation.testbed_uuid}`
    }

    const getMaxExtendTime = () => {
      if (currentAllocation.value?.can_extend_until) {
        const until = new Date(currentAllocation.value.can_extend_until)
        const now = new Date()
        const diff = Math.floor((until - now) / 1000 / 60)
        return Math.max(0, diff)
      }
      return 30
    }

    const isExpired = (allocation) => {
      return new Date(allocation.expires_at) < new Date()
    }

    const getTimeProgress = (allocation) => {
      const created = new Date(allocation.created_at).getTime()
      const expires = new Date(allocation.expires_at).getTime()
      const now = Date.now()
      const total = expires - created
      const elapsed = now - created
      return Math.min(100, Math.max(0, (elapsed / total) * 100))
    }

    const getProgressClass = (allocation) => {
      const progress = getTimeProgress(allocation)
      if (progress > 90) return 'progress-danger'
      if (progress > 70) return 'progress-warning'
      return 'progress-success'
    }

    const getTimeRemaining = (allocation) => {
      const expires = new Date(allocation.expires_at)
      const now = new Date()
      const diff = expires - now
      if (diff <= 0) return '已过期'
      const hours = Math.floor(diff / (1000 * 60 * 60))
      const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
      if (hours > 0) return `剩余 ${hours}小时${minutes}分钟`
      return `剩余 ${minutes}分钟`
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

    const formatServiceTarget = (serviceTarget) => {
      const labels = { 'robot': 'Robot', 'normal': '普通用户' }
      return labels[serviceTarget] || '-'
    }

    const formatTime = (time) => {
      if (!time) return '-'
      return new Date(time).toLocaleString('zh-CN')
    }

    onMounted(() => {
      fetchAllocations()
      fetchCategories()
    })

    return {
      allocations, categories, loading, extending, showExtendModal, activeTab,
      extendForm, currentAllocation, selectedCategoryQuota,
      activeCount, releasedCount, expiredCount, filteredAllocations,
      emptyTitle, emptyMessage,
      releaseAllocation, extendAllocation, confirmExtend, viewDetail,
      getMaxExtendTime, isExpired, getTimeProgress, getProgressClass,
      getTimeRemaining, getStatusClass, getStatusLabel,
      formatServiceTarget, formatTime
    }
  }
}
</script>

<style scoped>
.my-testbed-page {
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

.content-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  overflow: hidden;
}

.tabs-container {
  display: flex;
  gap: 0.25rem;
  padding: 1rem 1.25rem;
  background: #F8FAFC;
  border-bottom: 1px solid #E2E8F0;
}

.tab {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #64748B;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab svg {
  width: 16px;
  height: 16px;
}

.tab:hover {
  background: #F1F5F9;
  color: #475569;
}

.tab.active {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.25);
}

.tab-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 0.375rem;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.2);
}

.tab:not(.active) .tab-count {
  background: #E2E8F0;
  color: #64748B;
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

.allocations-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;
  padding: 1.25rem;
}

.allocation-card {
  background: #FAFAFC;
  border: 1px solid #E2E8F0;
  border-radius: 12px;
  overflow: hidden;
  transition: all 0.2s ease;
}

.allocation-card:hover {
  border-color: #CBD5E1;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  background: white;
  border-bottom: 1px solid #E2E8F0;
}

.card-title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.testbed-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #7C3AED 0%, #A78BFA 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.testbed-icon svg {
  width: 20px;
  height: 20px;
  color: white;
}

.card-title-info h3 {
  font-size: 0.9375rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
}

.category-name {
  font-size: 0.75rem;
  color: #64748B;
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

.card-body {
  padding: 1.25rem;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-label {
  font-size: 0.75rem;
  color: #64748B;
  font-weight: 500;
}

.info-value {
  font-size: 0.875rem;
  color: #1E293B;
}

.text-danger {
  color: #DC2626;
}

.time-progress {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px dashed #E2E8F0;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
  font-size: 0.75rem;
  color: #64748B;
}

.time-remaining {
  font-weight: 600;
  color: #7C3AED;
}

.progress-bar {
  height: 6px;
  background: #E2E8F0;
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  transition: width 0.3s ease;
}

.progress-success {
  background: linear-gradient(90deg, #10B981 0%, #34D399 100%);
}

.progress-warning {
  background: linear-gradient(90deg, #F59E0B 0%, #FBBF24 100%);
}

.progress-danger {
  background: linear-gradient(90deg, #EF4444 0%, #F87171 100%);
}

.connection-section {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px dashed #E2E8F0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: #64748B;
  margin-bottom: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.section-title svg {
  width: 14px;
  height: 14px;
}

.connection-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.5rem;
}

.connection-item {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.connection-label {
  font-size: 0.6875rem;
  color: #94A3B8;
}

.connection-value {
  font-size: 0.8125rem;
  font-family: 'Monaco', 'Menlo', monospace;
  color: #1E293B;
}

.connection-value.password {
  background: #F1F5F9;
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
}

.card-footer {
  display: flex;
  gap: 0.5rem;
  padding: 1rem 1.25rem;
  background: white;
  border-top: 1px solid #E2E8F0;
}

.action-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
  font-weight: 500;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-btn svg {
  width: 14px;
  height: 14px;
}

.action-btn-extend {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  color: white;
}

.action-btn-extend:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.3);
}

.action-btn-danger {
  background: #FEF2F2;
  color: #DC2626;
}

.action-btn-danger:hover {
  background: #FEE2E2;
}

.action-btn-info {
  background: #EFF6FF;
  color: #2563EB;
}

.action-btn-info:hover {
  background: #DBEAFE;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
  max-height: 90vh;
  overflow: hidden;
}

.modal-small {
  width: 420px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  background: linear-gradient(135deg, #F5F3FF 0%, #EDE9FE 100%);
  border-bottom: 1px solid #E2E8F0;
}

.modal-header-content {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.modal-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-icon svg {
  width: 18px;
  height: 18px;
  color: white;
}

.modal-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
}

.modal-close {
  width: 32px;
  height: 32px;
  border: none;
  background: white;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.modal-close svg {
  width: 18px;
  height: 18px;
  color: #64748B;
}

.modal-close:hover {
  background: #F1F5F9;
}

.modal-form {
  padding: 1.5rem;
}

.extend-info {
  background: #F8FAFC;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1.25rem;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
}

.info-row:not(:last-child) {
  border-bottom: 1px solid #E2E8F0;
}

.info-label {
  font-size: 0.875rem;
  color: #64748B;
}

.info-value {
  font-size: 0.875rem;
  font-weight: 500;
  color: #1E293B;
}

.info-value.highlight {
  color: #7C3AED;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  font-size: 0.8125rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.375rem;
}

.required {
  color: #DC2626;
}

.form-control {
  width: 100%;
  padding: 0.625rem 0.875rem;
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

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding-top: 1rem;
  border-top: 1px solid #E2E8F0;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.25);
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.35);
}

.btn-secondary {
  background: #F1F5F9;
  color: #64748B;
  border: 1px solid #E2E8F0;
}

.btn-secondary:hover {
  background: #E2E8F0;
  color: #475569;
}

@media (max-width: 1024px) {
  .allocations-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .tabs-container {
    flex-wrap: wrap;
  }

  .tab {
    flex: 1;
    min-width: 100px;
    justify-content: center;
  }

  .modal-small {
    width: 95%;
    max-width: 95%;
  }
}
</style>
