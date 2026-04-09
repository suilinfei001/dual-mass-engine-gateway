<template>
  <div class="quota-policy-page">
    <div class="page-header">
      <div class="header-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      </div>
      <div class="header-content">
        <h1>配额策略管理</h1>
        <p>配置用户和组的资源配额</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal">
        <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        创建配额策略
      </button>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="policies.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      </div>
      <h3>还没有配置任何配额策略</h3>
      <p>创建配额策略来控制资源分配</p>
      <button class="btn btn-primary" @click="openCreateModal">
        创建配额策略
      </button>
    </div>

    <div v-else class="table-container">
      <table class="data-table">
        <thead>
          <tr>
            <th>类别</th>
            <th>服务对象</th>
            <th>实例限制</th>
            <th>默认时长</th>
            <th>自动补充</th>
            <th>优先级</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="policy in policies" :key="policy.uuid">
            <td :title="policy.category_name || '-'">
              <div class="category-cell">
                <div class="category-icon">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                  </svg>
                </div>
                <span>{{ policy.category_name || '-' }}</span>
              </div>
            </td>
            <td :title="formatServiceTarget(policy.service_target)">
              <span class="badge" :class="policy.service_target === 'robot' ? 'badge-blue' : 'badge-green'">
                {{ formatServiceTarget(policy.service_target) }}
              </span>
            </td>
            <td :title="`${policy.min_instances} ~ ${policy.max_instances}`">
              <div class="range-cell">
                <span class="range-min">{{ policy.min_instances }}</span>
                <span class="range-separator">~</span>
                <span class="range-max">{{ policy.max_instances }}</span>
              </div>
            </td>
            <td :title="formatDuration(policy.max_lifetime_seconds)">{{ formatDuration(policy.max_lifetime_seconds) }}</td>
            <td :title="policy.auto_replenish ? '开启' : '关闭'">
              <div class="auto-replenish-cell">
                <span class="status-dot" :class="policy.auto_replenish ? 'active' : 'inactive'"></span>
                <span>{{ policy.auto_replenish ? '开启' : '关闭' }}</span>
                <span v-if="policy.auto_replenish" class="threshold">阈值: {{ policy.replenish_threshold }}</span>
              </div>
            </td>
            <td :title="policy.priority">
              <span class="priority-badge">{{ policy.priority }}</span>
            </td>
            <td>
              <div class="action-buttons">
                <button class="btn-icon" @click="editPolicy(policy)" title="编辑">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </button>
                <button v-if="isAdmin" class="btn-icon btn-icon-danger" @click="deletePolicy(policy)" title="删除">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showEditModal" class="modal-overlay" @click.self="showEditModal = false">
      <div class="modal modal-large">
        <div class="modal-header">
          <div class="modal-icon">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
            </svg>
          </div>
          <h3>编辑配额策略</h3>
          <span class="modal-subtitle">{{ currentPolicy?.category_name }}</span>
          <button class="modal-close" @click="showEditModal = false">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <form @submit.prevent="savePolicy">
          <div class="modal-body">
            <div class="form-row">
              <div class="form-group">
                <label for="serviceTarget">服务对象 <span class="required">*</span></label>
                <select id="serviceTarget" v-model="formData.service_target" class="form-control" required>
                  <option value="robot">Robot</option>
                  <option value="normal">普通用户</option>
                </select>
                <small>该配额策略适用的服务对象类型</small>
              </div>
              <div class="form-group">
                <label for="priority">优先级</label>
                <input id="priority" v-model.number="formData.priority" type="number" min="0" class="form-control" />
                <small>数字越小优先级越高，0为最高优先级</small>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="minInstances">最小实例数</label>
                <input id="minInstances" v-model.number="formData.min_instances" type="number" min="0" class="form-control" />
                <small>最小保持的可用实例数</small>
              </div>
              <div class="form-group">
                <label for="maxInstances">最大实例数 <span class="required">*</span></label>
                <input id="maxInstances" v-model.number="formData.max_instances" type="number" min="1" class="form-control" required />
                <small>该类别的最大实例数</small>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="maxLifetime">默认使用时长 (分钟) <span class="required">*</span></label>
                <input id="maxLifetime" v-model.number="formData.max_lifetime_minutes" type="number" min="5" class="form-control" required />
                <small>分配的默认有效期，超过此时间将自动回收（最少5分钟）</small>
              </div>
              <div class="form-group"></div>
            </div>

            <div class="form-section">
              <h4>自动补充设置</h4>
              <div class="toggle-group">
                <label class="toggle">
                  <input v-model="formData.auto_replenish" type="checkbox" />
                  <span class="toggle-slider"></span>
                </label>
                <div class="toggle-info">
                  <span class="toggle-label">启用自动补充</span>
                  <small>当可用 testbed 数量低于阈值时自动补充</small>
                </div>
              </div>

              <div v-if="formData.auto_replenish" class="form-row">
                <div class="form-group">
                  <label for="replenishThreshold">补充阈值 <span class="required">*</span></label>
                  <input id="replenishThreshold" v-model.number="formData.replenish_threshold" type="number" min="1" class="form-control" required />
                  <small>当可用数量低于此值时触发补充</small>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showEditModal = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              <span v-if="submitting" class="spinner-small"></span>
              {{ submitting ? '保存中...' : '保存' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal modal-large">
        <div class="modal-header">
          <div class="modal-icon">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
            </svg>
          </div>
          <h3>创建配额策略</h3>
          <button class="modal-close" @click="showCreateModal = false">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <form @submit.prevent="createPolicy">
          <div class="modal-body">
            <div class="form-group">
              <label for="category">选择类别 <span class="required">*</span></label>
              <select id="category" v-model="formData.category_uuid" class="form-control" required>
                <option value="">选择类别</option>
                <option v-for="category in categories" :key="category.uuid" :value="category.uuid">
                  {{ category.name }}
                </option>
              </select>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="serviceTargetCreate">服务对象 <span class="required">*</span></label>
                <select id="serviceTargetCreate" v-model="formData.service_target" class="form-control" required>
                  <option value="robot">Robot</option>
                  <option value="normal">普通用户</option>
                </select>
                <small>该配额策略适用的服务对象类型</small>
              </div>
              <div class="form-group">
                <label for="priorityCreate">优先级</label>
                <input id="priorityCreate" v-model.number="formData.priority" type="number" min="0" class="form-control" />
                <small>数字越小优先级越高，0为最高优先级</small>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="minInstancesCreate">最小实例数</label>
                <input id="minInstancesCreate" v-model.number="formData.min_instances" type="number" min="0" class="form-control" />
                <small>最小保持的可用实例数</small>
              </div>
              <div class="form-group">
                <label for="maxInstancesCreate">最大实例数 <span class="required">*</span></label>
                <input id="maxInstancesCreate" v-model.number="formData.max_instances" type="number" min="1" class="form-control" required />
                <small>该类别的最大实例数</small>
              </div>
            </div>

            <div class="form-row">
              <div class="form-group">
                <label for="maxLifetimeCreate">默认使用时长 (分钟) <span class="required">*</span></label>
                <input id="maxLifetimeCreate" v-model.number="formData.max_lifetime_minutes" type="number" min="5" class="form-control" required />
                <small>分配的默认有效期，超过此时间将自动回收（最少5分钟）</small>
              </div>
              <div class="form-group"></div>
            </div>

            <div class="form-section">
              <h4>自动补充设置</h4>
              <div class="toggle-group">
                <label class="toggle">
                  <input v-model="formData.auto_replenish" type="checkbox" />
                  <span class="toggle-slider"></span>
                </label>
                <div class="toggle-info">
                  <span class="toggle-label">启用自动补充</span>
                  <small>当可用 testbed 数量低于阈值时自动补充</small>
                </div>
              </div>

              <div v-if="formData.auto_replenish" class="form-row">
                <div class="form-group">
                  <label for="replenishThresholdCreate">补充阈值 <span class="required">*</span></label>
                  <input id="replenishThresholdCreate" v-model.number="formData.replenish_threshold" type="number" min="1" class="form-control" required />
                  <small>当可用数量低于此值时触发补充</small>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="showCreateModal = false">取消</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              <span v-if="submitting" class="spinner-small"></span>
              {{ submitting ? '创建中...' : '创建' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { adminAPI } from '../../api/resourcePool'
import { useDialog } from '../../composables/useDialog'

export default {
  name: 'QuotaPolicy',
  setup() {
    const dialog = useDialog()

    const policies = ref([])
    const categories = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const showEditModal = ref(false)
    const showCreateModal = ref(false)
    const currentPolicy = ref(null)

    const isAdmin = computed(() => localStorage.getItem('userRole') === 'admin')

    const defaultFormData = {
      category_uuid: '',
      service_target: 'normal',
      min_instances: 0,
      max_instances: 10,
      max_lifetime_minutes: 60,
      auto_replenish: false,
      replenish_threshold: 5,
      priority: 100
    }

    const formData = ref({ ...defaultFormData })

    const openCreateModal = () => {
      formData.value = { ...defaultFormData }
      showCreateModal.value = true
    }

    const fetchPolicies = async () => {
      loading.value = true
      try {
        const data = await adminAPI.getQuotaPolicies()
        const policyList = data.data || data.policies || []

        const catData = await adminAPI.getCategories()
        const categoryList = catData.data || catData.categories || []
        categories.value = categoryList

        policies.value = policyList.map(policy => {
          const category = categoryList.find(c => c.uuid === policy.category_uuid)
          return {
            ...policy,
            category_name: category?.name || null
          }
        })
      } catch (error) {
        console.error('Failed to fetch policies:', error)
        dialog.alertError('获取配额策略失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    const editPolicy = (policy) => {
      currentPolicy.value = policy
      formData.value = {
        service_target: policy.service_target || 'normal',
        min_instances: policy.min_instances ?? 0,
        max_instances: policy.max_instances ?? 10,
        max_lifetime_minutes: Math.floor((policy.max_lifetime_seconds ?? 3600) / 60),
        auto_replenish: policy.auto_replenish ?? false,
        replenish_threshold: policy.replenish_threshold ?? 5,
        priority: policy.priority ?? 100
      }
      showEditModal.value = true
    }

    const savePolicy = async () => {
      submitting.value = true
      try {
        const updateData = {
          service_target: formData.value.service_target,
          min_instances: formData.value.min_instances,
          max_instances: formData.value.max_instances,
          max_lifetime_seconds: formData.value.max_lifetime_minutes * 60,
          auto_replenish: formData.value.auto_replenish,
          replenish_threshold: formData.value.auto_replenish ? formData.value.replenish_threshold : 0,
          priority: formData.value.priority
        }

        await adminAPI.updateQuotaPolicy(currentPolicy.value.uuid, updateData)
        dialog.alertSuccess('更新成功')
        showEditModal.value = false
        fetchPolicies()
      } catch (error) {
        console.error('Failed to save policy:', error)
        dialog.alertError('保存失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }

    const createPolicy = async () => {
      submitting.value = true
      try {
        const createData = {
          category_uuid: formData.value.category_uuid,
          service_target: formData.value.service_target,
          min_instances: formData.value.min_instances || 0,
          max_instances: formData.value.max_instances,
          max_lifetime_seconds: formData.value.max_lifetime_minutes * 60,
          auto_replenish: formData.value.auto_replenish,
          replenish_threshold: formData.value.auto_replenish ? formData.value.replenish_threshold : 0,
          priority: formData.value.priority
        }

        await adminAPI.createQuotaPolicy(createData)
        dialog.alertSuccess('创建成功')
        showCreateModal.value = false
        fetchPolicies()
      } catch (error) {
        console.error('Failed to create policy:', error)
        dialog.alertError('创建失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }

    const formatDuration = (seconds) => {
      if (!seconds) return '-'
      const minutes = Math.floor(seconds / 60)
      if (minutes >= 60) {
        const hours = Math.floor(minutes / 60)
        const remainingMinutes = minutes % 60
        if (remainingMinutes > 0) {
          return `${hours} 小时 ${remainingMinutes} 分钟`
        }
        return `${hours} 小时`
      }
      return `${minutes} 分钟`
    }

    const formatServiceTarget = (target) => {
      if (target === 'robot') return 'Robot'
      if (target === 'normal') return '普通用户'
      return '-'
    }

    const deletePolicy = async (policy) => {
      const confirmed = await dialog.confirm(`确定要删除配额策略吗？\n\n类别: ${policy.category_name}\n服务对象: ${formatServiceTarget(policy.service_target)}`)
      if (!confirmed) {
        return
      }
      try {
        await adminAPI.deleteQuotaPolicy(policy.uuid)
        dialog.alertSuccess('删除成功')
        fetchPolicies()
      } catch (error) {
        console.error('Failed to delete policy:', error)
        dialog.alertError('删除失败: ' + error.message)
      }
    }

    onMounted(fetchPolicies)

    return {
      policies,
      categories,
      loading,
      submitting,
      showEditModal,
      showCreateModal,
      currentPolicy,
      formData,
      openCreateModal,
      editPolicy,
      savePolicy,
      createPolicy,
      deletePolicy,
      formatDuration,
      formatServiceTarget,
      isAdmin
    }
  }
}
</script>

<style scoped>
.quota-policy-page {
  max-width: 1200px;
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
  background: linear-gradient(135deg, #DB2777 0%, #F472B6 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(219, 39, 119, 0.25);
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
  background: linear-gradient(135deg, #FDF2F8 0%, #FCE7F3 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.empty-icon svg {
  width: 32px;
  height: 32px;
  color: #DB2777;
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
  margin: 0 0 1.5rem;
}

.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  overflow: hidden;
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

.category-cell {
  display: flex;
  align-items: center;
  gap: 0.625rem;
}

.category-icon {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #0891B2 0%, #22D3EE 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.category-icon svg {
  width: 16px;
  height: 16px;
  color: white;
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 6px;
}

.badge-blue {
  background: #EFF6FF;
  color: #2563EB;
}

.badge-green {
  background: #ECFDF5;
  color: #059669;
}

.range-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-family: 'Monaco', 'Menlo', monospace;
}

.range-min {
  color: #059669;
  font-weight: 600;
}

.range-separator {
  color: #94A3B8;
}

.range-max {
  color: #EA580C;
  font-weight: 600;
}

.auto-replenish-cell {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 0.375rem;
}

.status-dot.active {
  background: #10B981;
  box-shadow: 0 0 0 2px rgba(16, 185, 129, 0.2);
}

.status-dot.inactive {
  background: #94A3B8;
}

.threshold {
  font-size: 0.6875rem;
  color: #64748B;
}

.priority-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 32px;
  padding: 0.25rem 0.5rem;
  background: #F5F3FF;
  color: #7C3AED;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 6px;
}

.action-buttons {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  gap: 0.375rem;
  align-items: center;
}

.btn-icon {
  width: 32px;
  height: 32px;
  border: none;
  background: #F8FAFC;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  color: #64748B;
}

.btn-icon svg {
  width: 16px;
  height: 16px;
}

.btn-icon:hover {
  background: #7C3AED;
  color: white;
}

.btn-icon-danger:hover {
  background: #EF4444;
  color: white;
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
  width: 100%;
  max-width: 480px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.modal-large {
  max-width: 720px;
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #E2E8F0;
}

.modal-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #DB2777 0%, #F472B6 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-icon svg {
  width: 20px;
  height: 20px;
  color: white;
}

.modal-header h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
}

.modal-subtitle {
  font-size: 0.875rem;
  color: #64748B;
  margin-left: auto;
}

.modal-close {
  width: 32px;
  height: 32px;
  border: none;
  background: #F1F5F9;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  color: #64748B;
}

.modal-close:hover {
  background: #E2E8F0;
  color: #1E293B;
}

.modal-close svg {
  width: 18px;
  height: 18px;
}

.modal-body {
  padding: 1.5rem;
  max-height: 60vh;
  overflow-y: auto;
}

.form-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group:last-child {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.5rem;
}

.form-group small {
  display: block;
  font-size: 0.75rem;
  color: #94A3B8;
  margin-top: 0.375rem;
}

.required {
  color: #EF4444;
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

.form-control::placeholder {
  color: #94A3B8;
}

.form-section {
  margin-top: 1.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid #E2E8F0;
}

.form-section h4 {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 1rem;
}

.toggle-group {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.toggle {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
  flex-shrink: 0;
}

.toggle input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #E2E8F0;
  transition: 0.2s;
  border-radius: 24px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: 0.2s;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.toggle input:checked + .toggle-slider {
  background: linear-gradient(135deg, #7C3AED 0%, #A78BFA 100%);
}

.toggle input:checked + .toggle-slider:before {
  transform: translateX(20px);
}

.toggle-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.toggle-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #1E293B;
}

.toggle-info small {
  font-size: 0.75rem;
  color: #94A3B8;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  background: #F8FAFC;
  border-top: 1px solid #E2E8F0;
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

.btn-primary {
  background: linear-gradient(135deg, #7C3AED 0%, #A78BFA 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.25);
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.35);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
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

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .modal {
    margin: 1rem;
    max-width: calc(100% - 2rem);
  }

  .modal-large {
    max-width: calc(100% - 2rem);
  }

  .data-table {
    display: block;
    overflow-x: auto;
  }
}
</style>
