<template>
  <div class="resource-instance-page">
    <div class="page-header">
      <div class="header-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
        </svg>
      </div>
      <div class="header-content">
        <h1>{{ isAdmin ? '资源实例管理' : '物理机资源' }}</h1>
        <p>{{ isAdmin ? '管理虚拟机和物理机资源实例' : '查看和管理物理机资源' }}</p>
      </div>
      <button v-if="!isAdmin" class="btn btn-primary" @click="openCreateModal">
        <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
        添加资源
      </button>
    </div>

    <div class="filters-card">
      <div class="filters">
        <div class="filter-group">
          <label>状态</label>
          <select v-model="filters.status" class="filter-select" @change="fetchResources">
            <option value="">全部</option>
            <option value="pending">检查中</option>
            <option value="active">可用</option>
            <option value="unreachable">不可用</option>
          </select>
        </div>
        <div v-if="isAdmin" class="filter-group">
          <label>类型</label>
          <select v-model="filters.type" class="filter-select" @change="fetchResources">
            <option value="">全部</option>
            <option value="virtual_machine">虚拟机</option>
            <option value="physical_machine">物理机</option>
          </select>
        </div>
        <div class="filter-spacer"></div>
        <div class="filter-group search-group">
          <div class="search-input-wrapper">
            <svg class="search-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              v-model="filters.search"
              type="text"
              class="search-input"
              placeholder="搜索主机地址..."
              @input="debounceFetch"
            />
          </div>
        </div>
      </div>
    </div>

    <div class="table-container">
      <div v-if="loading" class="loading-container">
        <div class="spinner"></div>
        <span>加载中...</span>
      </div>

      <div v-else-if="resources.length === 0" class="empty-state">
        <div class="empty-icon">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
          </svg>
        </div>
        <h3>没有找到资源实例</h3>
        <p>尝试调整筛选条件或添加新资源</p>
      </div>

      <table v-else class="data-table">
        <thead>
          <tr>
            <th>主机</th>
            <th>类型</th>
            <th>SSH 配置</th>
            <th>快照 ID</th>
            <th>状态</th>
            <th>权限</th>
            <th>创建者</th>
            <th class="actions-col">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="resource in resources" :key="resource.uuid">
            <td :title="resource.ip_address || resource.host || '-'">
              <div class="host-cell">
                <svg class="host-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                </svg>
                <span class="host-address">{{ resource.ip_address || resource.host || '-' }}</span>
              </div>
            </td>
            <td :title="getTypeLabel(resource.instance_type || resource.resource_type)">
              <span class="type-badge" :class="getTypeClass(resource.instance_type || resource.resource_type)">
                {{ getTypeLabel(resource.instance_type || resource.resource_type) }}
              </span>
            </td>
            <td>
              <div class="ssh-config-cell">
                <span class="ssh-user">{{ resource.ssh_user || 'root' }}</span>
                <span class="ssh-separator">@</span>
                <span class="ssh-port">{{ resource.port || resource.ssh_port || '-' }}</span>
              </div>
            </td>
            <td :title="resource.snapshot_id || '-'">
              <span v-if="resource.snapshot_id" class="snapshot-badge">{{ resource.snapshot_id }}</span>
              <span v-else class="empty-value">-</span>
            </td>
            <td :title="getStatusLabel(resource.status)">
              <span class="status-badge" :class="getStatusClass(resource.status)">
                <span class="status-dot"></span>
                {{ getStatusLabel(resource.status) }}
              </span>
            </td>
            <td :title="resource.is_public ? '公开' : '私有'">
              <span v-if="resource.is_public" class="permission-badge public">公开</span>
              <span v-else class="permission-badge private">私有</span>
            </td>
            <td :title="resource.created_by || '-'">{{ resource.created_by || '-' }}</td>
            <td>
              <div class="action-buttons">
                <button
                  v-if="isAdmin || canEditResource(resource)"
                  class="action-btn action-btn-test"
                  @click="testConnection(resource)"
                  title="测试连接"
                  :disabled="testing === resource.uuid"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                  </svg>
                </button>
                <button
                  v-if="isAdmin"
                  class="action-btn action-btn-info"
                  @click="viewDetails(resource)"
                  title="查看详情"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </button>
                <button
                  v-if="isAdmin"
                  class="action-btn action-btn-tasks"
                  @click="viewTasks(resource)"
                  title="查看任务"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
                  </svg>
                </button>
                <button
                  v-if="isAdmin && isVirtualMachine(resource) && resource.snapshot_id"
                  class="action-btn action-btn-rollback"
                  @click="confirmRestoreSnapshot(resource)"
                  title="快照回滚"
                  :disabled="restoring === resource.uuid"
                >
                  <svg v-if="restoring !== resource.uuid" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6" />
                  </svg>
                  <svg v-else class="spin-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle class="spin-path" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
                  </svg>
                </button>
                <template v-else-if="canEditResource(resource)">
                  <button
                    class="action-btn action-btn-edit"
                    @click="editResource(resource)"
                    title="编辑"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button
                    v-if="resource.status === 'available' || resource.status === 'active'"
                    class="action-btn action-btn-warning"
                    @click="setMaintenance(resource)"
                    title="设为不可用"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
                    </svg>
                  </button>
                  <button
                    v-if="resource.status === 'unreachable'"
                    class="action-btn action-btn-success"
                    @click="setAvailable(resource)"
                    title="设为可用"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </button>
                  <button
                    class="action-btn action-btn-danger"
                    @click="deleteResource(resource)"
                    title="删除"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </template>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="pagination.total > 0" class="pagination">
        <button class="pagination-btn" :disabled="pagination.page <= 1" @click="changePage(pagination.page - 1)">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
          上一页
        </button>
        <span class="page-info">第 {{ pagination.page }} 页，共 {{ totalPages }} 页 ({{ pagination.total }} 条)</span>
        <button class="pagination-btn" :disabled="pagination.page >= totalPages" @click="changePage(pagination.page + 1)">
          下一页
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </div>
    </div>

    <div v-if="showCreateModal || showEditModal" class="modal-overlay" @click.self="closeModals">
      <div class="modal">
        <div class="modal-header">
          <h3 class="modal-title">{{ showEditModal ? '编辑资源实例' : '添加资源实例' }}</h3>
          <button class="modal-close" @click="closeModals">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <form @submit.prevent="saveResource">
          <div class="modal-body">
            <div class="form-section">
              <h4>基本信息</h4>
              <div class="form-row">
                <div class="form-group">
                  <label for="name">名称 <span class="required">*</span></label>
                  <input id="name" v-model="formData.name" type="text" class="form-control" required />
                </div>
                <div class="form-group">
                  <label for="type">资源类型 <span class="required">*</span></label>
                  <select id="type" v-model="formData.resource_type" class="form-control" required>
                    <option value="">选择类型</option>
                    <option value="virtual_machine">虚拟机</option>
                    <option value="physical_machine">物理机</option>
                  </select>
                </div>
              </div>
            </div>

            <div class="form-section">
              <h4>连接配置</h4>
              <div class="form-row">
                <div class="form-group">
                  <label for="host">主机地址 <span class="required">*</span></label>
                  <input id="host" v-model="formData.host" type="text" class="form-control" placeholder="192.168.1.100" required />
                </div>
                <div class="form-group">
                  <label for="sshPort">SSH 端口 <span class="required">*</span></label>
                  <input id="sshPort" v-model.number="formData.ssh_port" type="number" class="form-control" required />
                </div>
              </div>
              <div class="form-row">
                <div class="form-group">
                  <label for="sshUser">SSH 用户名 <span class="required">*</span></label>
                  <input id="sshUser" v-model="formData.ssh_user" type="text" class="form-control" placeholder="root" required />
                </div>
                <div class="form-group">
                  <label for="passwd">SSH 密码 <span class="required">*</span></label>
                  <input id="passwd" v-model="formData.passwd" type="password" class="form-control" placeholder="请输入 SSH 密码" required />
                </div>
              </div>
            </div>

            <div v-if="formData.resource_type === 'virtual_machine'" class="form-section">
              <h4>快照配置</h4>
              <div class="form-row">
                <div class="form-group">
                  <label for="snapshotId">快照 ID</label>
                  <input id="snapshotId" v-model="formData.snapshot_id" type="text" class="form-control" placeholder="如: snapshot-v1.0" />
                  <small>用于回滚的快照标识</small>
                </div>
                <div class="form-group">
                  <label for="snapshotInstanceUuid">快照实例ID</label>
                  <input id="snapshotInstanceUuid" v-model="formData.snapshot_instance_uuid" type="text" class="form-control" placeholder="如: ins-xxxxxxxx" />
                  <small>虚拟机实例的唯一标识</small>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="closeModals">取消</button>
            <button type="button" class="btn btn-outline" :disabled="testingConnection" @click="testNewConnection">
              <svg class="btn-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              {{ testingConnection ? '测试中...' : '测试连接' }}
            </button>
            <button type="submit" class="btn btn-primary" :disabled="submitting">
              {{ submitting ? '保存中...' : '保存' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="showDetailsModal" class="modal-overlay" @click.self="closeDetailsModal">
      <div class="modal detail-modal">
        <div class="modal-header detail-header">
          <div class="header-icon-section">
            <div class="detail-type-icon" :class="isVirtualMachine(selectedResource) ? 'type-vm' : 'type-pm'">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
              </svg>
            </div>
            <div class="header-text">
              <h3 class="modal-title">资源实例详情</h3>
              <span class="detail-subtitle">{{ selectedResource?.ip_address || selectedResource?.host || '-' }}</span>
            </div>
          </div>
          <button class="modal-close" @click="closeDetailsModal">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="modal-body detail-body">
          <!-- 状态卡片 -->
          <div class="detail-status-card">
            <div class="status-indicator">
              <span class="status-dot" :class="getStatusClass(selectedResource?.status)"></span>
              <span class="status-text">{{ getStatusLabel(selectedResource?.status) }}</span>
            </div>
            <div class="type-badge-large" :class="isVirtualMachine(selectedResource) ? 'type-vm' : 'type-pm'">
              {{ getTypeLabel(selectedResource?.instance_type || selectedResource?.resource_type) }}
            </div>
          </div>

          <!-- 详情分组 -->
          <div class="detail-sections">
            <!-- 连接信息 -->
            <div class="detail-section">
              <div class="detail-section-header">
                <svg class="section-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
                </svg>
                <h4 class="section-title">连接信息</h4>
              </div>
              <div class="detail-list">
                <div class="detail-row">
                  <span class="detail-key">主机地址</span>
                  <span class="detail-value mono">{{ selectedResource?.ip_address || selectedResource?.host || '-' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-key">SSH 端口</span>
                  <span class="detail-value">{{ selectedResource?.port || selectedResource?.ssh_port || '-' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-key">SSH 用户</span>
                  <span class="detail-value">{{ selectedResource?.ssh_user || 'root' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-key">SSH 密码</span>
                  <span class="detail-value mono masked">{{ selectedResource?.passwd || '-' }}</span>
                </div>
              </div>
            </div>

            <!-- 快照信息 -->
            <div v-if="selectedResource?.snapshot_id || selectedResource?.snapshot_instance_uuid" class="detail-section">
              <div class="detail-section-header">
                <svg class="section-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                </svg>
                <h4 class="section-title">快照信息</h4>
              </div>
              <div class="detail-list">
                <div v-if="selectedResource?.snapshot_id" class="detail-row">
                  <span class="detail-key">快照 ID</span>
                  <span class="detail-value mono snapshot">{{ selectedResource.snapshot_id }}</span>
                </div>
                <div v-if="selectedResource?.snapshot_instance_uuid" class="detail-row">
                  <span class="detail-key">快照实例</span>
                  <span class="detail-value mono">{{ selectedResource.snapshot_instance_uuid }}</span>
                </div>
              </div>
            </div>

            <!-- 元信息 -->
            <div class="detail-section">
              <div class="detail-section-header">
                <svg class="section-icon" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <h4 class="section-title">元信息</h4>
              </div>
              <div class="detail-list">
                <div class="detail-row">
                  <span class="detail-key">创建者</span>
                  <span class="detail-value">{{ selectedResource?.created_by || '-' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-key">创建时间</span>
                  <span class="detail-value">{{ formatTime(selectedResource?.created_at) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="closeDetailsModal">关闭</button>
        </div>
      </div>
    </div>

    <div v-if="showTasksModal" class="modal-overlay" @click.self="closeTasksModal">
      <div class="modal modal-large">
        <div class="modal-header">
          <h3 class="modal-title">资源实例任务</h3>
          <button class="modal-close" @click="closeTasksModal">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="resource-info-bar">
            <span class="resource-label">资源:</span>
            <span class="resource-value">{{ selectedResourceForTasks?.ip_address || selectedResourceForTasks?.host }}</span>
          </div>

          <div class="task-filters">
            <div class="filter-group">
              <label>状态</label>
              <select v-model="taskFilters.status" class="form-control" @change="fetchTasks">
                <option value="">所有状态</option>
                <option value="pending">等待中</option>
                <option value="running">执行中</option>
                <option value="completed">已完成</option>
                <option value="failed">失败</option>
                <option value="cancelled">已取消</option>
              </select>
            </div>
            <div class="filter-group">
              <label>类型</label>
              <select v-model="taskFilters.type" class="form-control" @change="fetchTasks">
                <option value="">所有类型</option>
                <option value="deploy">部署</option>
                <option value="rollback">回滚</option>
                <option value="health_check">健康检查</option>
              </select>
            </div>
          </div>

          <div v-if="tasksLoading" class="loading-container">
            <div class="spinner"></div>
            <span>加载中...</span>
          </div>

          <table v-else-if="tasks.length > 0" class="data-table">
            <thead>
              <tr>
                <th>任务类型</th>
                <th>状态</th>
                <th>触发来源</th>
                <th>开始时间</th>
                <th>执行时长</th>
                <th>结果</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="task in tasks" :key="task.uuid">
                <td>
                  <span class="task-type-badge" :class="getTaskTypeClass(task.task_type)">
                    {{ task.task_type_name || task.task_type }}
                  </span>
                </td>
                <td>
                  <span class="status-badge" :class="getTaskStatusClass(task.status)">
                    <span class="status-dot"></span>
                    {{ task.status_name || task.status }}
                  </span>
                </td>
                <td>{{ task.trigger_source_name || task.trigger_source }}</td>
                <td>{{ formatTime(task.started_at) }}</td>
                <td>{{ task.duration_display || '-' }}</td>
                <td>
                  <span v-if="task.success === true" class="result-success">成功</span>
                  <span v-else-if="task.success === false" class="result-failed">{{ task.error_code || '失败' }}</span>
                  <span v-else class="text-muted">-</span>
                </td>
                <td>
                  <button
                    v-if="task.error_message"
                    class="action-btn action-btn-info"
                    @click="showTaskError(task)"
                    title="查看错误"
                  >
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>

          <div v-else class="empty-state">
            <div class="empty-icon">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
              </svg>
            </div>
            <h3>暂无任务记录</h3>
          </div>

          <div v-if="taskPagination.total > 0" class="pagination">
            <button class="pagination-btn" :disabled="taskPagination.page <= 1" @click="changeTaskPage(taskPagination.page - 1)">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
              </svg>
              上一页
            </button>
            <span class="page-info">{{ taskPagination.page }} / {{ taskTotalPages }}</span>
            <button class="pagination-btn" :disabled="taskPagination.page >= taskTotalPages" @click="changeTaskPage(taskPagination.page + 1)">
              下一页
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </button>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="closeTasksModal">关闭</button>
        </div>
      </div>
    </div>

    <!-- 快照回滚确认对话框 -->
    <div v-if="showRestoreModal" class="modal-overlay" @click.self="showRestoreModal = false">
      <div class="modal modal-small">
        <div class="modal-header">
          <div class="modal-title-section">
            <svg class="modal-icon modal-icon-warning" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <h3 class="modal-title">快照回滚确认</h3>
          </div>
          <button class="modal-close" @click="showRestoreModal = false">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="restore-info">
            <p>确定要将虚拟机回滚到以下快照吗？</p>
            <div class="restore-details">
              <div class="restore-detail-item">
                <span class="restore-detail-label">主机:</span>
                <span class="restore-detail-value">{{ restoringResource?.ip_address || restoringResource?.host || '-' }}</span>
              </div>
              <div class="restore-detail-item">
                <span class="restore-detail-label">快照 ID:</span>
                <span class="restore-detail-value snapshot-id">{{ restoringResource?.snapshot_id || '-' }}</span>
              </div>
            </div>
            <p class="restore-warning">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              回滚操作需要几分钟时间，期间该资源将不可用
            </p>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="showRestoreModal = false">取消</button>
          <button type="button" class="btn btn-warning" @click="executeRestoreSnapshot" :disabled="restoring">
            <svg v-if="restoring" class="btn-spinner" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="spinner-circle" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" stroke-dasharray="32" stroke-dashoffset="32" />
            </svg>
            <span>{{ restoring ? '回滚中...' : '确认回滚' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { adminAPI, externalAPI } from '../../api/resourcePool'
import { useDialog } from '../../composables/useDialog'

export default {
  name: 'ResourceInstanceList',
  setup() {
    const dialog = useDialog()
    const resources = ref([])
    const loading = ref(false)
    const submitting = ref(false)
    const testing = ref(null)
    const restoring = ref(null)
    const showCreateModal = ref(false)
    const showEditModal = ref(false)
    const showDetailsModal = ref(false)
    const showTasksModal = ref(false)
    const showRestoreModal = ref(false)
    const editingResource = ref(null)
    const selectedResource = ref(null)
    const selectedResourceForTasks = ref(null)
    const restoringResource = ref(null)
    const tasks = ref([])
    const tasksLoading = ref(false)

    const currentUsername = computed(() => localStorage.getItem('username') || '')
    const isAdmin = computed(() => localStorage.getItem('userRole') === 'admin')

    const filters = ref({
      status: '',
      type: '',
      search: ''
    })

    const pagination = ref({
      page: 1,
      page_size: 20,
      total: 0
    })

    const taskFilters = ref({
      status: '',
      type: ''
    })

    const taskPagination = ref({
      page: 1,
      page_size: 10,
      total: 0
    })

    const formData = ref({
      name: '',
      resource_type: 'virtual_machine',
      host: '',
      ssh_port: 22,
      ssh_user: 'root',
      passwd: '',
      snapshot_id: '',
      snapshot_instance_uuid: ''
    })

    const testingConnection = ref(false)

    const testNewConnection = async () => {
      if (!formData.value.host || !formData.value.ssh_port || !formData.value.ssh_user) {
        dialog.alertWarning('请填写主机地址、SSH端口和用户名')
        return
      }
      testingConnection.value = true
      try {
        const result = await adminAPI.testConnection({
          host: formData.value.host,
          ssh_port: formData.value.ssh_port,
          ssh_user: formData.value.ssh_user,
          password: formData.value.passwd
        })
        if (result.success) {
          dialog.alertSuccess('连接成功！')
        } else {
          dialog.alertError('连接失败: ' + (result.message || '未知错误'))
        }
      } catch (error) {
        dialog.alertError('测试连接失败: ' + error.message)
      } finally {
        testingConnection.value = false
      }
    }

    const totalPages = computed(() => {
      return Math.ceil(pagination.value.total / pagination.value.page_size)
    })

    const taskTotalPages = computed(() => {
      return Math.ceil(taskPagination.value.total / taskPagination.value.page_size)
    })

    let debounceTimer = null

    const fetchResources = async () => {
      loading.value = true
      try {
        const params = {
          page: pagination.value.page,
          page_size: pagination.value.page_size,
          ...filters.value
        }
        Object.keys(params).forEach(key => {
          if (params[key] === '' || params[key] === null || params[key] === undefined) {
            delete params[key]
          }
        })

        if (!isAdmin.value) {
          params.resource_type = 'physical_machine'
        }

        let data
        if (isAdmin.value) {
          data = await adminAPI.getResourceInstances(params)
          resources.value = data.data || data.resource_instances || []
        } else {
          data = await externalAPI.getPublicResourceInstances(params)
          resources.value = data.data || data.instances || []
        }
        pagination.value.total = data.total || resources.value.length
      } catch (error) {
        console.error('Failed to fetch resources:', error)
        dialog.alertError('获取资源实例列表失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    const saveResource = async () => {
      submitting.value = true
      try {
        if (editingResource.value) {
          await adminAPI.updateResourceInstance(editingResource.value.uuid, formData.value)
          dialog.alertSuccess('更新成功')
        } else {
          await adminAPI.createResourceInstance(formData.value)
          dialog.alertSuccess('创建成功')
        }
        closeModals()
        fetchResources()
      } catch (error) {
        console.error('Failed to save resource:', error)
        dialog.alertError('保存失败: ' + error.message)
      } finally {
        submitting.value = false
      }
    }

    const editResource = (resource) => {
      editingResource.value = resource
      formData.value = {
        name: resource.ip_address || resource.name,
        resource_type: normalizeResourceType(resource.instance_type || resource.resource_type),
        host: resource.ip_address || resource.host,
        ssh_port: resource.port || resource.ssh_port || 22,
        ssh_user: resource.ssh_user || 'root',
        passwd: resource.passwd || '',
        snapshot_id: resource.snapshot_id || '',
        snapshot_instance_uuid: resource.snapshot_instance_uuid || ''
      }
      showEditModal.value = true
    }

    const deleteResource = async (resource) => {
      const confirmed = await dialog.confirm(`确定要删除资源实例 "${resource.ip_address || resource.name}" 吗？\n\n注意：如果该资源关联了 Testbed，将无法删除。`)
      if (!confirmed) return
      try {
        await adminAPI.deleteResourceInstance(resource.uuid)
        dialog.alertSuccess('删除成功')
        fetchResources()
      } catch (error) {
        dialog.alertError('删除失败: ' + error.message)
      }
    }

    const setMaintenance = async (resource) => {
      const confirmed = await dialog.confirm(`确定要将 "${resource.ip_address || resource.name}" 设为不可用吗？`)
      if (!confirmed) return
      try {
        await adminAPI.updateResourceInstance(resource.uuid, { status: 'unreachable' })
        fetchResources()
      } catch (error) {
        dialog.alertError('操作失败: ' + error.message)
      }
    }

    const setAvailable = async (resource) => {
      const confirmed = await dialog.confirm(`确定要将 "${resource.ip_address || resource.name}" 设为可用吗？`)
      if (!confirmed) return
      try {
        await adminAPI.updateResourceInstance(resource.uuid, { status: 'active' })
        fetchResources()
      } catch (error) {
        dialog.alertError('操作失败: ' + error.message)
      }
    }

    const closeModals = () => {
      showCreateModal.value = false
      showEditModal.value = false
      editingResource.value = null
      resetForm()
    }

    const openCreateModal = () => {
      resetForm()
      showCreateModal.value = true
    }

    const resetForm = () => {
      formData.value = {
        name: '',
        resource_type: 'virtual_machine',
        host: '',
        ssh_port: 22,
        ssh_user: 'root',
        passwd: '',
        snapshot_id: '',
        snapshot_instance_uuid: ''
      }
    }

    const changePage = (page) => {
      pagination.value.page = page
      fetchResources()
    }

    const debounceFetch = () => {
      clearTimeout(debounceTimer)
      debounceTimer = setTimeout(() => {
        pagination.value.page = 1
        fetchResources()
      }, 300)
    }

    const getStatusClass = (status) => {
      const classes = {
        'pending': 'status-pending',
        'active': 'status-active',
        'unreachable': 'status-unreachable'
      }
      return classes[status] || 'status-pending'
    }

    const getStatusLabel = (status) => {
      const labels = {
        'pending': '检查中',
        'active': '可用',
        'unreachable': '不可用'
      }
      return labels[status] || status
    }

    const getTypeLabel = (type) => {
      const labels = {
        'virtual_machine': '虚拟机',
        'physical_machine': '物理机',
        'VirtualMachine': '虚拟机',
        'Machine': '物理机'
      }
      return labels[type] || type
    }

    const getTypeClass = (type) => {
      const normalized = normalizeResourceType(type)
      return normalized === 'virtual_machine' ? 'type-vm' : 'type-pm'
    }

    const normalizeResourceType = (type) => {
      const mapping = {
        'VirtualMachine': 'virtual_machine',
        'Machine': 'physical_machine',
        'virtual_machine': 'virtual_machine',
        'physical_machine': 'physical_machine'
      }
      return mapping[type] || 'virtual_machine'
    }

    const formatTime = (time) => {
      if (!time) return '-'
      return new Date(time).toLocaleString('zh-CN')
    }

    const viewDetails = (resource) => {
      selectedResource.value = resource
      showDetailsModal.value = true
    }

    const closeDetailsModal = () => {
      showDetailsModal.value = false
      selectedResource.value = null
    }

    const viewTasks = (resource) => {
      selectedResourceForTasks.value = resource
      showTasksModal.value = true
      taskPagination.value.page = 1
      fetchTasks()
    }

    const closeTasksModal = () => {
      showTasksModal.value = false
      selectedResourceForTasks.value = null
      tasks.value = []
    }

    const fetchTasks = async () => {
      if (!selectedResourceForTasks.value) return
      tasksLoading.value = true
      try {
        const params = {
          page: taskPagination.value.page,
          page_size: taskPagination.value.page_size
        }
        if (taskFilters.value.status) params.status = taskFilters.value.status
        if (taskFilters.value.type) params.task_type = taskFilters.value.type

        const result = await adminAPI.getResourceInstanceTasks(selectedResourceForTasks.value.uuid, params)
        tasks.value = result.tasks || result.data || []
        taskPagination.value.total = result.total || 0
      } catch (error) {
        console.error('Failed to fetch tasks:', error)
        dialog.alertError('获取任务列表失败: ' + error.message)
      } finally {
        tasksLoading.value = false
      }
    }

    const changeTaskPage = (page) => {
      taskPagination.value.page = page
      fetchTasks()
    }

    const showTaskError = (task) => {
      dialog.alertError(`任务失败\n\n错误代码: ${task.error_code}\n错误信息: ${task.error_message}`)
    }

    const getTaskTypeClass = (type) => {
      const classes = {
        'deploy': 'type-deploy',
        'rollback': 'type-rollback',
        'health_check': 'type-health'
      }
      return classes[type] || 'type-default'
    }

    const getTaskStatusClass = (status) => {
      const classes = {
        'pending': 'status-pending',
        'running': 'status-running',
        'completed': 'status-completed',
        'failed': 'status-failed',
        'cancelled': 'status-cancelled'
      }
      return classes[status] || 'status-pending'
    }

    const canEditResource = (resource) => {
      return resource.created_by === currentUsername.value
    }

    const testConnection = async (resource) => {
      testing.value = resource.uuid
      try {
        const result = await adminAPI.checkResourceHealth(resource.uuid)
        await fetchResources()
        if (result.success || result.healthy) {
          dialog.alertSuccess(`连接测试成功！\n\n状态: ${result.status || 'healthy'}\nIP: ${result.ip_address}\n端口: ${result.port}\n\n资源实例状态已更新为: 可用`)
        } else {
          dialog.alertError(`连接测试失败！\n\n${result.message || '无法连接到资源实例'}\n\n资源实例状态已更新为: unreachable`)
        }
      } catch (error) {
        console.error('Health check failed:', error)
        dialog.alertError(`连接测试失败！\n\n${error.message || '无法连接到资源实例'}`)
      } finally {
        testing.value = null
      }
    }

    // 判断是否为虚拟机类型
    const isVirtualMachine = (resource) => {
      const type = resource.instance_type || resource.resource_type
      return type === 'virtual_machine' || type === 'VirtualMachine'
    }

    // 快照回滚确认
    const confirmRestoreSnapshot = (resource) => {
      restoringResource.value = resource
      showRestoreModal.value = true
    }

    // 执行快照回滚
    const executeRestoreSnapshot = async () => {
      if (!restoringResource.value) return

      restoring.value = restoringResource.value.uuid
      try {
        await adminAPI.restoreSnapshot(restoringResource.value.uuid)
        dialog.alertSuccess('快照回滚已启动，请等待几分钟让回滚操作完成')
        showRestoreModal.value = false
        restoringResource.value = null
      } catch (error) {
        console.error('Snapshot restore failed:', error)
        dialog.alertError(`快照回滚失败: ${error.message}`)
      } finally {
        restoring.value = null
      }
    }

    onMounted(fetchResources)

    return {
      resources, loading, submitting, testing, restoring, showCreateModal, showEditModal,
      showDetailsModal, showTasksModal, showRestoreModal, filters, pagination, formData, selectedResource,
      selectedResourceForTasks, restoringResource, tasks, tasksLoading, taskFilters, taskPagination,
      isAdmin, totalPages, taskTotalPages, saveResource, editResource, deleteResource,
      setMaintenance, setAvailable, testConnection, testNewConnection, closeModals,
      viewDetails, closeDetailsModal, viewTasks, closeTasksModal, fetchTasks,
      changeTaskPage, changePage, debounceFetch, getStatusClass, getStatusLabel,
      getTypeLabel, getTypeClass, formatTime, showTaskError, getTaskTypeClass,
      getTaskStatusClass, canEditResource, openCreateModal,
      isVirtualMachine, confirmRestoreSnapshot, executeRestoreSnapshot
    }
  }
}
</script>

<style scoped>
.resource-instance-page {
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
  background: linear-gradient(135deg, #6366F1 0%, #A5B4FC 100%);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.25);
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

.filters-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  padding: 1rem;
  margin-bottom: 1.5rem;
}

.filters {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  align-items: flex-end;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.filter-group label {
  font-size: 0.75rem;
  font-weight: 500;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.filter-spacer {
  flex: 1;
}

.filter-select {
  padding: 0.5rem 2rem 0.5rem 0.75rem;
  font-size: 0.875rem;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  background: white;
  color: #1E293B;
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='%2364748B'%3E%3Cpath stroke-linecap='round' stroke-linejoin='round' stroke-width='2' d='M19 9l-7 7-7-7'%3E%3C/path%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.5rem center;
  background-size: 1rem;
  min-width: 120px;
}

.filter-select:focus {
  outline: none;
  border-color: #7C3AED;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
}

.search-group {
  flex: 0 1 280px;
}

.search-input-wrapper {
  position: relative;
}

.search-icon {
  position: absolute;
  left: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  width: 16px;
  height: 16px;
  color: #94A3B8;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 0.5rem 0.75rem 0.5rem 2.25rem;
  font-size: 0.875rem;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  background: white;
  color: #1E293B;
  transition: all 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: #7C3AED;
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.1);
}

.search-input::placeholder {
  color: #94A3B8;
}

.form-control {
  padding: 0.5rem 0.75rem;
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

.table-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05), 0 4px 12px rgba(0, 0, 0, 0.05);
  border: 1px solid #E2E8F0;
  overflow: hidden;
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
  margin: 0;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.data-table th {
  background: linear-gradient(to bottom, #F8FAFC, #F1F5F9);
  padding: 0.75rem 1rem;
  text-align: left;
  font-size: 0.6875rem;
  font-weight: 600;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  border-bottom: 1px solid #E2E8F0;
  white-space: nowrap;
}

.data-table th.actions-col {
  text-align: center;
}

.data-table td {
  padding: 0.875rem 1rem;
  border-bottom: 1px solid #F1F5F9;
  font-size: 0.8125rem;
  color: #1E293B;
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

.host-cell {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.host-icon {
  width: 16px;
  height: 16px;
  color: #7C3AED;
  flex-shrink: 0;
}

.host-address {
  font-family: 'Monaco', 'Menlo', monospace;
  font-weight: 500;
  color: #1E293B;
}

.ssh-config-cell {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.8125rem;
}

.ssh-user {
  color: #1E293B;
  font-weight: 500;
}

.ssh-separator {
  color: #94A3B8;
}

.ssh-port {
  font-family: 'Monaco', 'Menlo', monospace;
  color: #7C3AED;
  background: #F5F3FF;
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  font-size: 0.75rem;
}

.empty-value {
  color: #CBD5E1;
}

.type-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 6px;
}

.type-vm {
  background: linear-gradient(135deg, #EFF6FF 0%, #DBEAFE 100%);
  color: #2563EB;
}

.type-pm {
  background: linear-gradient(135deg, #ECFDF5 0%, #D1FAE5 100%);
  color: #059669;
}

.snapshot-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.5rem;
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.75rem;
  color: #92400E;
  font-weight: 500;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 6px;
}

.status-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
}

.status-pending {
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
  color: #92400E;
}

.status-pending .status-dot {
  background: #F59E0B;
}

.status-active {
  background: linear-gradient(135deg, #ECFDF5 0%, #D1FAE5 100%);
  color: #059669;
}

.status-active .status-dot {
  background: #10B981;
}

.status-unreachable {
  background: linear-gradient(135deg, #FEF2F2 0%, #FEE2E2 100%);
  color: #DC2626;
}

.status-unreachable .status-dot {
  background: #EF4444;
}

.status-running {
  background: #EFF6FF;
  color: #2563EB;
}

.status-running .status-dot {
  background: #3B82F6;
}

.status-completed {
  background: #ECFDF5;
  color: #059669;
}

.status-completed .status-dot {
  background: #10B981;
}

.status-failed {
  background: #FEF2F2;
  color: #DC2626;
}

.status-failed .status-dot {
  background: #EF4444;
}

.status-cancelled {
  background: #F1F5F9;
  color: #64748B;
}

.status-cancelled .status-dot {
  background: #94A3B8;
}

.permission-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 6px;
}

.permission-badge.public {
  background: linear-gradient(135deg, #ECFDF5 0%, #D1FAE5 100%);
  color: #059669;
}

.permission-badge.private {
  background: linear-gradient(135deg, #F1F5F9 0%, #E2E8F0 100%);
  color: #64748B;
}

.action-buttons {
  display: flex;
  flex-direction: row;
  flex-wrap: nowrap;
  gap: 0.25rem;
  align-items: center;
  justify-content: center;
}

.action-btn {
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.action-btn svg {
  width: 14px;
  height: 14px;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn-test {
  background: #F5F3FF;
  color: #7C3AED;
}

.action-btn-test:hover:not(:disabled) {
  background: #EDE9FE;
  transform: translateY(-1px);
}

.action-btn-info {
  background: #EFF6FF;
  color: #2563EB;
}

.action-btn-info:hover:not(:disabled) {
  background: #DBEAFE;
  transform: translateY(-1px);
}

.action-btn-tasks {
  background: #ECFDF5;
  color: #059669;
}

.action-btn-tasks:hover:not(:disabled) {
  background: #D1FAE5;
  transform: translateY(-1px);
}

.action-btn-edit {
  background: #F5F3FF;
  color: #7C3AED;
}

.action-btn-edit:hover:not(:disabled) {
  background: #EDE9FE;
  transform: translateY(-1px);
}

.action-btn-warning {
  background: #FEF3C7;
  color: #D97706;
}

.action-btn-warning:hover:not(:disabled) {
  background: #FDE68A;
  transform: translateY(-1px);
}

.action-btn-success {
  background: #ECFDF5;
  color: #059669;
}

.action-btn-success:hover:not(:disabled) {
  background: #D1FAE5;
  transform: translateY(-1px);
}

.action-btn-danger {
  background: #FEF2F2;
  color: #DC2626;
}

.action-btn-danger:hover:not(:disabled) {
  background: #FEE2E2;
  transform: translateY(-1px);
}

.action-btn-rollback {
  background: #FEF3C7;
  color: #D97706;
}

.action-btn-rollback:hover:not(:disabled) {
  background: #FDE68A;
  transform: translateY(-1px);
}

.action-btn-rollback:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spin-icon {
  width: 16px;
  height: 16px;
}

.spin-path {
  animation: spin 1s linear infinite;
  transform-origin: center;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  border-top: 1px solid #E2E8F0;
}

.pagination-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  border: 1px solid #E2E8F0;
  border-radius: 8px;
  background: white;
  color: #64748B;
  cursor: pointer;
  transition: all 0.2s ease;
}

.pagination-btn svg {
  width: 16px;
  height: 16px;
}

.pagination-btn:hover:not(:disabled) {
  background: #F8FAFC;
  border-color: #CBD5E1;
  color: #1E293B;
}

.pagination-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  font-size: 0.875rem;
  color: #64748B;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(15, 23, 42, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 1rem;
}

.modal {
  background: white;
  border-radius: 16px;
  width: 100%;
  max-width: 560px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}

.modal-large {
  max-width: 900px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid #E2E8F0;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #0F172A;
  margin: 0;
}

.modal-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: #64748B;
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.modal-close svg {
  width: 20px;
  height: 20px;
}

.modal-close:hover {
  background: #F1F5F9;
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
}

/* 详情页面样式 */
.detail-modal {
  max-width: 600px;
}

.detail-header {
  padding-bottom: 1rem;
  border-bottom: 1px solid #E2E8F0;
}

.header-icon-section {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.detail-type-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.detail-type-icon svg {
  width: 24px;
  height: 24px;
}

.detail-type-icon.type-vm {
  background: linear-gradient(135deg, #EFF6FF 0%, #DBEAFE 100%);
  color: #2563EB;
}

.detail-type-icon.type-pm {
  background: linear-gradient(135deg, #ECFDF5 0%, #D1FAE5 100%);
  color: #059669;
}

.header-text {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-subtitle {
  font-size: 0.875rem;
  color: #64748B;
  font-family: 'Monaco', 'Menlo', monospace;
}

.detail-body {
  padding: 1.5rem;
}

.detail-status-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem;
  background: linear-gradient(135deg, #F8FAFC 0%, #F1F5F9 100%);
  border-radius: 12px;
  margin-bottom: 1.5rem;
  border: 1px solid #E2E8F0;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-indicator .status-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.status-indicator .status-dot.status-pending {
  background: #F59E0B;
}

.status-indicator .status-dot.status-active {
  background: #10B981;
}

.status-indicator .status-dot.status-unreachable {
  background: #EF4444;
}

.status-text {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1E293B;
}

.type-badge-large {
  padding: 0.5rem 1rem;
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 600;
}

.type-badge-large.type-vm {
  background: #EFF6FF;
  color: #2563EB;
}

.type-badge-large.type-pm {
  background: #ECFDF5;
  color: #059669;
}

.detail-sections {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.detail-section {
  background: white;
  border-radius: 12px;
  border: 1px solid #E2E8F0;
  overflow: hidden;
}

.detail-section-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.875rem 1rem;
  background: #F8FAFC;
  border-bottom: 1px solid #E2E8F0;
}

.section-icon {
  width: 18px;
  height: 18px;
  color: #7C3AED;
}

.section-title {
  font-size: 0.8125rem;
  font-weight: 600;
  color: #64748B;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0;
}

.detail-list {
  display: flex;
  flex-direction: column;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.875rem 1rem;
  border-bottom: 1px solid #F1F5F9;
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-key {
  font-size: 0.875rem;
  color: #64748B;
  font-weight: 500;
}

.detail-row .detail-value {
  font-size: 0.875rem;
  color: #1E293B;
  font-weight: 500;
  text-align: right;
}

.detail-row .detail-value.mono {
  font-family: 'Monaco', 'Menlo', monospace;
  background: #F1F5F9;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}

.detail-row .detail-value.masked {
  letter-spacing: 4px;
  color: #94A3B8;
}

.detail-row .detail-value.snapshot {
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
  color: #92400E;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #E2E8F0;
  background: #F8FAFC;
}

.form-section {
  margin-bottom: 1.5rem;
}

.form-section:last-child {
  margin-bottom: 0;
}

.form-section h4 {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #E2E8F0;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-group label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
}

.required {
  color: #DC2626;
}

.form-group small {
  font-size: 0.75rem;
  color: #64748B;
}

.resource-info-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: #F8FAFC;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.resource-label {
  font-size: 0.875rem;
  color: #64748B;
}

.resource-value {
  font-size: 0.875rem;
  font-weight: 600;
  color: #1E293B;
  font-family: 'Monaco', 'Menlo', monospace;
}

.task-filters {
  display: flex;
  gap: 1rem;
  margin-bottom: 1rem;
}

.task-type-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: 6px;
}

.type-deploy {
  background: #F5F3FF;
  color: #7C3AED;
}

.type-rollback {
  background: #FEF3C7;
  color: #D97706;
}

.type-health {
  background: #EFF6FF;
  color: #2563EB;
}

.type-default {
  background: #F1F5F9;
  color: #64748B;
}

.result-success {
  color: #059669;
  font-weight: 500;
}

.result-failed {
  color: #DC2626;
  font-weight: 500;
}

.text-muted {
  color: #94A3B8;
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
  /* Ensure consistent height across all buttons */
  min-height: 40px;
  height: auto;
  line-height: 1;
  box-sizing: border-box;
}

.btn-primary {
  background: linear-gradient(135deg, #7C3AED 0%, #6366F1 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(124, 58, 237, 0.3);
}

.btn-secondary {
  background: white;
  color: #64748B;
  border: 1px solid #E2E8F0;
}

.btn-secondary:hover:not(:disabled) {
  background: #F8FAFC;
  border-color: #CBD5E1;
}

.btn-outline {
  background: white;
  color: #7C3AED;
  border: 1px solid #7C3AED;
}

.btn-outline:hover:not(:disabled) {
  background: #F5F3FF;
}

.btn-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
  display: inline-block;
  vertical-align: middle;
}

.btn-warning {
  background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%);
  color: white;
}

.btn-warning:hover:not(:disabled) {
  background: linear-gradient(135deg, #D97706 0%, #B45309 100%);
}

.btn-spinner {
  width: 16px;
  height: 16px;
  animation: spin 1s linear infinite;
}

.spinner-circle {
  animation: dash 1.5s ease-in-out infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes dash {
  from { stroke-dashoffset: 32; }
  to { stroke-dashoffset: 0; }
}

/* 快照回滚对话框样式 */
.modal-icon-warning {
  color: #F59E0B;
}

.restore-info {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.restore-details {
  background: #F8FAFC;
  border-radius: 8px;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.restore-detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.restore-detail-label {
  font-weight: 500;
  color: #64748B;
}

.restore-detail-value {
  font-weight: 600;
  color: #1E293B;
}

.snapshot-id {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  background: #FEF3C7;
  padding: 0.125rem 0.375rem;
  border-radius: 4px;
  color: #92400E;
}

.restore-warning {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: #FEF3C7;
  border-left: 4px solid #F59E0B;
  border-radius: 0 8px 8px 0;
  font-size: 0.875rem;
  color: #92400E;
}

.restore-warning svg {
  width: 16px;
  height: 16px;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .filters {
    flex-direction: column;
  }

  .filter-group {
    width: 100%;
  }

  .form-row {
    grid-template-columns: 1fr;
  }

  .detail-row {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .detail-row .detail-value {
    text-align: left;
  }

  .data-table {
    display: block;
    overflow-x: auto;
  }
}
</style>
