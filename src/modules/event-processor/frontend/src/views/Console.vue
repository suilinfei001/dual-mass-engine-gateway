<template>
  <div class="console-page">
    <div class="page-header">
      <h1 class="page-title">管理员控制台</h1>
      <p class="page-subtitle">管理系统配置、用户和清理数据</p>
    </div>

    <div class="console-card">
      <div class="tabs">
        <button
          class="tab"
          :class="{ active: activeTab === 'config' }"
          @click="activeTab = 'config'"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="tab-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
          系统配置
        </button>
        <button
          class="tab"
          :class="{ active: activeTab === 'users' }"
          @click="activeTab = 'users'"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="tab-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
          </svg>
          用户管理
        </button>
        <button
          class="tab"
          :class="{ active: activeTab === 'cleanup' }"
          @click="activeTab = 'cleanup'"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="tab-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          数据清理
        </button>
      </div>
      
      <div v-if="activeTab === 'config'" class="tab-content">
        <div class="config-section">
          <div class="section-header">
            <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
            <h3>AI 配置</h3>
          </div>
          <div class="config-form">
            <div class="form-group">
              <label for="aiIp">AI 服务器 IP</label>
              <input type="text" id="aiIp" v-model="config.ai_ip" class="form-control" placeholder="例如：192.168.112.38" />
            </div>
            <div class="form-group">
              <label for="aiModel">模型名称</label>
              <input type="text" id="aiModel" v-model="config.ai_model" class="form-control" placeholder="例如：Tome-pro" />
            </div>
            <div class="form-group">
              <label for="aiToken">API 令牌</label>
              <input type="password" id="aiToken" v-model="config.ai_token" class="form-control" />
            </div>
            <div class="button-group">
              <button class="btn btn-primary" @click="saveAIConfig" :disabled="saving">
                {{ saving ? '保存中...' : '保存 AI 配置' }}
              </button>
              <button
                class="btn btn-secondary"
                @click="testAIConfig"
                :disabled="testingAI || !config.ai_ip || !config.ai_model || !config.ai_token"
              >
                {{ testingAI ? '测试中...' : '测试连接' }}
              </button>
            </div>
            <div v-if="aiTestResult" class="test-result" :class="aiTestResult.success ? 'success' : 'error'">
              {{ aiTestResult.message }}
            </div>
          </div>
        </div>
        
        <div class="config-section">
          <div class="section-header">
            <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <h3>Azure DevOps 配置</h3>
          </div>
          <div class="config-form">
            <div class="form-group">
              <label for="azurePat">个人访问令牌 (PAT)</label>
              <input type="password" id="azurePat" v-model="config.azure_pat" class="form-control" />
            </div>
            <button class="btn btn-primary" @click="saveAzurePAT" :disabled="saving">
              {{ saving ? '保存中...' : '保存 Azure PAT' }}
            </button>
          </div>
        </div>
        
        <div class="config-section">
          <div class="section-header">
            <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <h3>事件接收器配置</h3>
          </div>
          <div class="config-form">
            <div v-if="!editingEventReceiverIP" class="config-display">
              <div class="config-item">
                <span class="config-label">事件接收器 IP：</span>
                <span class="config-value">{{ config.event_receiver_ip || '未配置' }}</span>
                <button class="btn btn-secondary btn-sm" @click="editingEventReceiverIP = true">修改</button>
                <button
                  class="btn btn-secondary btn-sm"
                  @click="testEventReceiverConnection"
                  :disabled="testingConnection"
                >
                  {{ testingConnection ? '测试中...' : '测试连接' }}
                </button>
              </div>
              <div v-if="connectionTestResult" class="test-result" :class="connectionTestResult.success ? 'success' : 'error'">
                {{ connectionTestResult.message }}
              </div>
            </div>
            <div v-else class="config-edit">
              <div class="form-group">
                <label for="eventReceiverIp">事件接收器 IP</label>
                <input type="text" id="eventReceiverIp" v-model="config.event_receiver_ip" class="form-control" placeholder="例如：http://10.4.111.141:5001" />
              </div>
              <div class="button-group">
                <button class="btn btn-secondary" @click="testEventReceiverConnection" :disabled="testingConnection">
                  {{ testingConnection ? '测试中...' : '测试' }}
                </button>
                <button class="btn btn-primary" @click="saveEventReceiverIP" :disabled="saving">
                  {{ saving ? '保存中...' : '保存' }}
                </button>
                <button class="btn btn-ghost" @click="cancelEditEventReceiverIP">取消</button>
              </div>
              <div v-if="connectionTestResult" class="test-result" :class="connectionTestResult.success ? 'success' : 'error'">
                {{ connectionTestResult.message }}
              </div>
            </div>
          </div>
        </div>

        <div class="config-section">
          <div class="section-header">
            <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <h3>日志保留配置</h3>
          </div>
          <div class="config-form">
            <div class="form-group">
              <label for="logRetentionDays">日志保留期限（天）</label>
              <input
                type="number"
                id="logRetentionDays"
                v-model.number="config.log_retention_days"
                class="form-control"
                min="1"
                max="30"
              />
              <small class="form-text">日志文件将保留这么多天后自动清理（1-30天，默认：7天）</small>
            </div>
            <button class="btn btn-primary" @click="saveLogRetention" :disabled="saving">
              {{ saving ? '保存中...' : '保存日志保留设置' }}
            </button>
          </div>
        </div>

        <div class="config-section">
          <div class="section-header">
            <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            <h3>AI 分析配置</h3>
          </div>
          <div class="config-form">
            <div class="form-group">
              <label for="aiConcurrency">AI 分析并发数（每个事件）</label>
              <input
                type="number"
                id="aiConcurrency"
                v-model.number="config.ai_concurrency"
                class="form-control"
                min="1"
                max="50"
              />
              <small class="form-text">每个事件同时分析的日志文件数量（1-50，默认：20）。</small>
            </div>
            <button class="btn btn-primary" @click="saveAIConcurrency" :disabled="saving">
              {{ saving ? '保存中...' : '保存 AI 并发配置' }}
            </button>
          </div>
        </div>

        <div class="config-section">
          <div class="section-header">
            <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            <h3>AI 请求池配置</h3>
          </div>
          <div class="config-form">
            <div class="form-group">
              <label for="aiRequestPoolSize">AI 请求池大小（全局）</label>
              <input
                type="number"
                id="aiRequestPoolSize"
                v-model.number="config.ai_request_pool_size"
                class="form-control"
                min="1"
                max="200"
              />
              <small class="form-text">所有事件的总并发 AI 请求数（1-200，默认：50）。</small>
            </div>
            <button class="btn btn-primary" @click="saveAIRequestPoolSize" :disabled="saving">
              {{ saving ? '保存中...' : '保存请求池大小' }}
            </button>
          </div>
        </div>

        <div class="config-section">
          <div class="section-header">
            <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
            <h3>CMP 配置</h3>
          </div>
          <div class="config-form">
            <div class="form-group">
              <label for="cmpAccessKey">CMP Access Key</label>
              <input type="text" id="cmpAccessKey" v-model="config.cmp_access_key" class="form-control" placeholder="输入 CMP Access Key" />
            </div>
            <div class="form-group">
              <label for="cmpSecretKey">CMP Secret Key</label>
              <input type="password" id="cmpSecretKey" v-model="config.cmp_secret_key" class="form-control" placeholder="输入 CMP Secret Key" />
            </div>
            <button class="btn btn-primary" @click="saveCMPConfig" :disabled="saving">
              {{ saving ? '保存中...' : '保存 CMP 配置' }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'users'" class="tab-content">
        <div class="search-section">
          <div class="search-form">
            <input
              type="text"
              v-model="userSearch"
              class="form-control"
              placeholder="搜索用户名或邮箱..."
              @keyup.enter="searchUsers"
            />
            <button class="btn btn-primary" @click="searchUsers">搜索</button>
          </div>
        </div>
        
        <div v-if="users.length > 0" class="table-container">
          <table class="data-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>用户名</th>
                <th>邮箱</th>
                <th>角色</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in users" :key="user.id">
                <td :title="user.id">{{ user.id }}</td>
                <td class="username-cell" :title="user.username">{{ user.username }}</td>
                <td :title="user.email">{{ user.email }}</td>
                <td :title="user.role">
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
                    修改密码
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          
          <div class="pagination">
            <div class="pagination-info">
              显示 {{ (currentPage - 1) * pageSize + 1 }} 到 {{ Math.min(currentPage * pageSize, totalUsers) }} 条，共 {{ totalUsers }} 条用户
            </div>
            <div class="pagination-controls">
              <button class="btn btn-sm btn-ghost" @click="prevPage" :disabled="currentPage === 1">上一页</button>
              <span class="page-info">第 {{ currentPage }} 页 / 共 {{ Math.ceil(totalUsers / pageSize) }} 页</span>
              <button class="btn btn-sm btn-ghost" @click="nextPage" :disabled="currentPage * pageSize >= totalUsers">下一页</button>
              <div class="page-size-selector">
                <label>每页:</label>
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
        
        <div v-else class="empty-state">
          <p>暂无用户数据</p>
        </div>
        
        <div v-if="selectedUser" class="modal-overlay" @click="selectedUser = null">
          <div class="modal" @click.stop>
            <div class="modal-header">
              <h3 class="modal-title">修改 {{ selectedUser.username }} 的密码</h3>
              <button class="modal-close" @click="selectedUser = null">&times;</button>
            </div>
            <form @submit.prevent="updatePassword">
              <div class="form-group">
                <label for="newPassword">新密码</label>
                <input type="password" id="newPassword" v-model="newPassword" class="form-control" required minlength="6" />
              </div>
              <div class="modal-footer">
                <button type="button" class="btn btn-ghost" @click="selectedUser = null">取消</button>
                <button type="submit" class="btn btn-primary" :disabled="saving">{{ saving ? '更新中...' : '更新密码' }}</button>
              </div>
            </form>
          </div>
        </div>
      </div>
      
      <div v-if="activeTab === 'cleanup'" class="tab-content">
        <div class="cleanup-section">
          <div class="section-header warning">
            <svg xmlns="http://www.w3.org/2000/svg" class="section-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <h3>数据清理</h3>
          </div>
          <p class="warning-text">警告：这些操作不可逆，请谨慎操作。</p>

          <h4 class="cleanup-group-title">事件处理器数据</h4>
          <div class="cleanup-actions">
            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理任务</h4>
                <p>删除系统中的所有任务（包括任务结果）。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupTasks" :disabled="cleaningTasks">{{ cleaningTasks ? '清理中...' : '清理任务' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理任务结果</h4>
                <p>删除系统中的所有任务结果（任务保留）。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupTaskResults" :disabled="cleaningTaskResults">{{ cleaningTaskResults ? '清理中...' : '清理任务结果' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理资源</h4>
                <p>删除系统中的所有可执行资源。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupResources" :disabled="cleaningResources">{{ cleaningResources ? '清理中...' : '清理资源' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理过期会话</h4>
                <p>删除系统中的所有过期会话。</p>
              </div>
              <button class="btn btn-warning" @click="cleanupSessions" :disabled="cleaningSessions">{{ cleaningSessions ? '清理中...' : '清理过期会话' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理所有会话</h4>
                <p>删除所有会话（强制所有用户退出登录）。</p>
              </div>
              <button class="btn btn-warning" @click="cleanupAllSessions" :disabled="cleaningAllSessions">{{ cleaningAllSessions ? '清理中...' : '清理所有会话' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理用户</h4>
                <p>删除系统中的所有非管理员用户。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupUsers" :disabled="cleaningUsers">{{ cleaningUsers ? '清理中...' : '清理用户' }}</button>
            </div>

            <div class="cleanup-item danger-zone">
              <div class="cleanup-content">
                <h4>清理全部数据</h4>
                <p>删除所有任务、资源和非管理员用户。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupAll" :disabled="cleaningAll">{{ cleaningAll ? '清理中...' : '清理全部数据' }}</button>
            </div>
          </div>

          <h4 class="cleanup-group-title">资源池数据</h4>
          <div class="cleanup-actions">
            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理 Testbed</h4>
                <p>删除资源池中的所有 Testbed。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupResourcePoolTestbeds" :disabled="cleaningResourcePoolTestbeds">{{ cleaningResourcePoolTestbeds ? '清理中...' : '清理 Testbed' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理分配记录</h4>
                <p>删除资源池中的所有分配记录。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupResourcePoolAllocations" :disabled="cleaningResourcePoolAllocations">{{ cleaningResourcePoolAllocations ? '清理中...' : '清理分配记录' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理资源实例</h4>
                <p>删除资源池中的所有资源实例。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupResourcePoolInstances" :disabled="cleaningResourcePoolInstances">{{ cleaningResourcePoolInstances ? '清理中...' : '清理资源实例' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理类别</h4>
                <p>删除资源池中的所有类别。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupResourcePoolCategories" :disabled="cleaningResourcePoolCategories">{{ cleaningResourcePoolCategories ? '清理中...' : '清理类别' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理配额策略</h4>
                <p>删除资源池中的所有配额策略。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupResourcePoolPolicies" :disabled="cleaningResourcePoolPolicies">{{ cleaningResourcePoolPolicies ? '清理中...' : '清理配额策略' }}</button>
            </div>

            <div class="cleanup-item">
              <div class="cleanup-content">
                <h4>清理任务记录</h4>
                <p>删除指定天数前的资源实例任务记录。</p>
                <div class="cleanup-input-group">
                  <input type="number" v-model.number="taskCleanupDays" class="form-control form-control-sm" min="1" max="365" placeholder="天数" />
                  <span class="input-suffix">天前的任务</span>
                </div>
              </div>
              <button class="btn btn-danger" @click="cleanupResourceInstanceTasks" :disabled="cleaningResourceInstanceTasks">{{ cleaningResourceInstanceTasks ? '清理中...' : '清理任务记录' }}</button>
            </div>

            <div class="cleanup-item danger-zone">
              <div class="cleanup-content">
                <h4>清理所有资源池数据</h4>
                <p>删除所有资源池数据。</p>
              </div>
              <button class="btn btn-danger" @click="cleanupResourcePoolAll" :disabled="cleaningResourcePoolAll">{{ cleaningResourcePoolAll ? '清理中...' : '清理所有资源池数据' }}</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted, watch } from 'vue'
import { useDialog } from '../composables/useDialog'

export default {
  name: 'Console',
  setup() {
    const dialog = useDialog()
    const activeTab = ref('config')
    const saving = ref(false)
    const testingConnection = ref(false)
    const testingAI = ref(false)
    const connectionTestResult = ref(null)
    const aiTestResult = ref(null)

    const cleaningTasks = ref(false)
    const cleaningTaskResults = ref(false)
    const cleaningResources = ref(false)
    const cleaningSessions = ref(false)
    const cleaningAllSessions = ref(false)
    const cleaningUsers = ref(false)
    const cleaningAll = ref(false)
    const cleaningResourcePoolTestbeds = ref(false)
    const cleaningResourcePoolAllocations = ref(false)
    const cleaningResourcePoolInstances = ref(false)
    const cleaningResourcePoolCategories = ref(false)
    const cleaningResourcePoolPolicies = ref(false)
    const cleaningResourceInstanceTasks = ref(false)
    const taskCleanupDays = ref(30)
    const cleaningResourcePoolAll = ref(false)

    const config = ref({
      ai_ip: '',
      ai_model: '',
      ai_token: '',
      azure_pat: '',
      event_receiver_ip: '',
      log_retention_days: 7,
      ai_concurrency: 20,
      ai_request_pool_size: 50,
      cmp_access_key: '',
      cmp_secret_key: ''
    })
    
    const editingEventReceiverIP = ref(false)
    const originalEventReceiverIP = ref('')
    
    const userSearch = ref('')
    const users = ref([])
    const selectedUser = ref(null)
    const newPassword = ref('')
    
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
            ai_request_pool_size: parseInt(data.data.ai_request_pool_size?.value || '50'),
            cmp_access_key: data.data.cmp_access_key?.value || '',
            cmp_secret_key: data.data.cmp_secret_key?.value || ''
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
          body: JSON.stringify({ ip: config.value.ai_ip, model: config.value.ai_model, token: config.value.ai_token })
        })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('AI 配置保存成功') } else { dialog.alertError(data.message || '保存失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { saving.value = false }
    }

    const testAIConfig = async () => {
      if (!config.value.ai_ip || !config.value.ai_model || !config.value.ai_token) {
        aiTestResult.value = { success: false, message: '请先填写所有 AI 配置字段' }
        return
      }
      testingAI.value = true
      aiTestResult.value = null
      try {
        const response = await fetch('/api/admin/config/ai/test', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ ip: config.value.ai_ip, model: config.value.ai_model, token: config.value.ai_token })
        })
        const data = await response.json()
        if (data.success) {
          aiTestResult.value = { success: true, message: `AI 连接成功！模型：${data.data?.model_name || config.value.ai_model}` }
        } else {
          aiTestResult.value = { success: false, message: data.message || 'AI 连接失败' }
        }
      } catch (error) {
        aiTestResult.value = { success: false, message: '网络错误：' + error.message }
      } finally { testingAI.value = false }
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
        if (data.success) { dialog.alertSuccess('Azure PAT 保存成功') } else { dialog.alertError(data.message || '保存失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { saving.value = false }
    }
    
    const saveEventReceiverIP = async () => {
      saving.value = true
      try {
        let ip = config.value.event_receiver_ip
        if (ip && !ip.startsWith('http://') && !ip.startsWith('https://')) { ip = 'http://' + ip }
        const response = await fetch('/api/admin/config/event-receiver', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ ip: ip })
        })
        const data = await response.json()
        if (data.success) {
          editingEventReceiverIP.value = false
          dialog.alertSuccess('事件接收器 IP 保存成功')
        } else { dialog.alertError(data.message || '保存失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { saving.value = false }
    }
    
    const cancelEditEventReceiverIP = () => {
      config.value.event_receiver_ip = originalEventReceiverIP.value
      editingEventReceiverIP.value = false
      connectionTestResult.value = null
    }

    const testEventReceiverConnection = async () => {
      const ip = config.value.event_receiver_ip
      if (!ip) { connectionTestResult.value = { success: false, message: '请先输入事件接收器 IP' }; return }
      let testUrl = ip
      if (!testUrl.startsWith('http://') && !testUrl.startsWith('https://')) { testUrl = 'http://' + testUrl }
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
          connectionTestResult.value = { success: true, message: `连接成功！找到 ${data.data?.event_count || 0} 个事件。` }
        } else {
          connectionTestResult.value = { success: false, message: data.message || '连接失败' }
        }
      } catch (error) { connectionTestResult.value = { success: false, message: '网络错误：' + error.message } }
      finally { testingConnection.value = false }
    }

    const saveLogRetention = async () => {
      const days = config.value.log_retention_days
      if (days < 1 || days > 30) { dialog.alertWarning('日志保留天数必须在 1 到 30 之间'); return }
      saving.value = true
      try {
        const response = await fetch('/api/admin/config/log-retention', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ days: days })
        })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess(`日志保留期限已设置为 ${days} 天`) } else { dialog.alertError(data.message || '保存失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { saving.value = false }
    }

    const saveAIConcurrency = async () => {
      const concurrency = config.value.ai_concurrency
      if (concurrency < 1 || concurrency > 50) { dialog.alertWarning('AI 并发数必须在 1 到 50 之间'); return }
      saving.value = true
      try {
        const response = await fetch('/api/admin/config/ai-concurrency', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ concurrency: concurrency })
        })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess(`AI 并发数已设置为 ${concurrency}`) } else { dialog.alertError(data.message || '保存失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { saving.value = false }
    }

    const saveAIRequestPoolSize = async () => {
      const poolSize = config.value.ai_request_pool_size
      const concurrency = config.value.ai_concurrency
      if (poolSize < 1 || poolSize > 200) { dialog.alertWarning('AI 请求池大小必须在 1 到 200 之间'); return }
      if (poolSize <= concurrency) { dialog.alertWarning(`AI 请求池大小 (${poolSize}) 必须大于 AI 并发数 (${concurrency})`); return }
      saving.value = true
      try {
        const response = await fetch('/api/admin/config/ai-request-pool-size', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ poolSize: poolSize })
        })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess(`AI 请求池大小已设置为 ${poolSize}`) } else { dialog.alertError(data.message || '保存失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { saving.value = false }
    }

    const saveCMPConfig = async () => {
      saving.value = true
      try {
        const response = await fetch('/api/admin/config/cmp', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ access_key: config.value.cmp_access_key, secret_key: config.value.cmp_secret_key })
        })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('CMP 配置保存成功') } else { dialog.alertError(data.message || '保存失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { saving.value = false }
    }

    const searchUsers = async () => {
      try {
        const response = await fetch(`/api/admin/users?keyword=${encodeURIComponent(userSearch.value)}&page=${currentPage.value}&pageSize=${pageSize.value}`, { credentials: 'include' })
        const data = await response.json()
        users.value = data.data || []
        totalUsers.value = data.total || 0
      } catch (error) { console.error('Failed to search users:', error) }
    }
    
    const showPasswordModal = (user) => { selectedUser.value = user; newPassword.value = '' }
    
    const changePageSize = (size) => { pageSize.value = size; currentPage.value = 1; searchUsers() }
    const nextPage = () => { if (currentPage.value * pageSize.value < totalUsers.value) { currentPage.value++; searchUsers() } }
    const prevPage = () => { if (currentPage.value > 1) { currentPage.value--; searchUsers() } }
    
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
        if (data.success) { dialog.alertSuccess('密码更新成功'); selectedUser.value = null } else { dialog.alertError(data.message || '更新失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { saving.value = false }
    }
    
    const cleanupTasks = async () => {
      const confirmed = await dialog.confirm('确定要删除所有任务吗？')
      if (!confirmed) return
      cleaningTasks.value = true
      try {
        const response = await fetch('/api/admin/cleanup/tasks', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有任务清理成功') } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { cleaningTasks.value = false }
    }

    const cleanupResources = async () => {
      const confirmed = await dialog.confirm('确定要删除所有资源吗？')
      if (!confirmed) return
      cleaningResources.value = true
      try {
        const response = await fetch('/api/admin/cleanup/resources', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有资源清理成功') } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { cleaningResources.value = false }
    }

    const cleanupUsers = async () => {
      const confirmed = await dialog.confirm('确定要删除所有非管理员用户吗？')
      if (!confirmed) return
      cleaningUsers.value = true
      try {
        const response = await fetch('/api/admin/cleanup/users', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有非管理员用户清理成功') } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { cleaningUsers.value = false }
    }

    const cleanupAll = async () => {
      const confirmed1 = await dialog.confirmDanger('确定要删除所有数据吗？此操作不可撤销！')
      if (!confirmed1) return
      const confirmed2 = await dialog.confirmDanger('这是最后的警告。您确定要继续吗？')
      if (!confirmed2) return
      cleaningAll.value = true
      try {
        const response = await fetch('/api/admin/cleanup', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有数据清理成功') } else { dialog.alertError(data.errors?.join(', ') || '清理失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { cleaningAll.value = false }
    }

    const cleanupTaskResults = async () => {
      const confirmed = await dialog.confirm('确定要删除所有任务结果吗？')
      if (!confirmed) return
      cleaningTaskResults.value = true
      try {
        const response = await fetch('/api/admin/cleanup/task-results', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有任务结果清理成功') } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { cleaningTaskResults.value = false }
    }

    const cleanupSessions = async () => {
      cleaningSessions.value = true
      try {
        const response = await fetch('/api/admin/cleanup/sessions', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess(`已清理 ${data.data?.count || 0} 个过期会话`) } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { cleaningSessions.value = false }
    }

    const cleanupAllSessions = async () => {
      const confirmed = await dialog.confirm('确定要删除所有会话吗？这将强制所有用户退出登录。')
      if (!confirmed) return
      cleaningAllSessions.value = true
      try {
        const response = await fetch('/api/admin/cleanup/all-sessions', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess(`已清理所有 ${data.data?.count || 0} 个会话`) } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误') } finally { cleaningAllSessions.value = false }
    }

    const cleanupResourcePoolTestbeds = async () => {
      const confirmed = await dialog.confirm('确定要删除所有 Testbed 吗？')
      if (!confirmed) return
      cleaningResourcePoolTestbeds.value = true
      try {
        const response = await fetch('/api/admin/cleanup/resource-pool/testbeds', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有 Testbed 清理成功') } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误：' + error.message) } finally { cleaningResourcePoolTestbeds.value = false }
    }

    const cleanupResourcePoolAllocations = async () => {
      const confirmed = await dialog.confirm('确定要删除所有分配记录吗？')
      if (!confirmed) return
      cleaningResourcePoolAllocations.value = true
      try {
        const response = await fetch('/api/admin/cleanup/resource-pool/allocations', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有分配记录清理成功') } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误：' + error.message) } finally { cleaningResourcePoolAllocations.value = false }
    }

    const cleanupResourcePoolInstances = async () => {
      const confirmed = await dialog.confirm('确定要删除所有资源实例吗？')
      if (!confirmed) return
      cleaningResourcePoolInstances.value = true
      try {
        const response = await fetch('/api/admin/cleanup/resource-pool/resource-instances', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有资源实例清理成功') } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误：' + error.message) } finally { cleaningResourcePoolInstances.value = false }
    }

    const cleanupResourcePoolCategories = async () => {
      const confirmed = await dialog.confirm('确定要删除所有类别吗？')
      if (!confirmed) return
      cleaningResourcePoolCategories.value = true
      try {
        const response = await fetch('/api/admin/cleanup/resource-pool/categories', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有类别清理成功') } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误：' + error.message) } finally { cleaningResourcePoolCategories.value = false }
    }

    const cleanupResourcePoolPolicies = async () => {
      const confirmed = await dialog.confirm('确定要删除所有配额策略吗？')
      if (!confirmed) return
      cleaningResourcePoolPolicies.value = true
      try {
        const response = await fetch('/api/admin/cleanup/resource-pool/quota-policies', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有配额策略清理成功') } else { dialog.alertError(data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误：' + error.message) } finally { cleaningResourcePoolPolicies.value = false }
    }

    const cleanupResourceInstanceTasks = async () => {
      const days = taskCleanupDays.value || 30
      const confirmed = await dialog.confirm(`确定要删除 ${days} 天前的所有资源实例任务记录吗？`)
      if (!confirmed) return
      cleaningResourceInstanceTasks.value = true
      try {
        const { adminAPI } = await import('../api/resourcePool')
        const result = await adminAPI.cleanupOldTasks({ days })
        if (result.success) { dialog.alertSuccess(`任务记录清理成功，删除了 ${result.deleted || 0} 条记录`) } else { dialog.alertError(result.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误：' + error.message) } finally { cleaningResourceInstanceTasks.value = false }
    }

    const cleanupResourcePoolAll = async () => {
      const confirmed1 = await dialog.confirmDanger('确定要删除所有资源池数据吗？此操作不可撤销！')
      if (!confirmed1) return
      const confirmed2 = await dialog.confirmDanger('这是最后的警告。您确定要继续吗？')
      if (!confirmed2) return
      cleaningResourcePoolAll.value = true
      try {
        const response = await fetch('/api/admin/cleanup/resource-pool/all', { method: 'POST', credentials: 'include' })
        const data = await response.json()
        if (data.success) { dialog.alertSuccess('所有资源池数据清理成功') } else { dialog.alertError(data.errors?.join(', ') || data.message || '清理失败') }
      } catch (error) { dialog.alertError('发生错误：' + error.message) } finally { cleaningResourcePoolAll.value = false }
    }

    onMounted(() => {
      fetchConfig()
      watch(activeTab, (newTab) => { if (newTab === 'users') { searchUsers() } })
    })
    
    return {
      activeTab, config, saving, testingConnection, testingAI, connectionTestResult, aiTestResult,
      cleaningTasks, cleaningTaskResults, cleaningResources, cleaningSessions, cleaningAllSessions,
      cleaningUsers, cleaningAll, cleaningResourcePoolTestbeds, cleaningResourcePoolAllocations,
      cleaningResourcePoolInstances, cleaningResourcePoolCategories, cleaningResourcePoolPolicies,
      cleaningResourceInstanceTasks, taskCleanupDays, cleanupResourceInstanceTasks, cleaningResourcePoolAll,
      userSearch, users, selectedUser, newPassword, editingEventReceiverIP, originalEventReceiverIP,
      fetchConfig, saveAIConfig, saveAzurePAT, saveEventReceiverIP, saveLogRetention, saveAIConcurrency,
      saveAIRequestPoolSize, saveCMPConfig, cancelEditEventReceiverIP, testEventReceiverConnection,
      testAIConfig, searchUsers, showPasswordModal, updatePassword, cleanupTasks, cleanupResources,
      cleanupTaskResults, cleanupSessions, cleanupAllSessions, cleanupUsers, cleanupAll,
      cleanupResourcePoolTestbeds, cleanupResourcePoolAllocations, cleanupResourcePoolInstances,
      cleanupResourcePoolCategories, cleanupResourcePoolPolicies, cleanupResourcePoolAll,
      currentPage, pageSize, totalUsers, changePageSize, nextPage, prevPage
    }
  }
}
</script>

<style scoped>
.console-page {
  max-width: 1100px;
  margin: 0 auto;
  padding: 2rem 1rem;
}

.page-header {
  margin-bottom: 2rem;
}

.page-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: #0C4A6E;
  margin-bottom: 0.375rem;
}

.page-subtitle {
  color: #64748B;
  font-size: 0.95rem;
}

.console-card {
  background: white;
  border-radius: 1rem;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  overflow: hidden;
}

.tabs {
  display: flex;
  background: #F8FAFC;
  border-bottom: 1px solid #E2E8F0;
  padding: 0 1rem;
}

.tab {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 1.25rem;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: #64748B;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab:hover {
  color: #0C4A6E;
}

.tab.active {
  color: #0EA5E9;
  border-bottom-color: #0EA5E9;
}

.tab-icon {
  width: 18px;
  height: 18px;
}

.tab-content {
  padding: 1.5rem;
}

.config-section {
  margin-bottom: 2rem;
  padding-bottom: 2rem;
  border-bottom: 1px solid #E2E8F0;
}

.config-section:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
}

.section-header.warning {
  color: #DC2626;
}

.section-icon {
  width: 22px;
  height: 22px;
  color: #0EA5E9;
}

.section-header.warning .section-icon {
  color: #DC2626;
}

.config-section h3 {
  font-size: 1rem;
  font-weight: 600;
  color: #0C4A6E;
  margin: 0;
}

.config-form {
  padding-left: 2.75rem;
}

.form-group {
  margin-bottom: 1rem;
  max-width: 400px;
}

.form-group label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.375rem;
}

.form-control {
  width: 100%;
  padding: 0.625rem 0.875rem;
  border: 1px solid #D1D5DB;
  border-radius: 0.5rem;
  font-size: 0.9rem;
  transition: all 0.2s ease;
  background: #F9FAFB;
}

.form-control:focus {
  outline: none;
  border-color: #0EA5E9;
  box-shadow: 0 0 0 3px rgba(14, 165, 233, 0.15);
  background: white;
}

.form-text {
  display: block;
  margin-top: 0.375rem;
  font-size: 0.8125rem;
  color: #6B7280;
}

.button-group {
  display: flex;
  gap: 0.5rem;
  margin-top: 1rem;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.btn-primary {
  background: linear-gradient(135deg, #0EA5E9, #0284C7);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(14, 165, 233, 0.3);
}

.btn-secondary {
  background: white;
  color: #374151;
  border: 1px solid #D1D5DB;
}

.btn-secondary:hover:not(:disabled) {
  background: #F3F4F6;
}

.btn-ghost {
  background: transparent;
  color: #6B7280;
}

.btn-ghost:hover {
  color: #374151;
  background: #F3F4F6;
}

.btn-danger {
  background: #EF4444;
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: #DC2626;
}

.btn-warning {
  background: #F59E0B;
  color: white;
}

.btn-warning:hover:not(:disabled) {
  background: #D97706;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.test-result {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.75rem;
  padding: 0.625rem 0.875rem;
  border-radius: 0.5rem;
  font-size: 0.875rem;
}

.test-result.success {
  background: #F0FDF4;
  color: #16A34A;
  border: 1px solid #BBF7D0;
}

.test-result.error {
  background: #FEF2F2;
  color: #DC2626;
  border: 1px solid #FECACA;
}

.config-display {
  padding-left: 2.75rem;
}

.config-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.config-label {
  font-weight: 500;
  color: #6B7280;
}

.config-value {
  flex: 1;
  color: #0C4A6E;
}

.search-section {
  margin-bottom: 1.5rem;
}

.search-form {
  display: flex;
  gap: 0.5rem;
  max-width: 400px;
}

.search-form .form-control {
  flex: 1;
}

.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.data-table th,
.data-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid #E2E8F0;
}

.data-table td {
  font-size: 0.9rem;
  color: #374151;
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

.data-table th {
  background: #F8FAFC;
  font-weight: 600;
  font-size: 0.8125rem;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.025em;
  white-space: nowrap;
}

.username-cell {
  font-weight: 500;
  color: #0C4A6E;
}

.badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
}

.badge-danger {
  background: #FEE2E2;
  color: #DC2626;
}

.badge-info {
  background: #DBEAFE;
  color: #2563EB;
}

.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid #E2E8F0;
  flex-wrap: wrap;
  gap: 1rem;
}

.pagination-info {
  font-size: 0.875rem;
  color: #6B7280;
}

.pagination-controls {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.page-info {
  font-size: 0.875rem;
  color: #64748B;
}

.page-size-selector {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.page-size-selector label {
  font-size: 0.875rem;
  color: #6B7280;
}

.page-size-selector select {
  padding: 0.375rem 0.5rem;
  border: 1px solid #D1D5DB;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  background: white;
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: #9CA3AF;
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
}

.modal {
  background: white;
  border-radius: 1rem;
  width: 90%;
  max-width: 400px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid #E2E8F0;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #0C4A6E;
  margin: 0;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #9CA3AF;
  cursor: pointer;
  line-height: 1;
}

.modal-close:hover {
  color: #6B7280;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #E2E8F0;
}

.cleanup-section h3 {
  margin-bottom: 0.5rem;
}

.warning-text {
  color: #DC2626;
  margin-bottom: 1.5rem;
  padding-left: 2.75rem;
}

.cleanup-group-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.025em;
  margin-bottom: 1rem;
  margin-top: 1.5rem;
}

.cleanup-group-title:first-of-type {
  margin-top: 0;
}

.cleanup-actions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1rem;
}

.cleanup-item {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 1.25rem;
  background: #F8FAFC;
  border-radius: 0.75rem;
  border: 1px solid #E2E8F0;
}

.cleanup-item.danger-zone {
  background: #FEF2F2;
  border-color: #FECACA;
}

.cleanup-content h4 {
  font-size: 0.95rem;
  font-weight: 600;
  color: #0C4A6E;
  margin: 0 0 0.25rem 0;
}

.cleanup-content p {
  font-size: 0.8125rem;
  color: #64748B;
  margin: 0;
}

.cleanup-input-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.cleanup-input-group .form-control-sm {
  width: 80px;
}

.input-suffix {
  font-size: 0.8125rem;
  color: #6B7280;
}

@media (max-width: 768px) {
  .tabs {
    overflow-x: auto;
  }
  
  .tab {
    padding: 0.75rem 1rem;
    white-space: nowrap;
  }
  
  .pagination {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .cleanup-actions {
    grid-template-columns: 1fr;
  }
}
</style>
