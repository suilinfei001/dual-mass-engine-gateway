<template>
  <div class="tasks-page">
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">Tasks</h2>
        <div class="filters">
          <select v-model="statusFilter" class="form-control" @change="fetchTasks">
            <option value="">All Status</option>
            <option value="pending">Pending</option>
            <option value="running">Running</option>
            <option value="passed">Passed</option>
            <option value="failed">Failed</option>
            <option value="timeout">Timeout</option>
            <option value="cancelled">Cancelled</option>
            <option value="skipped">Skipped</option>
            <option value="no_resource">No Resource</option>
          </select>
        </div>
      </div>
      
      <div v-if="loading" class="loading">
        <div class="spinner"></div>
      </div>
      
      <div v-else-if="tasks.length === 0" class="empty-state">
        <p>No tasks found</p>
      </div>
      
      <table v-else>
        <thead>
          <tr>
            <th>ID</th>
            <th>Task Name</th>
            <th>Event ID</th>
            <th>Stage</th>
            <th>Execute Order</th>
            <th>Status</th>
            <th>Start Time</th>
            <th>End Time</th>
            <th>Error</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="task in tasks" :key="task.id">
            <td>{{ task.id }}</td>
            <td>{{ task.task_name }}</td>
            <td>
              <router-link :to="`/events/${task.event_id}`">{{ task.event_id }}</router-link>
            </td>
            <td>{{ task.stage }}</td>
            <td>{{ task.execute_order }}</td>
            <td>
              <span class="badge" :class="getStatusClass(task.status)">
                {{ task.status }}
              </span>
            </td>
            <td>{{ formatTime(task.start_time) }}</td>
            <td>{{ formatTime(task.end_time) }}</td>
            <td>
              <span v-if="task.error_message" class="error-text" :title="task.error_message">
                {{ truncate(task.error_message, 30) }}
              </span>
              <span v-else>-</span>
            </td>
            <td>
              <button
                v-if="task.status === 'no_resource'"
                class="btn btn-sm btn-primary"
                :disabled="retryingTaskId === task.id"
                @click="retryAIMatching(task.id)"
              >
                {{ retryingTaskId === task.id ? 'Retrying...' : 'Retry AI' }}
              </button>
              <span v-else>-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'

export default {
  name: 'Tasks',
  setup() {
    const tasks = ref([])
    const loading = ref(false)
    const statusFilter = ref('')
    const retryingTaskId = ref(null)

    const fetchTasks = async () => {
      loading.value = true
      try {
        let url = '/api/tasks'
        if (statusFilter.value) {
          url += `?status=${statusFilter.value}`
        }
        const response = await fetch(url)
        const data = await response.json()
        tasks.value = data.data || []
      } catch (error) {
        console.error('Failed to fetch tasks:', error)
      } finally {
        loading.value = false
      }
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
          fetchTasks()
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
        'pending': 'badge-secondary',
        'running': 'badge-warning',
        'passed': 'badge-success',
        'failed': 'badge-danger',
        'timeout': 'badge-danger',
        'cancelled': 'badge-info',
        'skipped': 'badge-secondary',
        'no_resource': 'badge-warning'
      }
      return classes[status] || 'badge-secondary'
    }

    const formatTime = (time) => {
      if (!time) return '-'
      return new Date(time).toLocaleString()
    }

    const truncate = (str, length) => {
      if (!str) return ''
      return str.length > length ? str.substring(0, length) + '...' : str
    }

    onMounted(fetchTasks)

    return {
      tasks,
      loading,
      statusFilter,
      retryingTaskId,
      fetchTasks,
      retryAIMatching,
      getStatusClass,
      formatTime,
      truncate
    }
  }
}
</script>

<style scoped>
.tasks-page {
  max-width: 1200px;
  margin: 0 auto;
}

.filters {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.filters .form-control {
  width: 150px;
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: var(--text-secondary);
}

.error-text {
  color: var(--danger);
  cursor: help;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

.btn-primary {
  background-color: var(--primary);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary:hover:not(:disabled) {
  background-color: var(--primary-hover);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
