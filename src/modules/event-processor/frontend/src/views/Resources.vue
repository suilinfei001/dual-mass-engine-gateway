<template>
  <div class="resources-page">
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Executable Resources</h2>
        <button
          v-if="isLoggedIn && !isAdmin"
          class="btn btn-primary"
          @click="showCreateModal = true"
        >
          Create Resource
        </button>
      </div>
      
      <div class="tabs">
        <button
          class="tab"
          :class="{ active: activeTab === 'all' }"
          @click="activeTab = 'all'"
        >
          All Resources
        </button>
        <button
          v-if="isLoggedIn && !isAdmin"
          class="tab"
          :class="{ active: activeTab === 'my' }"
          @click="activeTab = 'my'"
        >
          My Resources
        </button>
      </div>
      
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
      </div>
      
      <div v-else-if="resources.length === 0" class="empty-state">
        <p>No resources found</p>
      </div>

      <div v-else class="table-wrapper">
        <table>
          <thead>
            <tr>
              <th class="col-name">Name</th>
              <th class="col-type">Type</th>
              <th class="col-skip">Skip</th>
              <th v-if="activeTab === 'my'" class="col-org">Organization</th>
              <th v-if="activeTab === 'my'" class="col-project">Project</th>
              <th class="col-pipeline">Pipeline ID</th>
              <th class="col-micro">Microservice</th>
              <th class="col-repo">Repo Path</th>
              <th class="col-creator">Creator</th>
              <th v-if="activeTab === 'my'" class="col-actions">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="resource in resources" :key="resource.id">
              <td class="cell-nowrap">{{ resource.resource_name }}</td>
              <td class="cell-nowrap">
                <span class="badge badge-info">{{ formatResourceType(resource.resource_type) }}</span>
              </td>
              <td class="cell-nowrap">
                <span v-if="resource.allow_skip" class="badge badge-warning">可跳过</span>
                <span v-else class="text-muted">-</span>
              </td>
              <td v-if="activeTab === 'my'" class="cell-nowrap">{{ resource.organization || '-' }}</td>
              <td v-if="activeTab === 'my'" class="cell-nowrap">{{ resource.project || '-' }}</td>
              <td class="cell-nowrap">{{ resource.pipeline_id }}</td>
              <td class="cell-nowrap">{{ resource.microservice_name || '-' }}</td>
              <td class="cell-truncate" :title="resource.repo_path">{{ resource.repo_path }}</td>
              <td class="cell-nowrap">{{ resource.creator_name }}</td>
              <td v-if="activeTab === 'my'" class="cell-actions">
                <template v-if="canEdit(resource)">
                  <div class="action-buttons">
                    <button class="btn btn-secondary btn-sm" @click="editResource(resource)">Edit</button>
                    <button class="btn btn-danger btn-sm" @click="confirmDelete(resource)">Delete</button>
                  </div>
                </template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    
    <!-- Create/Edit Modal -->
    <div v-if="showCreateModal || editingResource" class="modal-overlay" @click="closeModal">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">{{ editingResource ? 'Edit Resource' : 'Create Resource' }}</h3>
          <button class="modal-close" @click="closeModal">&times;</button>
        </div>
        
        <form @submit.prevent="handleSubmit">
          <!-- Allow Skip Option - Placed at the top -->
          <div class="form-group skip-option">
            <label class="checkbox-label">
              <input
                type="checkbox"
                id="allowSkip"
                v-model="form.allow_skip"
              />
              <span>允许跳过此检查项</span>
            </label>
            <small class="text-muted">勾选后，此检查项可以被跳过执行</small>
          </div>

          <!-- Skip Mode: Only show basic fields -->
          <template v-if="form.allow_skip">
            <div class="skip-mode-notice">
              <p><strong>跳过模式</strong> - 只需填写基本信息，Azure 配置将自动设置为空</p>
            </div>

            <div class="form-group">
              <label for="resourceType">Resource Type *</label>
              <select
                id="resourceType"
                v-model="form.resource_type"
                class="form-control"
                :disabled="editingResource"
                :class="{ 'form-control-disabled': editingResource }"
                required
              >
                <option value="">Select type</option>
                <option value="basic_ci_all">Basic CI All</option>
                <option value="deployment_deployment">Deployment</option>
                <option value="specialized_tests_api_test">API Test</option>
                <option value="specialized_tests_module_e2e">Module E2E</option>
                <option value="specialized_tests_agent_e2e">Agent E2E</option>
                <option value="specialized_tests_ai_e2e">AI E2E</option>
              </select>
              <small v-if="editingResource" class="text-muted">Resource type cannot be changed</small>
            </div>

            <div class="form-group">
              <label for="repoPath">Repo Path *</label>
              <input
                type="text"
                id="repoPath"
                v-model="form.repo_path"
                class="form-control"
                required
                placeholder="e.g., github.com/org/repo"
              />
            </div>

            <div class="form-group">
              <label for="description">Description</label>
              <textarea
                id="description"
                v-model="form.description"
                class="form-control"
                rows="3"
                placeholder="此检查项可以跳过"
              ></textarea>
              <small class="text-muted">描述会自动添加"此检查项可以跳过"标记</small>
            </div>
          </template>

          <!-- Normal Mode: Show all fields -->
          <template v-else>
            <div class="form-group">
              <label for="resourceType">Resource Type *</label>
              <select
                id="resourceType"
                v-model="form.resource_type"
                class="form-control"
                :disabled="editingResource"
                :class="{ 'form-control-disabled': editingResource }"
                required
              >
                <option value="">Select type</option>
                <option value="basic_ci_all">Basic CI All</option>
                <option value="deployment_deployment">Deployment</option>
                <option value="specialized_tests_api_test">API Test</option>
                <option value="specialized_tests_module_e2e">Module E2E</option>
                <option value="specialized_tests_agent_e2e">Agent E2E</option>
                <option value="specialized_tests_ai_e2e">AI E2E</option>
              </select>
              <small v-if="editingResource" class="text-muted">Resource type cannot be changed</small>
            </div>

            <div class="form-group">
              <label for="organization">Azure DevOps Organization *</label>
              <input
                type="text"
                id="organization"
                v-model="form.organization"
                class="form-control"
                required
                placeholder="e.g., MyOrg"
              />
              <small class="text-muted">Azure DevOps organization name</small>
            </div>

            <div class="form-group">
              <label for="project">Azure DevOps Project *</label>
              <input
                type="text"
                id="project"
                v-model="form.project"
                class="form-control"
                required
                placeholder="e.g., MyProject"
              />
              <small class="text-muted">Azure DevOps project name</small>
            </div>

            <div class="form-group">
              <label for="pipelineId">Pipeline ID *</label>
              <input
                type="number"
                id="pipelineId"
                v-model.number="form.pipeline_id"
                class="form-control"
                required
              />
            </div>

            <div class="form-group">
              <label for="pipelineParams">Pipeline Parameters (JSON)</label>
              <textarea
                id="pipelineParams"
                v-model="form.pipeline_params"
                class="form-control"
                rows="6"
                placeholder='{
  "TRIVY_EXIT_CODE": "1",
  "SKIP_SONARQUBE": "0",
  "SOURCE_CODE_BRANCH": "main",
  "BUILD_TYPE": "opensource",
  "BUILD_ARM": "false"
}'
              ></textarea>
              <small class="text-muted">JSON格式，例如：{"TRIVY_EXIT_CODE": "1", "BUILD_TYPE": "opensource"}</small>
            </div>

            <div class="form-group">
              <label for="microserviceName">Microservice Name</label>
              <input
                type="text"
                id="microserviceName"
                v-model="form.microservice_name"
                class="form-control"
              />
            </div>

            <div class="form-group">
              <label for="podName">Pod Name</label>
              <input
                type="text"
                id="podName"
                v-model="form.pod_name"
                class="form-control"
              />
            </div>

            <div class="form-group">
              <label for="repoPath">Repo Path *</label>
              <input
                type="text"
                id="repoPath"
                v-model="form.repo_path"
                class="form-control"
                required
              />
            </div>

            <div class="form-group">
              <label for="description">Description</label>
              <textarea
                id="description"
                v-model="form.description"
                class="form-control"
                rows="3"
              ></textarea>
            </div>
          </template>

          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeModal">Cancel</button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </form>
      </div>
    </div>
    
    <!-- Delete Confirmation Modal -->
    <div v-if="deletingResource" class="modal-overlay" @click="deletingResource = null">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">Confirm Delete</h3>
          <button class="modal-close" @click="deletingResource = null">&times;</button>
        </div>
        <p>Are you sure you want to delete resource "{{ deletingResource.resource_name }}"?</p>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="deletingResource = null">Cancel</button>
          <button class="btn btn-danger" @click="deleteResource" :disabled="submitting">
            {{ submitting ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted, watch } from 'vue'

export default {
  name: 'Resources',
  setup() {
    const resources = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const activeTab = ref('all')
    const showCreateModal = ref(false)
    const editingResource = ref(null)
    const deletingResource = ref(null)
    
    const isLoggedIn = computed(() => localStorage.getItem('isLoggedIn') === 'true')
    const isAdmin = computed(() => localStorage.getItem('userRole') === 'admin')
    
    const form = ref({
      resource_name: '',
      resource_type: '',
      allow_skip: false,
      organization: '',
      project: '',
      pipeline_id: null,
      pipeline_params: '',
      microservice_name: '',
      pod_name: '',
      repo_path: '',
      description: ''
    })
    
    const fetchResources = async () => {
      loading.value = true
      try {
        const url = activeTab.value === 'my' ? '/api/resources/my' : '/api/resources'
        const response = await fetch(url, { credentials: 'include' })
        const data = await response.json()
        resources.value = data.data || []
      } catch (error) {
        console.error('Failed to fetch resources:', error)
      } finally {
        loading.value = false
      }
    }
    
    const canEdit = (resource) => {
      if (!isLoggedIn.value) return false
      if (isAdmin.value) return true
      const username = localStorage.getItem('username')
      return resource.creator_name === username
    }
    
    const editResource = (resource) => {
      editingResource.value = resource
      // Convert pipeline_params object to JSON string for editing
      const pipelineParamsJson = resource.pipeline_params
        ? JSON.stringify(resource.pipeline_params, null, 2)
        : ''

      form.value = {
        resource_name: resource.resource_name,
        resource_type: resource.resource_type,
        allow_skip: resource.allow_skip || false,
        organization: resource.organization || '',
        project: resource.project || '',
        pipeline_id: resource.pipeline_id,
        pipeline_params: pipelineParamsJson,
        microservice_name: resource.microservice_name || '',
        pod_name: resource.pod_name || '',
        repo_path: resource.repo_path,
        description: resource.description || ''
      }
    }

    const closeModal = () => {
      showCreateModal.value = false
      editingResource.value = null
      form.value = {
        resource_name: '',
        resource_type: '',
        allow_skip: false,
        organization: '',
        project: '',
        pipeline_id: null,
        pipeline_params: '',
        microservice_name: '',
        pod_name: '',
        repo_path: '',
        description: ''
      }
    }
    
    const handleSubmit = async () => {
      submitting.value = true
      try {
        const url = editingResource.value
          ? `/api/resources/${editingResource.value.id}`
          : '/api/resources'
        const method = editingResource.value ? 'PUT' : 'POST'

        // Prepare the payload
        const payload = { ...form.value }

        // Handle skip mode
        if (payload.allow_skip) {
          // Clear Azure-related fields for skip mode
          payload.organization = ''
          payload.project = ''
          payload.pipeline_id = null
          payload.pipeline_params = {}
          payload.microservice_name = ''
          payload.pod_name = ''
          // Auto-populate description if empty
          if (!payload.description || !payload.description.trim()) {
            payload.description = '此检查项可以跳过'
          } else if (!payload.description.includes('可以跳过')) {
            payload.description = payload.description + ' (此检查项可以跳过)'
          }
        } else {
          // Parse pipeline_params as JSON if provided
          if (payload.pipeline_params && payload.pipeline_params.trim()) {
            try {
              payload.pipeline_params = JSON.parse(payload.pipeline_params)
            } catch (e) {
              alert('Pipeline Parameters must be valid JSON format')
              submitting.value = false
              return
            }
          } else {
            payload.pipeline_params = {}
          }
        }

        // Auto-generate resource_name if not provided (for new resources)
        if (!editingResource.value && !payload.resource_name) {
          const timestamp = Date.now().toString().slice(-6)
          payload.resource_name = `${payload.resource_type}_${timestamp}`
        }

        const response = await fetch(url, {
          method,
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify(payload)
        })

        const data = await response.json()

        if (data.success) {
          closeModal()
          fetchResources()
        } else {
          alert(data.message || 'Operation failed')
        }
      } catch (error) {
        console.error('Failed to save resource:', error)
        alert('An error occurred')
      } finally {
        submitting.value = false
      }
    }
    
    const confirmDelete = (resource) => {
      deletingResource.value = resource
    }
    
    const deleteResource = async () => {
      submitting.value = true
      try {
        const response = await fetch(`/api/resources/${deletingResource.value.id}`, {
          method: 'DELETE',
          credentials: 'include'
        })
        
        const data = await response.json()
        
        if (data.success) {
          deletingResource.value = null
          fetchResources()
        } else {
          alert(data.message || 'Delete failed')
        }
      } catch (error) {
        console.error('Failed to delete resource:', error)
        alert('An error occurred')
      } finally {
        submitting.value = false
      }
    }
    
    const formatResourceType = (type) => {
      const types = {
        'basic_ci_all': 'Basic CI',
        'deployment_deployment': 'Deployment',
        'specialized_tests_api_test': 'API Test',
        'specialized_tests_module_e2e': 'Module E2E',
        'specialized_tests_agent_e2e': 'Agent E2E',
        'specialized_tests_ai_e2e': 'AI E2E'
      }
      return types[type] || type
    }
    
    watch(activeTab, fetchResources)
    onMounted(fetchResources)
    
    return {
      resources,
      loading,
      submitting,
      activeTab,
      showCreateModal,
      editingResource,
      deletingResource,
      form,
      isLoggedIn,
      isAdmin,
      fetchResources,
      canEdit,
      editResource,
      closeModal,
      handleSubmit,
      confirmDelete,
      deleteResource,
      formatResourceType
    }
  }
}
</script>

<style scoped>
.resources-page {
  max-width: 1200px;
  margin: 0 auto;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: var(--text-secondary);
}

textarea.form-control {
  resize: vertical;
}

table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
}

table th,
table td {
  text-align: left;
  padding: 0.875rem 1rem;
  border-bottom: 1px solid var(--border);
}

table th {
  background-color: #F8FAFC;
  font-weight: 600;
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #64748B;
}

table th:first-child {
  border-top-left-radius: var(--radius-sm);
}

table th:last-child {
  border-top-right-radius: var(--radius-sm);
}

table tbody tr:hover {
  background-color: var(--bg-main);
}

table tbody tr:last-child td {
  border-bottom: none;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.card-title {
  margin: 0;
}

.tabs {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  border-bottom: 2px solid var(--border);
}

.tab {
  padding: 0.75rem 1.5rem;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  color: var(--text-secondary);
  font-weight: 500;
  font-size: 0.9375rem;
  transition: all 0.2s ease;
  margin-bottom: -2px;
}

.tab:hover {
  color: var(--primary);
}

.tab.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}

/* Table wrapper for horizontal scroll */
.table-wrapper {
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

/* Prevent text wrapping in cells */
.cell-nowrap {
  white-space: nowrap;
}

/* Truncate long text with ellipsis */
.cell-truncate {
  max-width: 200px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Actions column styling */
.cell-actions {
  white-space: nowrap;
  width: 140px;
}

.action-buttons {
  display: inline-flex;
  gap: 0.5rem;
}

.action-buttons .btn-sm {
  margin: 0;
}

/* Column width hints */
.col-name { min-width: 150px; }
.col-type { min-width: 100px; }
.col-skip { min-width: 70px; }
.col-org { min-width: 120px; }
.col-project { min-width: 120px; }
.col-pipeline { min-width: 80px; }
.col-micro { min-width: 120px; }
.col-repo { min-width: 200px; }
.col-creator { min-width: 100px; }
.col-actions { min-width: 140px; }

/* Badge styles */
.badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 4px;
}

.badge-info {
  background-color: #17a2b8;
  color: white;
}

.badge-warning {
  background-color: #ffc107;
  color: #212529;
}

.text-muted {
  color: #6c757d;
}

.form-control:disabled,
.form-control-disabled {
  background-color: #e9ecef;
  color: #6c757d;
  cursor: not-allowed;
}

.form-control:disabled:hover,
.form-control-disabled:hover {
  background-color: #e9ecef;
}

/* Skip option styling */
.skip-option {
  padding: 1rem;
  background-color: #f8f9fa;
  border: 2px solid #dee2e6;
  border-radius: 8px;
  margin-bottom: 1.5rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 600;
  font-size: 1rem;
  cursor: pointer;
  margin-bottom: 0.25rem;
}

.checkbox-label input[type="checkbox"] {
  width: 20px;
  height: 20px;
  cursor: pointer;
}

.checkbox-label span {
  user-select: none;
}

.skip-mode-notice {
  padding: 0.75rem 1rem;
  background-color: #e7f3ff;
  border-left: 4px solid #007bff;
  border-radius: 4px;
  margin-bottom: 1.5rem;
}

.skip-mode-notice p {
  margin: 0;
  color: #0056b3;
}

.skip-mode-notice strong {
  color: #004085;
}
</style>
