<template>
  <div class="pipeline-template-page">
    <div class="page-header">
      <div class="header-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
        </svg>
      </div>
      <div class="header-content">
        <h1>模块部署管道模板配置</h1>
        <p>管理 Azure DevOps 部署管道模板</p>
      </div>
      <button class="btn btn-primary" @click="showCreateModal = true">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        创建模板
      </button>
    </div>

    <div class="card">
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <span>加载中...</span>
      </div>

      <div v-else-if="templates.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
          </svg>
        </div>
        <h3>还没有创建任何部署管道模板</h3>
        <p>创建模板以配置自动化部署流程</p>
        <button class="btn btn-primary" @click="showCreateModal = true">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          创建第一个模板
        </button>
      </div>

      <div v-else class="templates-table-wrapper">
        <table class="templates-table">
          <thead>
            <tr>
              <th>模板名称</th>
              <th>组织 / 项目</th>
              <th>Pipeline ID</th>
              <th>流水线参数</th>
              <th>描述</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="template in templates" :key="template.id">
              <td :title="template.name">
                <div class="template-name-cell">
                  <svg class="template-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
                  </svg>
                  <span class="template-name">{{ template.name }}</span>
                </div>
              </td>
              <td :title="`${template.organization} / ${template.project}`">
                <span class="org-project-cell">{{ template.organization }} <span class="separator">/</span> {{ template.project }}</span>
              </td>
              <td :title="template.pipeline_id">
                <span class="pipeline-id">{{ template.pipeline_id }}</span>
              </td>
              <td :title="formatParams(template.pipeline_parameters)">
                <div class="params-cell" v-if="template.pipeline_parameters && Object.keys(template.pipeline_parameters).length > 0">
                  <span class="params-preview">{{ formatParams(template.pipeline_parameters) }}</span>
                  <span class="params-count">{{ Object.keys(template.pipeline_parameters).length }} 个参数</span>
                </div>
                <span v-else class="no-params">-</span>
              </td>
              <td :title="template.description || '-'">
                <span class="description">{{ template.description || '-' }}</span>
              </td>
              <td :title="template.enabled ? '启用' : '禁用'">
                <span :class="['status-badge', template.enabled ? 'status-enabled' : 'status-disabled']">
                  <span class="status-dot"></span>
                  {{ template.enabled ? '启用' : '禁用' }}
                </span>
              </td>
              <td>
                <div class="actions">
                  <button class="btn-icon" @click="editTemplate(template)" title="编辑">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    v-if="template.enabled"
                    class="btn-icon btn-icon-warning"
                    @click="toggleTemplate(template)"
                    title="禁用"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                    </svg>
                  </button>
                  <button
                    v-else
                    class="btn-icon btn-icon-success"
                    @click="toggleTemplate(template)"
                    title="启用"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </button>
                  <button class="btn-icon btn-icon-danger" @click="deleteTemplate(template)" title="删除">
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
    </div>

    <div v-if="showCreateModal || showEditModal" class="modal-overlay" @click.self="closeModals">
      <div class="modal">
        <div class="modal-header">
          <div class="modal-title-row">
            <div class="modal-icon">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
            </div>
            <h3 class="modal-title">{{ showEditModal ? '编辑部署管道模板' : '创建部署管道模板' }}</h3>
          </div>
          <button class="modal-close" @click="closeModals">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <form @submit.prevent="saveTemplate">
          <div class="modal-body">
            <div class="form-section">
              <h4 class="section-title">基本信息</h4>
              <div class="form-group">
                <label for="name">模板名称 <span class="required">*</span></label>
                <input
                  id="name"
                  v-model="formData.name"
                  type="text"
                  class="form-control"
                  placeholder="例如: event-processor-dev"
                  required
                />
              </div>
              <div class="form-group">
                <label for="description">描述</label>
                <textarea
                  id="description"
                  v-model="formData.description"
                  class="form-control"
                  rows="2"
                  placeholder="简要描述该模板的用途..."
                ></textarea>
              </div>
            </div>

            <div class="form-section">
              <h4 class="section-title">Azure DevOps 配置</h4>
              <div class="form-row">
                <div class="form-group">
                  <label for="organization">Azure DevOps 组织 <span class="required">*</span></label>
                  <input
                    id="organization"
                    v-model="formData.organization"
                    type="text"
                    class="form-control"
                    placeholder="例如: DefaultCollection"
                    required
                  />
                </div>
                <div class="form-group">
                  <label for="project">Azure DevOps 项目 <span class="required">*</span></label>
                  <input
                    id="project"
                    v-model="formData.project"
                    type="text"
                    class="form-control"
                    placeholder="例如: dept-quality-adapter"
                    required
                  />
                </div>
              </div>
              <div class="form-group">
                <label for="pipelineId">Pipeline ID <span class="required">*</span></label>
                <input
                  id="pipelineId"
                  v-model.number="formData.pipeline_id"
                  type="number"
                  class="form-control"
                  placeholder="Azure Pipeline ID"
                  required
                />
              </div>
            </div>

            <div class="form-section">
              <h4 class="section-title">流水线参数</h4>
              <div class="form-group">
                <label for="pipelineParams">流水线参数 (JSON)</label>
                <textarea
                  id="pipelineParams"
                  v-model="paramsText"
                  class="form-control code-editor"
                  rows="6"
                  placeholder='{"key": "value"}'
                  @input="onParamsChange"
                ></textarea>
                <p v-if="paramsError" class="error-text">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  {{ paramsError }}
                </p>
              </div>
            </div>

            <div class="form-section">
              <div class="form-group checkbox-group">
                <label class="checkbox-label">
                  <input type="checkbox" v-model="formData.enabled" />
                  <span class="checkbox-custom"></span>
                  <span class="checkbox-text">启用此模板</span>
                </label>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeModals">
              取消
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting || !!paramsError">
              <svg v-if="submitting" class="spinner-small" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              {{ submitting ? '保存中...' : '保存' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal modal-small">
        <div class="modal-header">
          <div class="modal-title-row">
            <div class="modal-icon modal-icon-danger">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h3 class="modal-title">确认删除</h3>
          </div>
          <button class="modal-close" @click="showDeleteModal = false">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="delete-warning">
            <p>确定要删除模板 <strong>{{ deletingTemplate?.name }}</strong> 吗？</p>
            <p class="warning-text">此操作无法撤销。</p>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="showDeleteModal = false">
            取消
          </button>
          <button type="button" class="btn btn-danger" :disabled="submitting" @click="confirmDelete">
            <svg v-if="submitting" class="spinner-small" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            {{ submitting ? '删除中...' : '删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { adminAPI } from '../../api/resourcePool'
import { useDialog } from '../../composables/useDialog'

export default {
  name: 'PipelineTemplateManage',
  setup() {
    const dialog = useDialog()
    const templates = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const showCreateModal = ref(false)
    const showEditModal = ref(false)
    const showDeleteModal = ref(false)
    const editingTemplate = ref(null)
    const deletingTemplate = ref(null)
    const paramsText = ref('')
    const paramsError = ref('')

    const formData = ref({
      name: '',
      description: '',
      organization: '',
      project: '',
      pipeline_id: 0,
      pipeline_parameters: {},
      enabled: true
    })

    const fetchTemplates = async () => {
      loading.value = true
      try {
        const data = await adminAPI.getPipelineTemplates()
        templates.value = data.templates || []
      } catch (error) {
        console.error('Failed to fetch pipeline templates:', error)
        dialog.alertError('获取部署管道模板列表失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    const onParamsChange = () => {
      paramsError.value = ''
      if (!paramsText.value.trim()) {
        formData.value.pipeline_parameters = {}
        return
      }
      try {
        formData.value.pipeline_parameters = JSON.parse(paramsText.value)
      } catch (e) {
        paramsError.value = 'JSON 格式错误: ' + e.message
      }
    }

    const formatParams = (params) => {
      if (!params || Object.keys(params).length === 0) return '-'
      const keys = Object.keys(params).slice(0, 2)
      if (Object.keys(params).length > 2) {
        return keys.join(', ') + '...'
      }
      return keys.join(', ')
    }

    const saveTemplate = async () => {
      submitting.value = true
      try {
        if (editingTemplate.value) {
          await adminAPI.updatePipelineTemplate(editingTemplate.value.id, formData.value)
          dialog.alertSuccess('模板更新成功')
        } else {
          await adminAPI.createPipelineTemplate(formData.value)
          dialog.alertSuccess('模板创建成功')
        }
        closeModals()
        await fetchTemplates()
      } catch (error) {
        console.error('Failed to save template:', error)
        dialog.alertError('保存模板失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }

    const editTemplate = (template) => {
      editingTemplate.value = template
      formData.value = {
        name: template.name,
        description: template.description,
        organization: template.organization,
        project: template.project,
        pipeline_id: template.pipeline_id,
        pipeline_parameters: template.pipeline_parameters || {},
        enabled: template.enabled
      }
      paramsText.value = JSON.stringify(template.pipeline_parameters || {}, null, 2)
      showEditModal.value = true
    }

    const deleteTemplate = (template) => {
      deletingTemplate.value = template
      showDeleteModal.value = true
    }

    const confirmDelete = async () => {
      submitting.value = true
      try {
        await adminAPI.deletePipelineTemplate(deletingTemplate.value.id)
        dialog.alertSuccess('模板删除成功')
        showDeleteModal.value = false
        await fetchTemplates()
      } catch (error) {
        console.error('Failed to delete template:', error)
        dialog.alertError('删除模板失败: ' + error.message)
      } finally {
        submitting.value = false
        deletingTemplate.value = null
      }
    }

    const toggleTemplate = async (template) => {
      try {
        if (template.enabled) {
          await adminAPI.disablePipelineTemplate(template.id)
          template.enabled = false
        } else {
          await adminAPI.enablePipelineTemplate(template.id)
          template.enabled = true
        }
      } catch (error) {
        console.error('Failed to toggle template:', error)
        dialog.alertError('切换模板状态失败: ' + error.message)
      }
    }

    const closeModals = () => {
      showCreateModal.value = false
      showEditModal.value = false
      editingTemplate.value = null
      paramsText.value = ''
      paramsError.value = ''
      formData.value = {
        name: '',
        description: '',
        organization: '',
        project: '',
        pipeline_id: 0,
        pipeline_parameters: {},
        enabled: true
      }
    }

    onMounted(() => {
      fetchTemplates()
    })

    return {
      templates,
      loading,
      submitting,
      showCreateModal,
      showEditModal,
      showDeleteModal,
      editingTemplate,
      deletingTemplate,
      formData,
      paramsText,
      paramsError,
      saveTemplate,
      editTemplate,
      deleteTemplate,
      confirmDelete,
      toggleTemplate,
      closeModals,
      onParamsChange,
      formatParams
    }
  }
}
</script>

<style scoped>
.pipeline-template-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  margin-bottom: 1.5rem;
  padding: 1.5rem;
  background: var(--bg-card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  border: 1px solid var(--border);
}

.header-icon {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  background: linear-gradient(135deg, #8B5CF6, #7C3AED);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.header-icon svg {
  width: 26px;
  height: 26px;
}

.header-content {
  flex: 1;
}

.header-content h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 0.25rem 0;
}

.header-content p {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0;
}

.card {
  background: var(--bg-card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  border: 1px solid var(--border);
  overflow: hidden;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem;
  gap: 1rem;
  color: var(--text-secondary);
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
}

.empty-icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.1), rgba(124, 58, 237, 0.1));
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.empty-icon svg {
  width: 40px;
  height: 40px;
  color: #8B5CF6;
}

.empty-state h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 0.5rem 0;
}

.empty-state p {
  color: var(--text-secondary);
  margin: 0 0 1.5rem 0;
}

.templates-table-wrapper {
  overflow-x: auto;
}

.templates-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.templates-table th,
.templates-table td {
  padding: 1rem 1.25rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

.templates-table td {
  /* Apply text overflow to all table cells */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Ensure inline elements in table cells also truncate */
.templates-table td > *,
.templates-table td > span,
.templates-table td > a {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

.templates-table th {
  background: var(--bg-main);
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
}

.templates-table tbody tr {
  transition: background 0.2s ease;
}

.templates-table tbody tr:hover {
  background: var(--bg-secondary);
}

.templates-table tbody tr:last-child td {
  border-bottom: none;
}

.template-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.template-icon {
  width: 18px;
  height: 18px;
  color: #8B5CF6;
  flex-shrink: 0;
}

.template-name {
  font-weight: 600;
  color: var(--text-primary);
}

.org-project-cell {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.separator {
  color: var(--text-secondary);
  font-weight: 400;
}

.pipeline-id {
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  font-size: 0.875rem;
  font-weight: 600;
  color: #8B5CF6;
  background: rgba(139, 92, 246, 0.1);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}

.params-cell {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.params-preview {
  font-size: 0.8125rem;
  color: var(--text-primary);
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.params-count {
  font-size: 0.6875rem;
  color: var(--text-secondary);
}

.no-params {
  color: var(--text-secondary);
}

.description {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-enabled {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
}

.status-enabled .status-dot {
  background: #10B981;
}

.status-disabled {
  background: rgba(107, 114, 128, 0.1);
  color: #6B7280;
}

.status-disabled .status-dot {
  background: #6B7280;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  border-radius: var(--radius);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.btn svg {
  width: 18px;
  height: 18px;
}

.btn-primary {
  background: linear-gradient(135deg, #8B5CF6, #7C3AED);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #7C3AED, #6D28D9);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-main);
  color: var(--text-primary);
  border: 1px solid var(--border);
}

.btn-secondary:hover {
  background: var(--bg-secondary);
  border-color: var(--text-secondary);
}

.btn-danger {
  background: linear-gradient(135deg, #EF4444, #DC2626);
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: linear-gradient(135deg, #DC2626, #B91C1C);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.btn-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  border: none;
  background: transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-secondary);
}

.btn-icon svg {
  width: 18px;
  height: 18px;
}

.btn-icon:hover {
  background: var(--bg-main);
  color: #8B5CF6;
}

.btn-icon-success:hover {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
}

.btn-icon-warning:hover {
  background: rgba(245, 158, 11, 0.1);
  color: #F59E0B;
}

.btn-icon-danger:hover {
  background: rgba(239, 68, 68, 0.1);
  color: #EF4444;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal {
  background: var(--bg-card);
  border-radius: var(--radius);
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 600px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-small {
  max-width: 450px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.modal-title-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.modal-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.1), rgba(124, 58, 237, 0.1));
  display: flex;
  align-items: center;
  justify-content: center;
  color: #8B5CF6;
}

.modal-icon svg {
  width: 20px;
  height: 20px;
}

.modal-icon-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #EF4444;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.modal-close {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.modal-close svg {
  width: 20px;
  height: 20px;
}

.modal-close:hover {
  background: var(--bg-main);
  color: var(--text-primary);
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
}

.form-section {
  margin-bottom: 1.5rem;
}

.form-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0 0 1rem 0;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid var(--border);
}

.form-group {
  margin-bottom: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: var(--text-primary);
  font-size: 0.875rem;
}

.required {
  color: #EF4444;
}

.form-control {
  width: 100%;
  padding: 0.625rem 0.875rem;
  border: 1px solid var(--border);
  border-radius: var(--radius);
  font-size: 0.875rem;
  background: var(--bg-primary);
  color: var(--text-primary);
  transition: all 0.2s ease;
}

.form-control:focus {
  outline: none;
  border-color: #8B5CF6;
  box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.1);
}

.form-control::placeholder {
  color: var(--text-secondary);
}

textarea.form-control {
  resize: vertical;
  min-height: 80px;
}

.code-editor {
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  font-size: 0.8125rem;
  background: var(--bg-main);
}

.error-text {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  color: #EF4444;
  font-size: 0.75rem;
  margin-top: 0.5rem;
}

.error-text svg {
  width: 14px;
  height: 14px;
}

.checkbox-group {
  margin-top: 0.5rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  cursor: pointer;
  font-weight: 400;
}

.checkbox-label input[type="checkbox"] {
  display: none;
}

.checkbox-custom {
  width: 20px;
  height: 20px;
  border: 2px solid var(--border);
  border-radius: 4px;
  position: relative;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom {
  background: linear-gradient(135deg, #8B5CF6, #7C3AED);
  border-color: #8B5CF6;
}

.checkbox-label input[type="checkbox"]:checked + .checkbox-custom::after {
  content: '';
  position: absolute;
  left: 6px;
  top: 2px;
  width: 5px;
  height: 10px;
  border: solid white;
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.checkbox-text {
  color: var(--text-primary);
  font-size: 0.875rem;
}

.delete-warning {
  text-align: center;
}

.delete-warning p {
  margin: 0 0 0.5rem 0;
  color: var(--text-primary);
}

.delete-warning strong {
  color: #8B5CF6;
}

.warning-text {
  color: #F59E0B;
  font-size: 0.875rem;
}

.spinner-small {
  width: 16px;
  height: 16px;
  animation: spin 0.8s linear infinite;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .templates-table {
    font-size: 0.8125rem;
  }

  .templates-table th,
  .templates-table td {
    padding: 0.75rem 1rem;
  }
}
</style>
