<template>
  <div class="console-page">
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Admin Console</h2>
      </div>
      
      <div class="tabs">
        <button
          class="tab"
          :class="{ active: activeTab === 'config' }"
          @click="activeTab = 'config'"
        >
          System Config
        </button>
        <button
          class="tab"
          :class="{ active: activeTab === 'users' }"
          @click="activeTab = 'users'"
        >
          User Management
        </button>
        <button
          class="tab"
          :class="{ active: activeTab === 'cleanup' }"
          @click="activeTab = 'cleanup'"
        >
          Data Cleanup
        </button>
      </div>
      
      <!-- System Config Tab -->
      <div v-if="activeTab === 'config'" class="tab-content">
        <div class="config-section">
          <h3>AI Configuration</h3>
          <div class="form-group">
            <label for="aiIp">AI Server IP</label>
            <input type="text" id="aiIp" v-model="config.ai_ip" class="form-control" placeholder="e.g., 192.168.112.38" />
          </div>
          <div class="form-group">
            <label for="aiModel">Model Name</label>
            <input type="text" id="aiModel" v-model="config.ai_model" class="form-control" placeholder="e.g., Tome-pro" />
          </div>
          <div class="form-group">
            <label for="aiToken">API Token</label>
            <input type="password" id="aiToken" v-model="config.ai_token" class="form-control" />
          </div>
          <div class="button-group">
            <button class="btn btn-primary" @click="saveAIConfig" :disabled="saving">
              {{ saving ? 'Saving...' : 'Save AI Config' }}
            </button>
            <button
              class="btn btn-info test-btn"
              @click="testAIConfig"
              :disabled="testingAI || !config.ai_ip || !config.ai_model || !config.ai_token"
              :title="!config.ai_ip || !config.ai_model || !config.ai_token ? 'Please fill in all AI configuration fields first' : 'Test AI configuration connection'"
            >
              {{ testingAI ? 'Testing...' : 'Test Connection' }}
            </button>
          </div>
          <div v-if="aiTestResult" class="test-result" :class="aiTestResult.success ? 'success' : 'error'">
            {{ aiTestResult.message }}
          </div>
        </div>
        
        <div class="config-section">
          <h3>Azure DevOps Configuration</h3>
          <div class="form-group">
            <label for="azurePat">Personal Access Token (PAT)</label>
            <input type="password" id="azurePat" v-model="config.azure_pat" class="form-control" />
          </div>
          <button class="btn btn-primary" @click="saveAzurePAT" :disabled="saving">
            {{ saving ? 'Saving...' : 'Save Azure PAT' }}
          </button>
        </div>
        
        <div class="config-section">
          <h3>Event Receiver Configuration</h3>
          <div v-if="!editingEventReceiverIP" class="config-display">
            <div class="config-item">
              <span class="config-label">Event Receiver IP:</span>
              <span class="config-value">{{ config.event_receiver_ip || 'Not configured' }}</span>
              <button class="btn btn-secondary" @click="editingEventReceiverIP = true">Change</button>
              <button
                class="btn btn-info test-btn"
                @click="testEventReceiverConnection"
                :disabled="testingConnection"
                :title="config.event_receiver_ip ? 'Test connection to Event Receiver' : 'Please configure Event Receiver IP first'"
              >
                {{ testingConnection ? 'Testing...' : 'Test Connection' }}
              </button>
            </div>
            <div v-if="connectionTestResult" class="test-result" :class="connectionTestResult.success ? 'success' : 'error'">
              {{ connectionTestResult.message }}
            </div>
          </div>
          <div v-else class="config-edit">
            <div class="form-group">
              <label for="eventReceiverIp">Event Receiver IP</label>
              <input type="text" id="eventReceiverIp" v-model="config.event_receiver_ip" class="form-control" placeholder="e.g., http://10.4.111.141:5001" />
            </div>
            <div class="button-group">
              <button class="btn btn-info" @click="testEventReceiverConnection" :disabled="testingConnection || !config.event_receiver_ip">
                {{ testingConnection ? 'Testing...' : 'Test' }}
              </button>
              <button class="btn btn-primary" @click="saveEventReceiverIP" :disabled="saving">
                {{ saving ? 'Saving...' : 'Save' }}
              </button>
              <button class="btn btn-secondary" @click="cancelEditEventReceiverIP">Cancel</button>
            </div>
            <div v-if="connectionTestResult" class="test-result" :class="connectionTestResult.success ? 'success' : 'error'">
              {{ connectionTestResult.message }}
            </div>
          </div>
        </div>

        <div class="config-section">
          <h3>Log Retention Configuration</h3>
          <div class="form-group">
            <label for="logRetentionDays">Log Retention Period (Days)</label>
            <input
              type="number"
              id="logRetentionDays"
              v-model.number="config.log_retention_days"
              class="form-control"
              min="1"
              max="30"
            />
            <small class="form-text">
              Log files will be retained for this many days before automatic cleanup (1-30 days, default: 7)
            </small>
          </div>
          <button class="btn btn-primary" @click="saveLogRetention" :disabled="saving">
            {{ saving ? 'Saving...' : 'Save Log Retention' }}
          </button>
        </div>

        <div class="config-section">
          <h3>AI Analysis Configuration</h3>
          <div class="form-group">
            <label for="aiConcurrency">AI Analysis Concurrency (Per Event)</label>
            <input
              type="number"
              id="aiConcurrency"
              v-model.number="config.ai_concurrency"
              class="form-control"
              min="1"
              max="50"
            />
            <small class="form-text">
              Number of log files to analyze concurrently per event (1-50, default: 20). Higher values process faster but use more API requests.
            </small>
          </div>
          <button class="btn btn-primary" @click="saveAIConcurrency" :disabled="saving">
            {{ saving ? 'Saving...' : 'Save AI Concurrency' }}
          </button>
        </div>

        <div class="config-section">
          <h3>AI Request Pool Configuration</h3>
          <div class="form-group">
            <label for="aiRequestPoolSize">AI Request Pool Size (Global)</label>
            <input
              type="number"
              id="aiRequestPoolSize"
              v-model.number="config.ai_request_pool_size"
              class="form-control"
              min="1"
              max="200"
            />
            <small class="form-text">
              Total concurrent AI requests across all events (1-200, default: 50). This must be greater than AI Analysis Concurrency. Recommended: not more than 2/3 of your AI model's actual concurrent capacity.
            </small>
          </div>
          <button class="btn btn-primary" @click="saveAIRequestPoolSize" :disabled="saving">
            {{ saving ? 'Saving...' : 'Save Request Pool Size' }}
          </button>
        </div>
      </div>
      
      <!-- User Management Tab -->
      <div v-if="activeTab === 'users'" class="tab-content">
        <div class="search-section">
          <div class="search-form">
            <input
              type="text"
              v-model="userSearch"
              class="form-control"
              placeholder="Search by username or email..."
              @keyup.enter="searchUsers"
            />
            <button class="btn btn-primary" @click="searchUsers">Search</button>
          </div>
        </div>
        
        <div v-if="users.length > 0">
          <table>
            <thead>
              <tr>
                <th>ID</th>
                <th>Username</th>
                <th>Email</th>
                <th>Role</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td>{{ user.id }}</td>
                <td>{{ user.username }}</td>
                <td>{{ user.email }}</td>
                <td>
                  <span class="badge" :class="user.role === 'admin' ? 'badge-danger' : 'badge-info'">
                    {{ user.role }}
                  </span>
                </td>
                <td>
                  <button
                    v-if="user.role !== 'admin'"
                    class="btn btn-secondary btn-sm"
                    @click="showPasswordModal(user)"
                  >
                    Change Password
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          
          <!-- 分页控件 -->
          <div class="pagination">
            <div class="pagination-info">
              Showing {{ (currentPage - 1) * pageSize + 1 }} to {{ Math.min(currentPage * pageSize, totalUsers) }} of {{ totalUsers }} users
            </div>
            <div class="pagination-controls">
              <button class="btn btn-sm" @click="prevPage" :disabled="currentPage === 1">
                Previous
              </button>
              <span class="page-info">
                Page {{ currentPage }} of {{ Math.ceil(totalUsers / pageSize) }}
              </span>
              <button class="btn btn-sm" @click="nextPage" :disabled="currentPage * pageSize >= totalUsers">
                Next
              </button>
              <div class="page-size-selector">
                <label>Page size:</label>
                <select v-model="pageSize" @change="changePageSize(pageSize)">
                  <option value="10">10</option>
                  <option value="20">20</option>
                  <option value="50">50</option>
                  <option value="100">100</option>
                </select>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Password Change Modal -->
        <div v-if="selectedUser" class="modal-overlay" @click="selectedUser = null">
          <div class="modal" @click.stop>
            <div class="modal-header">
              <h3 class="modal-title">Change Password for {{ selectedUser.username }}</h3>
              <button class="modal-close" @click="selectedUser = null">&times;</button>
            </div>
            <form @submit.prevent="updatePassword">
              <div class="form-group">
                <label for="newPassword">New Password</label>
                <input
                  type="password"
                  id="newPassword"
                  v-model="newPassword"
                  class="form-control"
                  required
                  minlength="6"
                />
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-secondary" @click="selectedUser = null">Cancel</button>
                <button type="submit" class="btn btn-primary" :disabled="saving">
                  {{ saving ? 'Updating...' : 'Update Password' }}
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>
      
      <!-- Data Cleanup Tab -->
      <div v-if="activeTab === 'cleanup'" class="tab-content">
        <div class="cleanup-section">
          <h3>Data Cleanup</h3>
          <p class="warning-text">Warning: These actions are irreversible. Please proceed with caution.</p>

          <div class="cleanup-actions">
            <div class="cleanup-item">
              <h4>Cleanup Tasks</h4>
              <p>Delete all tasks (including task results) from the system.</p>
              <button class="btn btn-danger" @click="cleanupTasks" :disabled="cleaningTasks">
                {{ cleaningTasks ? 'Cleaning...' : 'Cleanup Tasks' }}
              </button>
            </div>

            <div class="cleanup-item">
              <h4>Cleanup Task Results</h4>
              <p>Delete all task results from the system (tasks remain).</p>
              <button class="btn btn-danger" @click="cleanupTaskResults" :disabled="cleaningTaskResults">
                {{ cleaningTaskResults ? 'Cleaning...' : 'Cleanup Task Results' }}
              </button>
            </div>

            <div class="cleanup-item">
              <h4>Cleanup Resources</h4>
              <p>Delete all executable resources from the system.</p>
              <button class="btn btn-danger" @click="cleanupResources" :disabled="cleaningResources">
                {{ cleaningResources ? 'Cleaning...' : 'Cleanup Resources' }}
              </button>
            </div>

            <div class="cleanup-item">
              <h4>Cleanup Sessions</h4>
              <p>Delete all expired sessions from the system.</p>
              <button class="btn btn-warning" @click="cleanupSessions" :disabled="cleaningSessions">
                {{ cleaningSessions ? 'Cleaning...' : 'Cleanup Expired Sessions' }}
              </button>
            </div>

            <div class="cleanup-item">
              <h4>Cleanup All Sessions</h4>
              <p>Delete all sessions (force logout all users).</p>
              <button class="btn btn-warning" @click="cleanupAllSessions" :disabled="cleaningAllSessions">
                {{ cleaningAllSessions ? 'Cleaning...' : 'Cleanup All Sessions' }}
              </button>
            </div>

            <div class="cleanup-item">
              <h4>Cleanup Users</h4>
              <p>Delete all non-admin users from the system.</p>
              <button class="btn btn-danger" @click="cleanupUsers" :disabled="cleaningUsers">
                {{ cleaningUsers ? 'Cleaning...' : 'Cleanup Users' }}
              </button>
            </div>

            <div class="cleanup-item danger-zone">
              <h4>Cleanup All</h4>
              <p>Delete all tasks, resources, and non-admin users.</p>
              <button class="btn btn-danger" @click="cleanupAll" :disabled="cleaningAll">
                {{ cleaningAll ? 'Cleaning...' : 'Cleanup All Data' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, watch } from 'vue'

export default {
  name: 'Console',
  setup() {
    const activeTab = ref('config')
    const saving = ref(false)
    const testingConnection = ref(false)
    const testingAI = ref(false)
    const connectionTestResult = ref(null)
    const aiTestResult = ref(null)

    // Separate cleaning states for each cleanup operation
    const cleaningTasks = ref(false)
    const cleaningTaskResults = ref(false)
    const cleaningResources = ref(false)
    const cleaningSessions = ref(false)
    const cleaningAllSessions = ref(false)
    const cleaningUsers = ref(false)
    const cleaningAll = ref(false)

    const config = ref({
      ai_ip: '',
      ai_model: '',
      ai_token: '',
      azure_pat: '',
      event_receiver_ip: '',
      log_retention_days: 7,
      ai_concurrency: 20,
      ai_request_pool_size: 50
    })
    
    const editingEventReceiverIP = ref(false)
    const originalEventReceiverIP = ref('')
    
    const userSearch = ref('')
    const users = ref([])
    const selectedUser = ref(null)
    const newPassword = ref('')
    
    // 分页相关状态
    const currentPage = ref(1)
    const pageSize = ref(20)
    const totalUsers = ref(0)
    
    const fetchConfig = async () => {
      try {
        const response = await fetch('/api/admin/config', { credentials: 'include' })
        const data = await response.json()
        if (data.success) {
          config.value = {
            ai_ip: data.data.ai_ip?.value || '',
            ai_model: data.data.ai_model?.value || '',
            ai_token: data.data.ai_token?.value || '',
            azure_pat: data.data.azure_pat?.value || '',
            event_receiver_ip: data.data.event_receiver_ip?.value || '',
            log_retention_days: parseInt(data.data.log_retention_days?.value || '7'),
            ai_concurrency: parseInt(data.data.ai_concurrency?.value || '20'),
            ai_request_pool_size: parseInt(data.data.ai_request_pool_size?.value || '50')
          }
          originalEventReceiverIP.value = data.data.event_receiver_ip?.value || ''
        }
      } catch (error) {
        console.error('Failed to fetch config:', error)
      }
    }
    
    const saveAIConfig = async () => {
      saving.value = true
      try {
        const response = await fetch('/api/admin/config/ai', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({
            ip: config.value.ai_ip,
            model: config.value.ai_model,
            token: config.value.ai_token
          })
        })
        const data = await response.json()
        if (data.success) {
          alert('AI config saved successfully')
        } else {
          alert(data.message || 'Failed to save')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        saving.value = false
      }
    }

    const testAIConfig = async () => {
      if (!config.value.ai_ip || !config.value.ai_model || !config.value.ai_token) {
        aiTestResult.value = { success: false, message: 'Please fill in all AI configuration fields' }
        return
      }

      testingAI.value = true
      aiTestResult.value = null

      try {
        const response = await fetch('/api/admin/config/ai/test', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({
            ip: config.value.ai_ip,
            model: config.value.ai_model,
            token: config.value.ai_token
          })
        })
        const data = await response.json()

        if (data.success) {
          const modelName = data.data?.model_name || config.value.ai_model
          aiTestResult.value = {
            success: true,
            message: `AI connection successful! Model: ${modelName}`
          }
        } else {
          aiTestResult.value = {
            success: false,
            message: data.message || 'AI connection failed'
          }
        }
      } catch (error) {
        aiTestResult.value = {
          success: false,
          message: 'Network error: ' + error.message
        }
      } finally {
        testingAI.value = false
      }
    }
    
    const saveAzurePAT = async () => {
      saving.value = true
      try {
        const response = await fetch('/api/admin/config/azure-pat', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ pat: config.value.azure_pat })
        })
        const data = await response.json()
        if (data.success) {
          alert('Azure PAT saved successfully')
        } else {
          alert(data.message || 'Failed to save')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        saving.value = false
      }
    }
    
    const saveEventReceiverIP = async () => {
      saving.value = true
      try {
        let ip = config.value.event_receiver_ip
        if (ip && !ip.startsWith('http://') && !ip.startsWith('https://')) {
          ip = 'http://' + ip
        }
        const response = await fetch('/api/admin/config/event-receiver', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ ip: ip })
        })
        const data = await response.json()
        if (data.success) {
          editingEventReceiverIP.value = false
          alert('Event receiver IP saved successfully')
        } else {
          alert(data.message || 'Failed to save')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        saving.value = false
      }
    }
    
    const cancelEditEventReceiverIP = () => {
      config.value.event_receiver_ip = originalEventReceiverIP.value
      editingEventReceiverIP.value = false
      connectionTestResult.value = null
    }

    const testEventReceiverConnection = async () => {
      const ip = config.value.event_receiver_ip
      if (!ip) {
        connectionTestResult.value = { success: false, message: 'Please enter Event Receiver IP first' }
        return
      }

      // Normalize URL
      let testUrl = ip
      if (!testUrl.startsWith('http://') && !testUrl.startsWith('https://')) {
        testUrl = 'http://' + testUrl
      }

      testingConnection.value = true
      connectionTestResult.value = null

      try {
        const response = await fetch('/api/admin/config/event-receiver/test', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ ip: testUrl })
        })
        const data = await response.json()

        if (data.success) {
          const eventCount = data.data?.event_count || 0
          connectionTestResult.value = {
            success: true,
            message: `Connection successful! Found ${eventCount} event(s).`
          }
        } else {
          connectionTestResult.value = {
            success: false,
            message: data.message || 'Connection failed'
          }
        }
      } catch (error) {
        connectionTestResult.value = {
          success: false,
          message: 'Network error: ' + error.message
        }
      } finally {
        testingConnection.value = false
      }
    }

    const saveLogRetention = async () => {
      const days = config.value.log_retention_days
      if (days < 1 || days > 30) {
        alert('Log retention days must be between 1 and 30')
        return
      }
      saving.value = true
      try {
        const response = await fetch('/api/admin/config/log-retention', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ days: days })
        })
        const data = await response.json()
        if (data.success) {
          alert(`Log retention period set to ${days} days`)
        } else {
          alert(data.message || 'Failed to save')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        saving.value = false
      }
    }

    const saveAIConcurrency = async () => {
      const concurrency = config.value.ai_concurrency
      if (concurrency < 1 || concurrency > 50) {
        alert('AI concurrency must be between 1 and 50')
        return
      }
      saving.value = true
      try {
        const response = await fetch('/api/admin/config/ai-concurrency', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ concurrency: concurrency })
        })
        const data = await response.json()
        if (data.success) {
          alert(`AI concurrency set to ${concurrency}`)
        } else {
          alert(data.message || 'Failed to save')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        saving.value = false
      }
    }

    const saveAIRequestPoolSize = async () => {
      const poolSize = config.value.ai_request_pool_size
      const concurrency = config.value.ai_concurrency
      if (poolSize < 1 || poolSize > 200) {
        alert('AI request pool size must be between 1 and 200')
        return
      }
      if (poolSize <= concurrency) {
        alert(`AI request pool size (${poolSize}) must be greater than AI concurrency (${concurrency})`)
        return
      }
      saving.value = true
      try {
        const response = await fetch('/api/admin/config/ai-request-pool-size', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ poolSize: poolSize })
        })
        const data = await response.json()
        if (data.success) {
          alert(`AI request pool size set to ${poolSize}`)
        } else {
          alert(data.message || 'Failed to save')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        saving.value = false
      }
    }

    const searchUsers = async () => {
      try {
        const keyword = userSearch.value || ''
        const response = await fetch(`/api/admin/users?keyword=${encodeURIComponent(keyword)}&page=${currentPage.value}&pageSize=${pageSize.value}`, {
          credentials: 'include'
        })
        const data = await response.json()
        users.value = data.data || []
        totalUsers.value = data.total || 0
      } catch (error) {
        console.error('Failed to search users:', error)
      }
    }
    
    const showPasswordModal = (user) => {
      selectedUser.value = user
      newPassword.value = ''
    }
    
    // 分页控制函数
    const changePage = (page) => {
      currentPage.value = page
      searchUsers()
    }
    
    const changePageSize = (size) => {
      pageSize.value = size
      currentPage.value = 1
      searchUsers()
    }
    
    const nextPage = () => {
      if (currentPage.value * pageSize.value < totalUsers.value) {
        currentPage.value++
        searchUsers()
      }
    }
    
    const prevPage = () => {
      if (currentPage.value > 1) {
        currentPage.value--
        searchUsers()
      }
    }
    
    const updatePassword = async () => {
      saving.value = true
      try {
        const response = await fetch(`/api/admin/users/${selectedUser.value.id}/password`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ new_password: newPassword.value })
        })
        const data = await response.json()
        if (data.success) {
          alert('Password updated successfully')
          selectedUser.value = null
        } else {
          alert(data.message || 'Failed to update')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        saving.value = false
      }
    }
    
    const cleanupTasks = async () => {
      if (!confirm('Are you sure you want to delete all tasks?')) return
      cleaningTasks.value = true
      try {
        const response = await fetch('/api/admin/cleanup/tasks', {
          method: 'POST',
          credentials: 'include'
        })
        const data = await response.json()
        if (data.success) {
          alert('All tasks cleaned up successfully')
        } else {
          alert(data.message || 'Failed to cleanup')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        cleaningTasks.value = false
      }
    }
    
    const cleanupResources = async () => {
      if (!confirm('Are you sure you want to delete all resources?')) return
      cleaningResources.value = true
      try {
        const response = await fetch('/api/admin/cleanup/resources', {
          method: 'POST',
          credentials: 'include'
        })
        const data = await response.json()
        if (data.success) {
          alert('All resources cleaned up successfully')
        } else {
          alert(data.message || 'Failed to cleanup')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        cleaningResources.value = false
      }
    }
    
    const cleanupUsers = async () => {
      if (!confirm('Are you sure you want to delete all non-admin users?')) return
      cleaningUsers.value = true
      try {
        const response = await fetch('/api/admin/cleanup/users', {
          method: 'POST',
          credentials: 'include'
        })
        const data = await response.json()
        if (data.success) {
          alert('All non-admin users cleaned up successfully')
        } else {
          alert(data.message || 'Failed to cleanup')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        cleaningUsers.value = false
      }
    }
    
    const cleanupAll = async () => {
      if (!confirm('Are you sure you want to delete ALL data? This cannot be undone!')) return
      if (!confirm('This is your last warning. Are you absolutely sure?')) return
      cleaningAll.value = true
      try {
        const response = await fetch('/api/admin/cleanup', {
          method: 'POST',
          credentials: 'include'
        })
        const data = await response.json()
        if (data.success) {
          alert('All data cleaned up successfully')
        } else {
          alert(data.errors?.join(', ') || 'Failed to cleanup')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        cleaningAll.value = false
      }
    }

    const cleanupTaskResults = async () => {
      if (!confirm('Are you sure you want to delete all task results?')) return
      cleaningTaskResults.value = true
      try {
        const response = await fetch('/api/admin/cleanup/task-results', {
          method: 'POST',
          credentials: 'include'
        })
        const data = await response.json()
        if (data.success) {
          alert('All task results cleaned up successfully')
        } else {
          alert(data.message || 'Failed to cleanup')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        cleaningTaskResults.value = false
      }
    }

    const cleanupSessions = async () => {
      cleaningSessions.value = true
      try {
        const response = await fetch('/api/admin/cleanup/sessions', {
          method: 'POST',
          credentials: 'include'
        })
        const data = await response.json()
        if (data.success) {
          const count = data.data?.count || 0
          alert(`Cleaned up ${count} expired session(s)`)
        } else {
          alert(data.message || 'Failed to cleanup')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        cleaningSessions.value = false
      }
    }

    const cleanupAllSessions = async () => {
      if (!confirm('Are you sure you want to delete all sessions? This will force logout all users.')) return
      cleaningAllSessions.value = true
      try {
        const response = await fetch('/api/admin/cleanup/all-sessions', {
          method: 'POST',
          credentials: 'include'
        })
        const data = await response.json()
        if (data.success) {
          const count = data.data?.count || 0
          alert(`Cleaned up all ${count} session(s)`)
        } else {
          alert(data.message || 'Failed to cleanup')
        }
      } catch (error) {
        alert('An error occurred')
      } finally {
        cleaningAllSessions.value = false
      }
    }

    onMounted(() => {
      fetchConfig()
      
      // 监听 activeTab 变化，当切换到 users 标签页时自动加载用户
      watch(activeTab, (newTab) => {
        if (newTab === 'users') {
          searchUsers()
        }
      })
    })
    
    return {
      activeTab,
      config,
      saving,
      cleaningTasks,
      cleaningTaskResults,
      cleaningResources,
      cleaningSessions,
      cleaningAllSessions,
      cleaningUsers,
      cleaningAll,
      testingConnection,
      testingAI,
      connectionTestResult,
      aiTestResult,
      userSearch,
      users,
      selectedUser,
      newPassword,
      editingEventReceiverIP,
      originalEventReceiverIP,
      fetchConfig,
      saveAIConfig,
      saveAzurePAT,
      saveEventReceiverIP,
      saveLogRetention,
      saveAIConcurrency,
      saveAIRequestPoolSize,
      cancelEditEventReceiverIP,
      testEventReceiverConnection,
      testAIConfig,
      searchUsers,
      showPasswordModal,
      updatePassword,
      cleanupTasks,
      cleanupResources,
      cleanupTaskResults,
      cleanupSessions,
      cleanupAllSessions,
      cleanupUsers,
      cleanupAll,
      // 分页相关
      currentPage,
      pageSize,
      totalUsers,
      changePage,
      changePageSize,
      nextPage,
      prevPage
    }
  }
}
</script>

<style scoped>
.console-page {
  max-width: 1000px;
  margin: 0 auto;
}

.tab-content {
  padding: 1.5rem 0;
}

.config-section {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid var(--border);
}

.config-section:last-child {
  border-bottom: none;
}

.config-section h3 {
  margin-bottom: 1rem;
  color: var(--text-primary);
  font-weight: 600;
}

.config-display {
  background: var(--bg-main);
  padding: 1rem;
  border-radius: var(--radius-sm);
}

.config-display .config-item {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.config-display .config-label {
  font-weight: 500;
  color: var(--text-secondary);
}

.config-display .config-value {
  flex: 1;
  color: var(--text-primary);
}

.config-edit .button-group {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.button-group {
  display: flex;
  gap: 0.5rem;
}

.search-section {
  margin-bottom: 1.5rem;
}

.search-form {
  display: flex;
  gap: 0.5rem;
}

.search-form .form-control {
  flex: 1;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

.cleanup-section h3 {
  margin-bottom: 0.5rem;
}

.warning-text {
  color: var(--danger);
  margin-bottom: 1.5rem;
}

.cleanup-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
}

.cleanup-item {
  background: var(--bg-main);
  padding: 1.5rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}

.cleanup-item h4 {
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.cleanup-item p {
  color: var(--text-secondary);
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.cleanup-item.danger-zone {
  background: var(--bg-warning);
  border: 1px solid var(--warning);
}

.pagination {
  margin-top: 1.5rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.pagination-info {
  font-size: 0.9rem;
  color: var(--text-secondary);
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.page-info {
  font-size: 0.9rem;
  font-weight: 500;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
}

.page-size-selector select {
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
}

.form-text {
  display: block;
  margin-top: 0.25rem;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.test-btn {
  margin-left: 0.5rem;
}

.test-result {
  margin-top: 0.75rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
}

.test-result.success {
  background-color: var(--bg-success);
  color: var(--success);
  border: 1px solid var(--success);
}

.test-result.error {
  background-color: var(--bg-danger);
  color: var(--danger);
  border: 1px solid var(--danger);
}

.btn-info {
  background-color: var(--info);
  color: white;
  transition: background-color 0.2s, opacity 0.2s;
}

.btn-info:focus {
  outline: 2px solid var(--info);
  outline-offset: 2px;
}

.btn-info:hover:not(:disabled) {
  background-color: var(--info-dark);
}

.btn-info:disabled {
  background-color: var(--info-light);
  cursor: not-allowed;
  opacity: 0.65;
}
</style>
