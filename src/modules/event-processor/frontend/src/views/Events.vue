<template>
  <div class="events-page">
    <div v-if="!eventReceiverConfigured" class="alert alert-warning">
      <svg xmlns="http://www.w3.org/2000/svg" class="alert-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
      </svg>
      <div class="alert-content">
        <p class="alert-title">Event Receiver API 未配置</p>
        <p class="alert-desc">请联系管理员在控制台配置 Event Receiver 服务器地址后才能使用。</p>
        <p class="alert-path">配置路径：控制台 (Console) → Event Receiver IP</p>
      </div>
    </div>
    
    <div class="card">
      <div class="card-header">
        <div class="card-title-group">
          <svg xmlns="http://www.w3.org/2000/svg" class="card-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
          <h2 class="card-title">事件处理</h2>
        </div>
        <button class="btn-refresh" @click="fetchEvents" :disabled="loading">
          <svg xmlns="http://www.w3.org/2000/svg" class="btn-icon" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>刷新</span>
        </button>
      </div>

      <div v-if="loading" class="loading">
        <div class="spinner"></div>
        <p class="loading-text">加载中...</p>
      </div>

      <div v-else-if="events.length === 0" class="empty-state">
        <svg xmlns="http://www.w3.org/2000/svg" class="empty-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
        </svg>
        <p>暂无事件数据</p>
      </div>

      <div v-else class="table-container">
        <table class="events-table">
          <thead>
            <tr>
              <th>事件 ID</th>
              <th>类型</th>
              <th>仓库</th>
              <th>分支</th>
              <th>作者</th>
              <th>状态</th>
              <th>当前任务</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="event in events" :key="event.id">
              <td :title="`#${event.event_id || event.id}`">
                <span class="event-id">#{{ event.event_id || event.id }}</span>
              </td>
              <td :title="event.event_type">
                <span class="badge" :class="getTypeClass(event.event_type)">
                  <svg v-if="event.event_type === 'push'" xmlns="http://www.w3.org/2000/svg" class="badge-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 12h14M12 5l7 7-7 7" />
                  </svg>
                  <svg v-else xmlns="http://www.w3.org/2000/svg" class="badge-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                  {{ event.event_type }}
                </span>
              </td>
              <td :title="event.repo_name || event.repository">
                <span class="repo-name">{{ event.repo_name || event.repository }}</span>
              </td>
              <td :title="event.branch || '-'">
                <span class="branch-name">
                  <svg xmlns="http://www.w3.org/2000/svg" class="branch-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
                  </svg>
                  {{ event.branch || '-' }}
                </span>
              </td>
              <td :title="getAuthor(event)">
                <span class="author-name">{{ getAuthor(event) }}</span>
              </td>
              <td :title="event.status || event.event_status || 'pending'">
                <span class="badge" :class="getStatusClass(event.status || event.event_status)">
                  {{ event.status || event.event_status || 'pending' }}
                </span>
              </td>
              <td :title="getCurrentTask(event)">
                <span class="task-name">{{ getCurrentTask(event) }}</span>
              </td>
              <td :title="formatTime(event.created_at)">
                <span class="time-text">{{ formatTime(event.created_at) }}</span>
              </td>
              <td>
                <button class="btn btn-primary btn-sm" @click="openTasksModal(event)">
                  <svg xmlns="http://www.w3.org/2000/svg" class="btn-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                  查看
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="showTasksModal" class="modal-overlay" @click="closeTasksModal">
      <div class="modal modal-large" @click.stop>
        <div class="modal-header">
          <div class="modal-title-group">
            <svg xmlns="http://www.w3.org/2000/svg" class="modal-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
            </svg>
            <h3 class="modal-title">任务列表</h3>
          </div>
          <span class="modal-subtitle">{{ selectedEvent?.repo_name || selectedEvent?.repository || '未知' }}</span>
          <button class="modal-close" @click="closeTasksModal">
            <svg xmlns="http://www.w3.org/2000/svg" class="close-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="modal-body">
          <div class="event-info-card">
            <div class="info-item">
              <span class="info-label">事件 ID</span>
              <span class="info-value">{{ selectedEvent?.event_id || selectedEvent?.id }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">类型</span>
              <span class="info-value">{{ selectedEvent?.event_type }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">分支</span>
              <span class="info-value">{{ selectedEvent?.branch || '-' }}</span>
            </div>
          </div>

          <div v-if="loadingTasks" class="loading">
            <div class="spinner"></div>
            <p class="loading-text">加载任务中...</p>
          </div>

          <div v-else-if="tasks.length === 0" class="empty-state">
            <svg xmlns="http://www.w3.org/2000/svg" class="empty-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
            </svg>
            <p>此事件没有找到任务</p>
          </div>

          <div v-else class="tasks-list">
            <div v-for="task in tasks" :key="task.id" class="task-card">
              <div class="task-header">
                <div class="task-name-group">
                  <svg xmlns="http://www.w3.org/2000/svg" class="task-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span class="task-name">{{ task.task_name }}</span>
                </div>
                <div class="task-actions">
                  <span class="badge" :class="getStatusClass(task.status)">
                    {{ task.status }}
                  </span>
                  <button
                    v-if="task.status === 'passed' || task.status === 'failed' || task.status === 'timeout' || task.status === 'cancelled'"
                    class="btn btn-secondary btn-sm"
                    @click="openTaskDetailsModal(task)"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" class="btn-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                    详情
                  </button>
                </div>
              </div>
              <div class="task-details">
                <div class="detail-item">
                  <svg xmlns="http://www.w3.org/2000/svg" class="detail-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                  </svg>
                  <span class="detail-label">任务 ID：</span>
                  <span class="detail-value">{{ task.task_id ? task.task_id.substring(0, 8) + '...' : '-' }}</span>
                </div>
                <div class="detail-item">
                  <svg xmlns="http://www.w3.org/2000/svg" class="detail-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                  </svg>
                  <span class="detail-label">阶段：</span>
                  <span class="detail-value">{{ task.stage || '-' }}</span>
                </div>
                <div class="detail-item">
                  <svg xmlns="http://www.w3.org/2000/svg" class="detail-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
                  </svg>
                  <span class="detail-label">顺序：</span>
                  <span class="detail-value">{{ task.execute_order }}</span>
                </div>
                <div class="detail-item">
                  <svg xmlns="http://www.w3.org/2000/svg" class="detail-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span class="detail-label">开始：</span>
                  <span class="detail-value">{{ formatTime(task.start_time) }}</span>
                </div>
                <div class="detail-item">
                  <svg xmlns="http://www.w3.org/2000/svg" class="detail-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <span class="detail-label">结束：</span>
                  <span class="detail-value">{{ formatTime(task.end_time) }}</span>
                </div>
              </div>
              <div v-if="task.error_message" class="task-error">
                <svg xmlns="http://www.w3.org/2000/svg" class="error-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <span class="error-text">{{ task.error_message }}</span>
              </div>
              <div v-if="task.status === 'no_resource'" class="task-retry">
                <button
                  class="btn btn-primary btn-sm"
                  :disabled="retryingTaskId === task.id"
                  @click="retryAIMatching(task.id)"
                >
                  <svg v-if="retryingTaskId !== task.id" xmlns="http://www.w3.org/2000/svg" class="btn-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                  </svg>
                  {{ retryingTaskId === task.id ? '重试中...' : '重试 AI 匹配' }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeTasksModal">
            <svg xmlns="http://www.w3.org/2000/svg" class="btn-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
            关闭
          </button>
        </div>
      </div>
    </div>

    <div v-if="showTaskDetailsModal" class="modal-overlay" @click="closeTaskDetailsModal">
      <div class="modal modal-large" @click.stop>
        <div class="modal-header">
          <div class="modal-title-group">
            <svg xmlns="http://www.w3.org/2000/svg" class="modal-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <h3 class="modal-title">任务详情</h3>
          </div>
          <span class="modal-subtitle">{{ selectedTask?.task_name }}</span>
          <div class="modal-header-actions">
            <span v-if="selectedTask?.analyzing" class="analyzing-indicator">
              <span class="analyzing-dot"></span>
              AI 分析中...
            </span>
            <button
              class="btn btn-secondary btn-sm"
              @click="refreshTaskDetails"
              :disabled="refreshingTask || selectedTask?.analyzing"
              title="使用 AI 重新分析日志"
            >
              <svg v-if="!refreshingTask && !selectedTask?.analyzing" xmlns="http://www.w3.org/2000/svg" class="btn-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              <span v-if="refreshingTask">启动中...</span>
              <span v-else-if="selectedTask?.analyzing">分析中...</span>
              <span v-else>重新分析日志</span>
            </button>
            <button class="modal-close" @click="closeTaskDetailsModal">
              <svg xmlns="http://www.w3.org/2000/svg" class="close-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <div class="modal-body">
          <div v-if="selectedTask" class="task-detail-full">
            <div class="detail-section">
              <h4 class="section-title">
                <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                基本信息
              </h4>
              <div class="detail-grid">
                <div class="detail-item">
                  <span class="detail-label">任务 ID</span>
                  <span class="detail-value mono">{{ selectedTask.task_id || '-' }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">任务名称</span>
                  <span class="detail-value">{{ selectedTask.task_name }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">请求 URL</span>
                  <span class="detail-value url">{{ selectedTask.request_url || '-' }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">执行顺序</span>
                  <span class="detail-value">{{ selectedTask.execute_order }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">开始时间</span>
                  <span class="detail-value">{{ formatTime(selectedTask.start_time) }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">结束时间</span>
                  <span class="detail-value">{{ formatTime(selectedTask.end_time) }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">状态</span>
                  <span class="detail-value">
                    <span class="badge" :class="getStatusClass(selectedTask.status)">
                      {{ selectedTask.status }}
                    </span>
                  </span>
                </div>
                <div v-if="selectedTask.build_id" class="detail-item">
                  <span class="detail-label">构建 ID</span>
                  <span class="detail-value">{{ selectedTask.build_id }}</span>
                </div>
              </div>
            </div>

            <div v-if="selectedTask.error_message" class="detail-section error-section">
              <h4 class="section-title">
                <svg xmlns="http://www.w3.org/2000/svg" class="section-icon error" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                错误信息
              </h4>
              <div class="error-box">{{ selectedTask.error_message }}</div>
            </div>

            <div v-if="selectedTask.results && selectedTask.results.length > 0" class="detail-section">
              <h4 class="section-title">
                <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                检查结果
              </h4>
              <div class="results-list">
                <div v-for="result in selectedTask.results" :key="result.check_type || result.id" class="result-card">
                  <div class="result-header">
                    <span class="result-check-type">{{ result.check_type }}</span>
                    <span class="badge" :class="getResultStatusClass(result.result)">
                      {{ result.result }}
                    </span>
                  </div>
                  <div v-if="result.extra && Object.keys(result.extra).length > 0" class="result-extra">
                    <span class="extra-label">额外信息</span>
                    <pre class="extra-content">{{ JSON.stringify(result.extra, null, 2) }}</pre>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="!selectedTask.results || selectedTask.results.length === 0" class="detail-section">
              <div class="no-results">
                <svg xmlns="http://www.w3.org/2000/svg" class="no-results-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <p>此任务暂无结果</p>
              </div>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button class="btn btn-secondary" @click="closeTaskDetailsModal">
            <svg xmlns="http://www.w3.org/2000/svg" class="btn-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
            关闭
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useDialog } from '../composables/useDialog'

export default {
  name: 'Events',
  setup() {
    const dialog = useDialog()
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
      try {
        const response = await fetch(`/api/tasks/${task.id}`, {
          credentials: 'include'
        })
        const data = await response.json()
        if (data.success && data.data) {
          selectedTask.value = data.data
        } else {
          selectedTask.value = task
        }
      } catch (error) {
        console.error('Failed to fetch task details:', error)
        selectedTask.value = task
      }
      showTaskDetailsModal.value = true
      if (selectedTask.value?.analyzing) {
        startPollingTaskDetails()
      }
    }

    const closeTaskDetailsModal = () => {
      showTaskDetailsModal.value = false
      selectedTask.value = null
      if (pollingInterval) {
        clearInterval(pollingInterval)
        pollingInterval = null
      }
    }

    const refreshTaskDetails = async () => {
      if (!selectedTask.value || refreshingTask.value) return

      refreshingTask.value = true
      try {
        const response = await fetch(`/api/tasks/${selectedTask.value.id}/reanalyze`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          credentials: 'include'
        })
        const data = await response.json()

        if (data.success) {
          if (selectedTask.value) {
            selectedTask.value.analyzing = true
          }
          startPollingTaskDetails()
        } else {
          dialog.alertError('分析失败：' + (data.message || '未知错误'))
        }
      } catch (error) {
        console.error('Failed to reanalyze task logs:', error)
        dialog.alertError('启动分析失败：' + error.message)
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
          selectedTask.value = updatedTask

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
      if (pollingInterval) {
        clearInterval(pollingInterval)
      }
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
          dialog.alertSuccess(`AI 匹配成功！匹配的资源：${data.data.resource_name}`)
          if (selectedEvent.value) {
            await openTasksModal(selectedEvent.value)
          }
        } else if (data.code === 'AI_NOT_CONFIGURED') {
          dialog.alertWarning('AI 服务未配置。\n\n请联系管理员配置 AI 设置：\n1. 前往管理员控制台\n2. 进入 AI 配置\n3. 输入 AI IP、模型和令牌')
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
      fetchEvents,
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

.alert {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.25rem 1.5rem;
  border-radius: 12px;
  margin-bottom: 1.5rem;
}

.alert-warning {
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
  border: 1px solid #F59E0B;
}

.alert-icon {
  width: 24px;
  height: 24px;
  color: #B45309;
  flex-shrink: 0;
  margin-top: 2px;
}

.alert-content {
  flex: 1;
}

.alert-title {
  font-weight: 600;
  color: #92400E;
  margin-bottom: 0.25rem;
}

.alert-desc {
  font-size: 0.875rem;
  color: #B45309;
  margin-bottom: 0.25rem;
}

.alert-path {
  font-size: 0.8125rem;
  color: #D97706;
}

.card {
  background: #FFFFFF;
  border-radius: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08), 0 4px 12px rgba(0, 0, 0, 0.05);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  border: 1px solid #E2E8F0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #E2E8F0;
}

.card-title-group {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.card-icon {
  width: 24px;
  height: 24px;
  color: #6366F1;
}

.card-title {
  font-size: 1.25rem;
  font-weight: 700;
  color: #1E1B4B;
  letter-spacing: -0.025em;
}

.btn-refresh {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: #F5F3FF;
  color: #6366F1;
  border: 1px solid #C7D2FE;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-refresh:hover:not(:disabled) {
  background: #EDE9FE;
  border-color: #A5B4FC;
}

.btn-refresh:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-icon {
  width: 16px;
  height: 16px;
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.table-container {
  overflow-x: auto;
  border-radius: 8px;
  border: 1px solid #E2E8F0;
}

.events-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 900px;
  table-layout: fixed;
}

.events-table th,
.events-table td {
  padding: 1rem 1.25rem;
  text-align: left;
}

.events-table td {
  /* Apply text overflow to all table cells */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.events-table th {
  background: linear-gradient(180deg, #F8FAFC 0%, #F1F5F9 100%);
  font-weight: 600;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #64748B;
  border-bottom: 2px solid #E2E8F0;
}

.events-table th:nth-child(1),
.events-table td:nth-child(1) {
  width: 80px;
}

.events-table th:nth-child(2),
.events-table td:nth-child(2) {
  width: 100px;
}

.events-table th:nth-child(3),
.events-table td:nth-child(3) {
  width: 200px;
  max-width: 200px;
}

.events-table th:nth-child(4),
.events-table td:nth-child(4) {
  width: 120px;
}

.events-table th:nth-child(5),
.events-table td:nth-child(5) {
  width: 100px;
}

.events-table th:nth-child(6),
.events-table td:nth-child(6) {
  width: 100px;
}

.events-table th:nth-child(7),
.events-table td:nth-child(7) {
  width: 140px;
}

.events-table th:nth-child(8),
.events-table td:nth-child(8) {
  width: 150px;
}

.events-table th:nth-child(9),
.events-table td:nth-child(9) {
  width: 80px;
}

.events-table td {
  border-bottom: 1px solid #F1F5F9;
  vertical-align: middle;
}

/* Ensure inline elements in table cells also truncate */
.events-table td span,
.events-table td a {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

.events-table tbody tr {
  transition: background-color 0.15s ease;
}

.events-table tbody tr:hover {
  background-color: #F8FAFC;
}

.events-table tbody tr:last-child td {
  border-bottom: none;
}

.event-id {
  font-weight: 600;
  color: #6366F1;
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.75rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.badge-icon {
  width: 14px;
  height: 14px;
}

.badge-success {
  background: linear-gradient(135deg, #D1FAE5 0%, #A7F3D0 100%);
  color: #047857;
}

.badge-danger {
  background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
  color: #B91C1C;
}

.badge-warning {
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
  color: #B45309;
}

.badge-info {
  background: linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%);
  color: #1D4ED8;
}

.badge-primary {
  background: linear-gradient(135deg, #EDE9FE 0%, #DDD6FE 100%);
  color: #6D28D9;
}

.badge-secondary {
  background: #F1F5F9;
  color: #475569;
}

.repo-name {
  font-weight: 500;
  color: #1E1B4B;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 180px;
}

.branch-name {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  color: #64748B;
  font-size: 0.875rem;
}

.branch-icon {
  width: 14px;
  height: 14px;
  color: #94A3B8;
}

.author-name {
  color: #475569;
}

.task-name {
  color: #475569;
  font-size: 0.875rem;
}

.time-text {
  color: #94A3B8;
  font-size: 0.8125rem;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn:focus {
  outline: 2px solid #6366F1;
  outline-offset: 2px;
}

.btn-primary {
  background: linear-gradient(135deg, #6366F1 0%, #4F46E5 100%);
  color: white;
  box-shadow: 0 1px 2px rgba(99, 102, 241, 0.3);
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #4F46E5 0%, #4338CA 100%);
  box-shadow: 0 2px 4px rgba(99, 102, 241, 0.4);
  transform: translateY(-1px);
}

.btn-secondary {
  background: #F1F5F9;
  color: #475569;
  border: 1px solid #E2E8F0;
}

.btn-secondary:hover:not(:disabled) {
  background: #E2E8F0;
  color: #1E1B4B;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  gap: 1rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid #E2E8F0;
  border-top-color: #6366F1;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.loading-text {
  color: #64748B;
  font-size: 0.875rem;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  color: #94A3B8;
}

.empty-icon {
  width: 64px;
  height: 64px;
  color: #CBD5E1;
  margin-bottom: 1rem;
}

.empty-state p {
  font-size: 0.9375rem;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
  padding: 1rem;
}

.modal {
  background: #FFFFFF;
  border-radius: 16px;
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  overflow: hidden;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  display: flex;
  flex-direction: column;
}

.modal-large {
  max-width: 900px;
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #E2E8F0;
  background: linear-gradient(180deg, #FFFFFF 0%, #F8FAFC 100%);
  flex-shrink: 0;
}

.modal-title-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.modal-icon {
  width: 20px;
  height: 20px;
  color: #6366F1;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 700;
  color: #1E1B4B;
}

.modal-subtitle {
  font-size: 0.875rem;
  color: #64748B;
  margin-left: auto;
}

.modal-header-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-left: auto;
}

.modal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  color: #94A3B8;
  transition: all 0.2s ease;
}

.modal-close:hover {
  background: #F1F5F9;
  color: #1E1B4B;
}

.close-icon {
  width: 20px;
  height: 20px;
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
  border-top: 1px solid #E2E8F0;
  background: #F8FAFC;
  flex-shrink: 0;
}

.event-info-card {
  display: flex;
  gap: 1.5rem;
  padding: 1rem 1.25rem;
  background: linear-gradient(135deg, #F5F3FF 0%, #EDE9FE 100%);
  border-radius: 10px;
  margin-bottom: 1.5rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-label {
  font-size: 0.75rem;
  color: #7C3AED;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-value {
  font-size: 0.9375rem;
  color: #1E1B4B;
  font-weight: 600;
}

.tasks-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.task-card {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
  border-radius: 12px;
  padding: 1.25rem;
  transition: all 0.2s ease;
}

.task-card:hover {
  border-color: #C7D2FE;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.1);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.task-name-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.task-icon {
  width: 20px;
  height: 20px;
  color: #6366F1;
}

.task-name {
  font-weight: 600;
  color: #1E1B4B;
}

.task-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.task-details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.detail-icon {
  width: 16px;
  height: 16px;
  color: #94A3B8;
  flex-shrink: 0;
}

.detail-label {
  font-size: 0.8125rem;
  color: #64748B;
}

.detail-value {
  font-size: 0.875rem;
  color: #1E1B4B;
}

.task-error {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  margin-top: 1rem;
  padding: 0.75rem 1rem;
  background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
  border-radius: 8px;
}

.error-icon {
  width: 18px;
  height: 18px;
  color: #DC2626;
  flex-shrink: 0;
  margin-top: 1px;
}

.error-text {
  font-size: 0.8125rem;
  color: #991B1B;
  line-height: 1.5;
}

.task-retry {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid #E2E8F0;
}

.analyzing-indicator {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.75rem;
  background: linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%);
  border-radius: 6px;
  font-size: 0.8125rem;
  color: #1D4ED8;
  font-weight: 500;
}

.analyzing-dot {
  width: 8px;
  height: 8px;
  background: #3B82F6;
  border-radius: 50%;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.8); }
}

.task-detail-full {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.detail-section {
  background: #F8FAFC;
  border-radius: 12px;
  padding: 1rem;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9375rem;
  font-weight: 600;
  color: #1E1B4B;
  margin-bottom: 1rem;
}

.section-icon {
  width: 18px;
  height: 18px;
  color: #6366F1;
}

.section-icon.error {
  color: #DC2626;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.75rem 1rem;
}

/* Responsive: 2 columns on tablets, 1 column on mobile */
@media (max-width: 768px) {
  .detail-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .detail-grid {
    grid-template-columns: 1fr;
  }
}

.detail-grid .detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  padding: 0.5rem;
  background: #FFFFFF;
  border-radius: 8px;
  min-height: 56px;
}

.detail-grid .detail-label {
  font-size: 0.75rem;
  color: #64748B;
  font-weight: 500;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.detail-grid .detail-value {
  font-size: 0.9375rem;
  color: #1E1B4B;
}

.detail-grid .detail-value.url {
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.8125rem;
  word-break: break-all;
}

.detail-grid .detail-value.mono {
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.8125rem;
}

.error-section {
  background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
  border: 1px solid #FCA5A5;
}

.error-box {
  background: #FFFFFF;
  border-radius: 8px;
  padding: 1rem;
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.8125rem;
  color: #991B1B;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}

.results-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.result-card {
  background: #FFFFFF;
  border: 1px solid #E2E8F0;
  border-radius: 10px;
  padding: 1rem;
  transition: all 0.2s ease;
}

.result-card:hover {
  border-color: #C7D2FE;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.result-check-type {
  font-weight: 600;
  color: #1E1B4B;
}

.result-extra {
  margin-top: 0.75rem;
  padding-top: 0.75rem;
  border-top: 1px solid #E2E8F0;
}

.extra-label {
  display: block;
  font-size: 0.75rem;
  color: #64748B;
  font-weight: 500;
  margin-bottom: 0.5rem;
}

.extra-content {
  background: #F8FAFC;
  border-radius: 6px;
  padding: 0.75rem;
  font-family: 'SF Mono', Monaco, 'Courier New', monospace;
  font-size: 0.75rem;
  color: #475569;
  overflow-x: auto;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}

.no-results {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: #94A3B8;
}

.no-results-icon {
  width: 48px;
  height: 48px;
  color: #CBD5E1;
  margin-bottom: 0.75rem;
}

.no-results p {
  font-size: 0.875rem;
}

@media (max-width: 768px) {
  .events-page {
    padding: 0;
  }
  
  .card {
    padding: 1rem;
    border-radius: 12px;
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
  
  .btn-refresh {
    width: 100%;
    justify-content: center;
  }
  
  .event-info-card {
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .task-details {
    grid-template-columns: 1fr;
  }
  
  .detail-grid {
    grid-template-columns: 1fr;
  }
  
  .modal {
    max-height: 100vh;
    border-radius: 0;
  }
  
  .modal-header {
    flex-wrap: wrap;
  }
  
  .modal-subtitle {
    order: 3;
    width: 100%;
    margin-left: 0;
    margin-top: 0.5rem;
  }
  
  .modal-header-actions {
    margin-left: auto;
  }
}

@media (prefers-reduced-motion: reduce) {
  .spinner,
  .analyzing-dot,
  .btn-icon.animate-spin {
    animation: none;
  }
  
  .btn,
  .task-card,
  .result-card,
  .events-table tbody tr {
    transition: none;
  }
}
</style>
