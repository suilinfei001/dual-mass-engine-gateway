<template>
  <div class="events-page">
    <div v-if="!eventReceiverConfigured" class="alert alert-warning">
      <p><strong>⚠️ Event Receiver API 未配置</strong></p>
      <p>请联系管理员在控制台配置 Event Receiver 服务器地址后才能使用。</p>
      <p>配置路径：控制台 (Console) → Event Receiver IP</p>
    </div>
    
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Events</h2>
      </div>
      
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
      </div>
      
      <div v-else-if="events.length === 0" class="empty-state">
        <p>No events found</p>
      </div>
      
      <table v-else class="events-table">
        <thead>
          <tr>
            <th>Event ID</th>
            <th>Type</th>
            <th>Repository</th>
            <th>Branch</th>
            <th>Author</th>
            <th>Status</th>
            <th>Current Task</th>
            <th>Created</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="event in events" :key="event.id">
            <td>{{ event.event_id || event.id }}</td>
            <td>
              <span class="badge" :class="getTypeClass(event.event_type)">{{ event.event_type }}</span>
            </td>
            <td>{{ event.repo_name || event.repository }}</td>
            <td>{{ event.branch || '-' }}</td>
            <td>{{ getAuthor(event) }}</td>
            <td>
              <span class="badge" :class="getStatusClass(event.status || event.event_status)">
                {{ event.status || event.event_status || 'pending' }}
              </span>
            </td>
            <td>{{ getCurrentTask(event) }}</td>
            <td>{{ formatTime(event.created_at) }}</td>
            <td>
              <button class="btn btn-primary btn-sm" @click="openTasksModal(event)">
                View
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Tasks Modal -->
    <div v-if="showTasksModal" class="modal-overlay" @click="closeTasksModal">
      <div class="modal modal-large" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">Tasks - {{ selectedEvent?.repo_name || selectedEvent?.repository || 'Unknown' }}</h3>
          <button class="modal-close" @click="closeTasksModal">&times;</button>
        </div>

        <div class="modal-body">
          <div class="event-info">
            <div class="info-row">
              <span class="label">Event ID:</span>
              <span class="value">{{ selectedEvent?.event_id || selectedEvent?.id }}</span>
            </div>
            <div class="info-row">
              <span class="label">Type:</span>
              <span class="value">{{ selectedEvent?.event_type }}</span>
            </div>
            <div class="info-row">
              <span class="label">Branch:</span>
              <span class="value">{{ selectedEvent?.branch || '-' }}</span>
            </div>
          </div>

          <div v-if="loadingTasks" class="loading">
            <div class="spinner"></div>
          </div>

          <div v-else-if="tasks.length === 0" class="empty-state">
            <p>No tasks found for this event</p>
          </div>

          <div v-else class="tasks-list">
            <div v-for="task in tasks" :key="task.id" class="task-item">
              <div class="task-header">
                <span class="task-name">{{ task.task_name }}</span>
                <div class="task-actions">
                  <span class="badge" :class="getStatusClass(task.status)">
                    {{ task.status }}
                  </span>
                  <button
                    v-if="task.status === 'passed' || task.status === 'failed' || task.status === 'timeout' || task.status === 'cancelled'"
                    class="btn btn-secondary btn-sm"
                    @click="openTaskDetailsModal(task)"
                  >
                    Details
                  </button>
                </div>
              </div>
              <div class="task-details">
                <div class="detail-item">
                  <span class="label">Task ID:</span>
                  <span class="value">{{ task.task_id ? task.task_id.substring(0, 8) + '...' : '-' }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">Stage:</span>
                  <span class="value">{{ task.stage || '-' }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">Order:</span>
                  <span class="value">{{ task.execute_order }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">Start Time:</span>
                  <span class="value">{{ formatTime(task.start_time) }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">End Time:</span>
                  <span class="value">{{ formatTime(task.end_time) }}</span>
                </div>
              </div>
              <div v-if="task.error_message" class="task-error">
                <span class="error-icon">!</span>
                <span class="error-text">{{ task.error_message }}</span>
              </div>
              <div v-if="task.status === 'no_resource'" class="task-retry">
                <button
                  class="btn btn-primary btn-sm"
                  :disabled="retryingTaskId === task.id"
                  @click="retryAIMatching(task.id)"
                >
                  {{ retryingTaskId === task.id ? 'Retrying...' : 'Retry AI Matching' }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeTasksModal">Close</button>
        </div>
      </div>
    </div>

    <!-- Task Details Modal -->
    <div v-if="showTaskDetailsModal" class="modal-overlay" @click="closeTaskDetailsModal">
      <div class="modal modal-large" @click.stop>
        <div class="modal-header">
          <h3 class="modal-title">Task Details - {{ selectedTask?.task_name }}</h3>
          <div class="modal-header-actions">
            <span v-if="selectedTask?.analyzing" class="analyzing-indicator">
              <span class="analyzing-dot"></span>
              AI Analysis in progress...
            </span>
            <button
              class="btn btn-secondary btn-sm"
              @click="refreshTaskDetails"
              :disabled="refreshingTask || selectedTask?.analyzing"
              title="Reanalyze logs with AI"
            >
              <span v-if="refreshingTask">Starting...</span>
              <span v-else-if="selectedTask?.analyzing">Analyzing...</span>
              <span v-else>Reanalyze Logs</span>
            </button>
            <button class="modal-close" @click="closeTaskDetailsModal">&times;</button>
          </div>
        </div>

        <div class="modal-body">
          <div v-if="selectedTask" class="task-detail-full">
            <div class="detail-section">
              <h4>Basic Information</h4>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="label">Task ID:</span>
                  <span class="value">{{ selectedTask.task_id || '-' }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">Task Name:</span>
                  <span class="value">{{ selectedTask.task_name }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">Request URL:</span>
                  <span class="value url">{{ selectedTask.request_url || '-' }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">Execute Order:</span>
                  <span class="value">{{ selectedTask.execute_order }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">Start Time:</span>
                  <span class="value">{{ formatTime(selectedTask.start_time) }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">End Time:</span>
                  <span class="value">{{ formatTime(selectedTask.end_time) }}</span>
                </div>
                <div class="detail-item">
                  <span class="label">Status:</span>
                  <span class="value">
                    <span class="badge" :class="getStatusClass(selectedTask.status)">
                      {{ selectedTask.status }}
                    </span>
                  </span>
                </div>
                <div v-if="selectedTask.build_id" class="detail-item">
                  <span class="label">Build ID:</span>
                  <span class="value">{{ selectedTask.build_id }}</span>
                </div>
              </div>
            </div>

            <div v-if="selectedTask.error_message" class="detail-section error-section">
              <h4>Error Message</h4>
              <div class="error-box">{{ selectedTask.error_message }}</div>
            </div>

            <div v-if="selectedTask.results && selectedTask.results.length > 0" class="detail-section">
              <h4>Results</h4>
              <div class="results-list">
                <div v-for="result in selectedTask.results" :key="result.check_type || result.id" class="result-detail-item">
                  <div class="result-header">
                    <span class="result-check-type">{{ result.check_type }}</span>
                    <span class="badge" :class="getResultStatusClass(result.result)">
                      {{ result.result }}
                    </span>
                  </div>
                  <div v-if="result.extra && Object.keys(result.extra).length > 0" class="result-extra">
                    <span class="extra-label">Extra:</span>
                    <pre class="extra-content">{{ JSON.stringify(result.extra, null, 2) }}</pre>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="!selectedTask.results || selectedTask.results.length === 0" class="detail-section">
              <p class="no-results">No results available for this task</p>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeTaskDetailsModal">Close</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'

export default {
  name: 'Events',
  setup() {
    const events = ref([])
    const loading = ref(false)
    const eventReceiverConfigured = ref(true)
    const eventReceiverMessage = ref('')
    const showTasksModal = ref(false)
    const showTaskDetailsModal = ref(false)
    const selectedEvent = ref(null)
    const selectedTask = ref(null)
    const tasks = ref([])
    const loadingTasks = ref(false)
    const retryingTaskId = ref(null)
    const refreshingTask = ref(false)
    let pollingInterval = null

    const checkEventReceiverConfig = async () => {
      try {
        const response = await fetch('/api/config/event-receiver')
        const data = await response.json()
        eventReceiverConfigured.value = data.configured !== false
        eventReceiverMessage.value = data.message || ''
      } catch (error) {
        eventReceiverConfigured.value = false
        eventReceiverMessage.value = 'Failed to check config'
      }
    }
    
    const fetchEvents = async () => {
      if (!eventReceiverConfigured.value) {
        events.value = []
        loading.value = false
        return
      }
      
      loading.value = true
      try {
        const response = await fetch('/api/events')
        const data = await response.json()
        events.value = data.data || []
      } catch (error) {
        console.error('Failed to fetch events:', error)
      } finally {
        loading.value = false
      }
    }
    
    const openTasksModal = async (event) => {
      selectedEvent.value = event
      showTasksModal.value = true
      tasks.value = []
      loadingTasks.value = true
      
      try {
        const response = await fetch(`/api/events/${event.id}/tasks`)
        const data = await response.json()
        tasks.value = data.data || []
      } catch (error) {
        console.error('Failed to fetch tasks:', error)
      } finally {
        loadingTasks.value = false
      }
    }
    
    const closeTasksModal = () => {
      showTasksModal.value = false
      selectedEvent.value = null
      tasks.value = []
    }

    const openTaskDetailsModal = async (task) => {
      // Fetch fresh task details from API to get results
      try {
        const response = await fetch(`/api/tasks/${task.id}`, {
          credentials: 'include'
        })
        const data = await response.json()
        if (data.success && data.data) {
          selectedTask.value = data.data
        } else {
          // Fallback to using the task from list if API fails
          selectedTask.value = task
        }
      } catch (error) {
        console.error('Failed to fetch task details:', error)
        // Fallback to using the task from list if API fails
        selectedTask.value = task
      }
      showTaskDetailsModal.value = true
      // Start polling if task is in analyzing state
      if (selectedTask.value?.analyzing) {
        startPollingTaskDetails()
      }
    }

    const closeTaskDetailsModal = () => {
      showTaskDetailsModal.value = false
      selectedTask.value = null
      // Clear polling interval when modal is closed
      if (pollingInterval) {
        clearInterval(pollingInterval)
        pollingInterval = null
      }
    }

    const refreshTaskDetails = async () => {
      if (!selectedTask.value || refreshingTask.value) return

      refreshingTask.value = true
      try {
        // Call the reanalyze API to trigger new AI analysis
        const response = await fetch(`/api/tasks/${selectedTask.value.id}/reanalyze`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          credentials: 'include'
        })
        const data = await response.json()

        if (data.success) {
          // Analysis started in background - start polling for updates
          if (selectedTask.value) {
            selectedTask.value.analyzing = true
          }
          startPollingTaskDetails()
        } else {
          alert('Analysis failed: ' + (data.message || 'Unknown error'))
        }
      } catch (error) {
        console.error('Failed to reanalyze task logs:', error)
        alert('Failed to start analysis: ' + error.message)
      } finally {
        refreshingTask.value = false
      }
    }

    const pollTaskDetails = async () => {
      if (!selectedTask.value) return

      try {
        const response = await fetch(`/api/tasks/${selectedTask.value.id}`, {
          credentials: 'include'
        })
        const data = await response.json()

        if (data.success && data.data) {
          const updatedTask = data.data
          // Update selected task with latest data
          selectedTask.value = updatedTask

          // If analysis is complete, stop polling
          if (!updatedTask.analyzing) {
            if (pollingInterval) {
              clearInterval(pollingInterval)
              pollingInterval = null
            }
          }
        }
      } catch (error) {
        console.error('Failed to poll task details:', error)
      }
    }

    const startPollingTaskDetails = () => {
      // Clear any existing interval
      if (pollingInterval) {
        clearInterval(pollingInterval)
      }
      // Start polling every 5 seconds
      pollingInterval = setInterval(() => {
        pollTaskDetails()
      }, 5000)
    }

    const getResultStatusClass = (result) => {
      const resultLower = result?.toLowerCase() || ''
      if (resultLower === 'pass' || resultLower === 'passed') {
        return 'badge-success'
      } else if (resultLower === 'fail' || resultLower === 'failed') {
        return 'badge-danger'
      } else if (resultLower === 'timeout') {
        return 'badge-warning'
      } else if (resultLower === 'cancelled' || resultLower === 'canceled') {
        return 'badge-secondary'
      } else if (resultLower === 'skipped') {
        return 'badge-secondary'
      } else if (resultLower === 'running') {
        return 'badge-info'
      }
      return 'badge-secondary'
    }

    const retryAIMatching = async (taskId) => {
      retryingTaskId.value = taskId
      try {
        const response = await fetch(`/api/tasks/${taskId}/retry-ai`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          credentials: 'include'
        })
        const data = await response.json()

        if (data.success) {
          alert(`AI matching successful! Matched resource: ${data.data.resource_name}`)
          // Refresh tasks for the current event
          if (selectedEvent.value) {
            await openTasksModal(selectedEvent.value)
          }
        } else if (data.code === 'AI_NOT_CONFIGURED') {
          alert('AI service is not configured.\n\nPlease ask the administrator to configure AI settings:\n1. Go to Admin Console\n2. Navigate to AI Configuration\n3. Enter AI IP, Model, and Token')
        } else {
          alert(data.message || 'AI matching failed')
        }
      } catch (error) {
        console.error('Failed to retry AI matching:', error)
        alert('Failed to retry AI matching')
      } finally {
        retryingTaskId.value = null
      }
    }

    const getStatusClass = (status) => {
      const classes = {
        'pending': 'badge-warning',
        'processing': 'badge-warning',
        'running': 'badge-info',
        'completed': 'badge-success',
        'passed': 'badge-success',
        'failed': 'badge-danger',
        'timeout': 'badge-danger',
        'cancelled': 'badge-secondary',
        'skipped': 'badge-secondary',
        'no_resource': 'badge-warning'
      }
      return classes[status] || 'badge-secondary'
    }

    const getTypeClass = (type) => {
      const classes = {
        'push': 'badge-info',
        'pull_request': 'badge-primary'
      }
      return classes[type] || 'badge-secondary'
    }
    
    const getResultIcon = (result) => {
      return result === 'pass' || result === 'passed' ? '+' : '-'
    }
    
    const formatTime = (time) => {
      if (!time) return '-'
      return new Date(time).toLocaleString()
    }

    const getAuthor = (event) => {
      return event.author || event.pusher || '-'
    }
    
    const getCurrentTask = (event) => {
      const status = event.status || event.event_status
      if (status === 'completed') {
        return event.last_task || event.last_task_name || event.current_task || event.current_task_name || '-'
      } else {
        return event.current_task || event.current_task_name || '-'
      }
    }
    
    onMounted(async () => {
      await checkEventReceiverConfig()
      if (eventReceiverConfigured.value) {
        await fetchEvents()
      }
    })
    
    return {
      events,
      loading,
      eventReceiverConfigured,
      eventReceiverMessage,
      showTasksModal,
      showTaskDetailsModal,
      selectedEvent,
      selectedTask,
      tasks,
      loadingTasks,
      retryingTaskId,
      refreshingTask,
      openTasksModal,
      closeTasksModal,
      openTaskDetailsModal,
      closeTaskDetailsModal,
      refreshTaskDetails,
      retryAIMatching,
      getStatusClass,
      getTypeClass,
      getResultIcon,
      getResultStatusClass,
      formatTime,
      getAuthor,
      getCurrentTask
    }
  }
}
</script>

<style scoped>
.events-page {
  max-width: 1400px;
  margin: 0 auto;
}

.events-table {
  width: 100%;
  border-collapse: collapse;
}

.events-table th,
.events-table td {
  padding: 0.875rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

.events-table th {
  background: var(--bg-main);
  font-weight: 600;
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
}

.events-table tr:hover {
  background: var(--bg-main);
}

.events-table tr:last-child td {
  border-bottom: none;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-large {
  max-width: 900px;
  max-height: 80vh;
  background: var(--bg-card);
  border-radius: var(--radius);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: var(--shadow-lg);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  background: #F8FAFC;
  color: #171717;
  border-bottom: 1px solid #E2E8F0;
}

.modal-title {
  margin: 0;
  font-size: 18px;
}

.modal-header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.modal-header-actions .btn {
  background: white;
  border: 1px solid #E2E8F0;
  color: #171717;
}

.modal-header-actions .btn:hover:not(:disabled) {
  background: #F1F5F9;
}

.modal-header-actions .btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.analyzing-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: rgba(13, 110, 253, 0.2);
  border-radius: 4px;
  font-size: 0.9rem;
  color: #6ea8fe;
}

.analyzing-dot {
  width: 8px;
  height: 8px;
  background: #6ea8fe;
  border-radius: 50%;
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.5;
    transform: scale(1.2);
  }
}

.modal-close {
  background: none;
  border: none;
  color: #64748B;
  font-size: 28px;
  cursor: pointer;
  line-height: 1;
}

.modal-body {
  padding: 24px;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  padding: 16px 24px;
  border-top: 1px solid #e2e8f0;
  text-align: right;
}

.event-info {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e2e8f0;
}

.info-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-row .label {
  font-size: 12px;
  color: #64748b;
  text-transform: uppercase;
}

.info-row .value {
  font-weight: 500;
}

.tasks-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-item {
  background: #f8fafc;
  border-radius: 8px;
  padding: 16px;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.task-name {
  font-weight: 600;
  font-size: 14px;
}

.task-details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.detail-item .label {
  font-size: 11px;
  color: #64748b;
}

.detail-item .value {
  font-size: 13px;
  font-weight: 500;
}

.task-urls {
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
}

.url-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: 8px;
}

.url-item .label {
  font-size: 11px;
  color: #64748b;
}

.url-item .url {
  font-size: 11px;
  word-break: break-all;
  color: #3b82f6;
}

.url-item .cancel-url {
  color: #f59e0b;
}

.task-error {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 12px;
  padding: 8px 12px;
  background: #fef2f2;
  border-radius: 6px;
}

.error-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: #ef4444;
  color: white;
  border-radius: 50%;
  font-weight: bold;
  font-size: 12px;
  flex-shrink: 0;
}

.error-text {
  font-size: 12px;
  color: #991b1b;
}

.task-results {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
}

.results-title {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  margin-bottom: 8px;
}

.result-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background: white;
  border-radius: 6px;
  margin-bottom: 6px;
  font-size: 13px;
}

.result-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  font-weight: bold;
  font-size: 14px;
}

.result-icon.pass,
.result-icon.passed {
  background: #dcfce7;
  color: #166534;
}

.result-icon.fail,
.result-icon.failed {
  background: #fee2e2;
  color: #991b1b;
}

.result-name {
  flex: 1;
}

.result-status {
  color: #64748b;
  font-size: 12px;
}

.badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 500;
}

.badge-primary {
  background: #fce7f3;
  color: #be185d;
}

.badge-info {
  background: #dbeafe;
  color: #1d4ed8;
}

.badge-success {
  background: #dcfce7;
  color: #166534;
}

.badge-warning {
  background: #fef3c7;
  color: #92400e;
}

.badge-danger {
  background: #fee2e2;
  color: #991b1b;
}

.badge-secondary {
  background: #f1f5f9;
  color: #475569;
}

.task-retry {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #e2e8f0;
  text-align: right;
}

.task-retry .btn-primary {
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  padding: 8px 16px;
  font-size: 13px;
  cursor: pointer;
  transition: background 0.2s;
}

.task-retry .btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.task-retry .btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

/* Task Actions */
.task-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

/* Task Details Modal Styles */
.task-detail-full {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.detail-section {
  background: #f8fafc;
  border-radius: 8px;
  padding: 16px;
}

.detail-section h4 {
  margin: 0 0 16px 0;
  font-size: 14px;
  font-weight: 600;
  color: #334155;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.error-section {
  background: #fef2f2;
}

.error-box {
  background: white;
  border-left: 3px solid #ef4444;
  padding: 12px;
  border-radius: 4px;
  font-size: 13px;
  color: #991b1b;
  word-break: break-word;
}

.value .url {
  color: #3b82f6;
  word-break: break-all;
  font-size: 12px;
}

/* Results List */
.results-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.result-detail-item {
  background: white;
  border-radius: 8px;
  padding: 12px;
  border-left: 3px solid #e2e8f0;
}

.result-detail-item .badge-success {
  border-left-color: #22c55e;
}

.result-detail-item .badge-danger {
  border-left-color: #ef4444;
}

.result-detail-item .badge-warning {
  border-left-color: #f59e0b;
}

.result-detail-item .badge-info {
  border-left-color: #3b82f6;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.result-check-type {
  font-weight: 600;
  font-size: 14px;
  color: #334155;
}

.result-extra {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #e2e8f0;
}

.extra-label {
  font-size: 11px;
  color: #64748b;
  display: block;
  margin-bottom: 4px;
}

.extra-content {
  background: #1e293b;
  color: #e2e8f0;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 12px;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.no-results {
  text-align: center;
  color: #64748b;
  padding: 20px;
}
</style>
