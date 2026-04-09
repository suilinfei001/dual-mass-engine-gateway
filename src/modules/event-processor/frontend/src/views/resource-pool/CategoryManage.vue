<template>
  <div class="category-manage-page">
    <div class="page-header">
      <div class="header-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
        </svg>
      </div>
      <div class="header-content">
        <h1>类别管理</h1>
        <p>管理资源类别和标签</p>
      </div>
      <button class="btn btn-primary" @click="showCreateModal = true">
        <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        创建类别
      </button>
    </div>

    <div v-if="loading" class="loading-container">
      <div class="spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="categories.length === 0" class="empty-state">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
        </svg>
      </div>
      <h3>还没有创建任何类别</h3>
      <p>创建类别来组织和管理您的资源</p>
      <button class="btn btn-primary" @click="showCreateModal = true">
        创建第一个类别
      </button>
    </div>

    <div v-else class="categories-grid">
      <div
        v-for="category in categories"
        :key="category.uuid"
        class="category-card"
      >
        <div class="card-header">
          <div class="card-icon">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
            </svg>
          </div>
          <h3>{{ category.name }}</h3>
          <div class="card-actions">
            <button class="btn-icon" @click="editCategory(category)" title="编辑">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
            </button>
            <button class="btn-icon btn-icon-danger" @click="deleteCategory(category)" title="删除">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </button>
          </div>
        </div>

        <div class="card-body">
          <p class="card-description">{{ category.description || '暂无描述' }}</p>

          <div class="stats-grid">
            <div class="stat-item stat-available">
              <span class="stat-value">{{ category.available_count || 0 }}</span>
              <span class="stat-label">可用</span>
            </div>
            <div class="stat-item stat-allocated">
              <span class="stat-value">{{ category.allocated_count || 0 }}</span>
              <span class="stat-label">已分配</span>
            </div>
            <div class="stat-item stat-total">
              <span class="stat-value">{{ category.total_count || 0 }}</span>
              <span class="stat-label">总计</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showCreateModal || showEditModal" class="modal-overlay" @click.self="closeModals">
      <div class="modal">
        <div class="modal-header">
          <div class="modal-icon">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
            </svg>
          </div>
          <h3>{{ showEditModal ? '编辑类别' : '创建类别' }}</h3>
          <button class="modal-close" @click="closeModals">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <form @submit.prevent="saveCategory">
          <div class="modal-body">
            <div class="form-group">
              <label for="name">类别名称 <span class="required">*</span></label>
              <input
                id="name"
                v-model="formData.name"
                type="text"
                class="form-control"
                placeholder="例如: MySQL 测试环境"
                required
              />
            </div>
            <div class="form-group">
              <label for="description">描述</label>
              <textarea
                id="description"
                v-model="formData.description"
                class="form-control"
                rows="3"
                placeholder="简要描述该类别的用途..."
              ></textarea>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeModals">
              取消
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              <span v-if="submitting" class="spinner-small"></span>
              {{ submitting ? '保存中...' : '保存' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { adminAPI } from '../../api/resourcePool'
import { useDialog } from '../../composables/useDialog'

export default {
  name: 'CategoryManage',
  setup() {
    const dialog = useDialog()

    const categories = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const showCreateModal = ref(false)
    const showEditModal = ref(false)
    const editingCategory = ref(null)

    const formData = ref({
      name: '',
      description: ''
    })

    const fetchCategories = async () => {
      loading.value = true
      try {
        const data = await adminAPI.getCategories()
        categories.value = data.data || data.categories || []
      } catch (error) {
        console.error('Failed to fetch categories:', error)
        dialog.alertError('获取类别列表失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    const saveCategory = async () => {
      submitting.value = true
      try {
        if (editingCategory.value) {
          await adminAPI.updateCategory(editingCategory.value.uuid, formData.value)
          dialog.alertSuccess('更新成功')
        } else {
          await adminAPI.createCategory(formData.value)
          dialog.alertSuccess('创建成功')
        }
        closeModals()
        fetchCategories()
      } catch (error) {
        console.error('Failed to save category:', error)
        dialog.alertError('保存失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }

    const editCategory = (category) => {
      editingCategory.value = category
      formData.value = {
        name: category.name,
        description: category.description || ''
      }
      showEditModal.value = true
    }

    const deleteCategory = async (category) => {
      const confirmed = await dialog.confirm(`确定要删除类别 "${category.name}" 吗？\n\n注意：如果该类别下还有 Testbed，将无法删除。`)
      if (!confirmed) {
        return
      }
      try {
        await adminAPI.deleteCategory(category.uuid)
        dialog.alertSuccess('删除成功')
        fetchCategories()
      } catch (error) {
        dialog.alertError('删除失败: ' + error.message)
      }
    }

    const closeModals = () => {
      showCreateModal.value = false
      showEditModal.value = false
      editingCategory.value = null
      resetForm()
    }

    const resetForm = () => {
      formData.value = {
        name: '',
        description: ''
      }
    }

    onMounted(fetchCategories)

    return {
      categories,
      loading,
      submitting,
      showCreateModal,
      showEditModal,
      formData,
      saveCategory,
      editCategory,
      deleteCategory,
      closeModals
    }
  }
}
</script>

<style scoped>
.category-manage-page {
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
  background: linear-gradient(135deg, #0891B2 0%, #22D3EE 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(8, 145, 178, 0.25);
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
  margin: 0 0 1.5rem;
}

.categories-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1.25rem;
}

.category-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  overflow: hidden;
  transition: all 0.2s ease;
}

.category-card:hover {
  border-color: #7C3AED;
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.15);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: linear-gradient(135deg, #F8FAFC 0%, #F1F5F9 100%);
  border-bottom: 1px solid #E2E8F0;
}

.card-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #0891B2 0%, #22D3EE 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.card-icon svg {
  width: 18px;
  height: 18px;
  color: white;
}

.card-header h3 {
  flex: 1;
  font-size: 0.9375rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.card-actions {
  display: flex;
  gap: 0.375rem;
}

.btn-icon {
  width: 32px;
  height: 32px;
  border: none;
  background: white;
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

.card-body {
  padding: 1rem;
}

.card-description {
  font-size: 0.8125rem;
  color: #64748B;
  margin: 0 0 1rem;
  line-height: 1.5;
  min-height: 2.4375rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.75rem;
  padding: 1rem;
  background: #F8FAFC;
  border-radius: 8px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
}

.stat-available .stat-value {
  color: #059669;
}

.stat-allocated .stat-value {
  color: #EA580C;
}

.stat-total .stat-value {
  color: #7C3AED;
}

.stat-label {
  font-size: 0.6875rem;
  color: #64748B;
  margin-top: 0.25rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
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
  background: linear-gradient(135deg, #0891B2 0%, #22D3EE 100%);
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
  flex: 1;
  font-size: 1.125rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0;
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

textarea.form-control {
  resize: vertical;
  min-height: 80px;
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

@media (max-width: 1024px) {
  .categories-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 640px) {
  .page-header {
    flex-wrap: wrap;
  }

  .header-content {
    flex-basis: calc(100% - 64px);
  }

  .page-header .btn {
    flex-basis: 100%;
    margin-top: 1rem;
  }

  .categories-grid {
    grid-template-columns: 1fr;
  }

  .modal {
    margin: 1rem;
    max-width: calc(100% - 2rem);
  }
}
</style>
