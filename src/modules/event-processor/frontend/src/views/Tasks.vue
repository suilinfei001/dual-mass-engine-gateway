<template>
  <div class="tasks-page">
    <div class="card">
      <div class="card-header">
        <h2 class="card-title">任务</h2>
        <div class="filters">
          <select v-model="statusFilter" class="form-control" @change="fetchTasks">
            <option value="">所有状态</option>
            <option value="pending">待处理</option>
            <option value="running">运行中</option>
            <option value="passed">通过</option>
            <option value="failed">失败</option>
            <option value="timeout">超时</option>
            <option value="cancelled">已取消</option>
            <option value="skipped">已跳过</option>
            <option value="no_resource">无资源</option>
          </select>
        </div>
      </div>

      <div v-if="loading" class="loading">
        <div class="spinner"></div>
      </div>

      <div v-else-if="tasks.length === 0" class="empty-state">
        <p>没有找到任务</p>
      </div>

      <table v-else>
        <thead>
          <tr>
            <th>ID</th>
            <th>任务名称</th>
            <th>事件 ID</th>
            <th>阶段</th>
            <th>执行顺序</th>
            <th>状态</th>
            <th>开始时间</th>
            <th>结束时间</th>
            <th>错误</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="task in tasks" :key="task.id">
            <td :title="task.id">{{ task.id }}</td>
            <td :title="task.task_name">{{ task.task_name }}</td>
            <td :title="task.event_id">
              <router-link :to="`/events/${task.event_id}`">{{ task.event_id }}</router-link>
            </td>
            <td :title="task.stage">{{ task.stage }}</td>
            <td :title="task.execute_order">{{ task.execute_order }}</td>
            <td :title="task.status">
              <span class="badge" :class="getStatusClass(task.status)">
                {{ task.status }}
              </span>
            </td>
            <td :title="formatTime(task.start_time)">{{ formatTime(task.start_time) }}</td>
            <td :title="formatTime(task.end_time)">{{ formatTime(task.end_time) }}</td>
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
import { ref, onMounted, inject } from 'vue'
import { useDialog } from '../composables/useDialog'

export default {
  name: 'Tasks',
  setup() {
    const dialog = useDialog()

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
          dialog.alertSuccess(`AI 匹配成功！\n\n匹配资源: ${data.data.resource_name}`)
          fetchTasks()
        } else if (data.code === 'AI_NOT_CONFIGURED') {
          dialog.alertWarning('AI 服务未配置。\n\n请联系管理员配置 AI 设置：\n1. 进入管理员控制台\n2. 导航到 AI 配置\n3. 输入 IP、模型和令牌')
        } else {
          dialog.alertError(data.message || 'AI 匹配失败')
        }
      } catch (error) {
        console.error('Failed to retry AI matching:', error)
        dialog.alertError('重试 AI 匹配失败')
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
