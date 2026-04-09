<template>
  <div class="testbed-detail-page">
    <div v-if="loading" class="loading">
      <div class="spinner"></div>
      <span>加载中...</span>
    </div>

    <div v-else-if="!testbed" class="empty-state">
      <div class="empty-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </div>
      <h3>Testbed 不存在</h3>
      <p>无法找到指定的 Testbed 信息</p>
      <button class="btn btn-primary" @click="goBack">返回列表</button>
    </div>

    <template v-else>
      <div class="page-header">
        <div class="header-icon">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
          </svg>
        </div>
        <div class="header-content">
          <h1>{{ testbed.name }}</h1>
          <p>Testbed 详细信息与配置</p>
        </div>
        <div class="header-status">
          <span class="status-badge" :class="getStatusClass(testbed.status)">
            <span class="status-dot"></span>
            {{ getStatusLabel(testbed.status) }}
          </span>
        </div>
        <button class="btn btn-secondary" @click="goBack">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
          返回
        </button>
      </div>

      <div class="content-grid">
        <div class="main-content">
          <div class="card">
            <div class="card-header">
              <div class="header-left">
                <div class="header-icon-small">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 class="card-title">基本信息</h3>
              </div>
            </div>
            <div class="info-grid">
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                  </svg>
                  UUID
                </div>
                <div class="info-value uuid">{{ testbed.uuid }}</div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                  </svg>
                  名称
                </div>
                <div class="info-value">{{ testbed.name }}</div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                  </svg>
                  类别
                </div>
                <div class="info-value">{{ testbed.category_name || '-' }}</div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
                  </svg>
                  服务对象
                </div>
                <div class="info-value">
                  <span class="service-badge" :class="getServiceClass(testbed.service_target)">
                    {{ formatServiceTarget(testbed.service_target) }}
                  </span>
                </div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  状态
                </div>
                <div class="info-value">
                  <span class="status-badge" :class="getStatusClass(testbed.status)">
                    <span class="status-dot"></span>
                    {{ getStatusLabel(testbed.status) }}
                  </span>
                </div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  创建时间
                </div>
                <div class="info-value">{{ formatTime(testbed.created_at) }}</div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                  </svg>
                  最后健康检查
                </div>
                <div class="info-value">{{ formatTime(testbed.last_health_check_at) }}</div>
              </div>
            </div>
          </div>

          <div class="card">
            <div class="card-header">
              <div class="header-left">
                <div class="header-icon-small">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                </div>
                <h3 class="card-title">连接信息</h3>
              </div>
            </div>
            <div class="info-grid">
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                  </svg>
                  主机
                </div>
                <div class="info-value host">{{ testbed.host }}</div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                  SSH 端口
                </div>
                <div class="info-value port">{{ testbed.ssh_port }}</div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  SSH 登录名
                </div>
                <div class="info-value">{{ testbed.ssh_user || '-' }}</div>
              </div>
              <div class="info-item full-width">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                  SSH 密码
                </div>
                <div class="info-value password-field">
                  <span class="password-value">
                    {{ showSSHPassword ? (testbed.ssh_passwd || '****') : '****' }}
                  </span>
                  <button class="btn-toggle" @click="showSSHPassword = !showSSHPassword">
                    <svg v-if="!showSSHPassword" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                    {{ showSSHPassword ? '隐藏' : '显示' }}
                  </button>
                </div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
                  </svg>
                  数据库端口
                </div>
                <div class="info-value port">{{ testbed.db_port }}</div>
              </div>
              <div class="info-item">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  数据库用户
                </div>
                <div class="info-value">{{ testbed.db_user }}</div>
              </div>
              <div class="info-item full-width">
                <div class="info-label">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                  数据库密码
                </div>
                <div class="info-value password-field">
                  <span class="password-value">
                    {{ showPassword ? (testbed.db_password || '****') : '****' }}
                  </span>
                  <button class="btn-toggle" @click="showPassword = !showPassword">
                    <svg v-if="!showPassword" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                    </svg>
                    {{ showPassword ? '隐藏' : '显示' }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <div v-if="testbed.resource_instance" class="card">
            <div class="card-header">
              <div class="header-left">
                <div class="header-icon-small">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                  </svg>
                </div>
                <h3 class="card-title">关联资源实例</h3>
              </div>
            </div>
            <div class="info-grid">
              <div class="info-item">
                <div class="info-label">名称</div>
                <div class="info-value">{{ testbed.resource_instance.name }}</div>
              </div>
              <div class="info-item">
                <div class="info-label">类型</div>
                <div class="info-value">
                  <span class="type-badge">{{ testbed.resource_instance.resource_type }}</span>
                </div>
              </div>
              <div class="info-item">
                <div class="info-label">快照 ID</div>
                <div class="info-value">{{ testbed.resource_instance.snapshot_id || '-' }}</div>
              </div>
              <div class="info-item">
                <div class="info-label">状态</div>
                <div class="info-value">
                  <span class="status-badge" :class="getStatusClass(testbed.resource_instance.status)">
                    <span class="status-dot"></span>
                    {{ getStatusLabel(testbed.resource_instance.status) }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div class="card">
            <div class="card-header">
              <div class="header-left">
                <div class="header-icon-small">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 class="card-title">操作历史</h3>
              </div>
            </div>
            <div v-if="!history || history.length === 0" class="empty-state-small">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p>暂无操作历史</p>
            </div>
            <table v-else class="history-table">
              <thead>
                <tr>
                  <th>时间</th>
                  <th>操作</th>
                  <th>操作人</th>
                  <th>详情</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in history" :key="item.id">
                  <td :title="formatTime(item.created_at)">
                    <span class="time-value">{{ formatTime(item.created_at) }}</span>
                  </td>
                  <td :title="item.action">
                    <span class="action-badge">{{ item.action }}</span>
                  </td>
                  <td :title="item.operator || '-'">{{ item.operator || '-' }}</td>
                  <td :title="item.details || '-'">{{ item.details || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <div class="side-content">
          <div v-if="currentAllocation" class="card allocation-card">
            <div class="card-header">
              <div class="header-left">
                <div class="header-icon-small header-icon-warning">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h3 class="card-title">当前分配</h3>
              </div>
            </div>
            <div class="allocation-content">
              <div class="allocation-user">
                <div class="user-avatar">{{ getInitial(currentAllocation.allocated_to) }}</div>
                <div class="user-info">
                  <span class="user-name">{{ currentAllocation.allocated_to }}</span>
                  <span class="user-label">当前使用者</span>
                </div>
              </div>
              <div class="allocation-details">
                <div class="detail-item">
                  <span class="detail-label">分配 UUID</span>
                  <span class="detail-value uuid">{{ currentAllocation.uuid }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">申请时间</span>
                  <span class="detail-value">{{ formatTime(currentAllocation.created_at) }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">过期时间</span>
                  <span class="detail-value expiry">{{ formatTime(currentAllocation.expires_at) }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">用途</span>
                  <span class="detail-value">{{ currentAllocation.purpose || '-' }}</span>
                </div>
              </div>
              <div v-if="isOwner(currentAllocation)" class="allocation-actions">
                <button class="btn btn-warning" @click="extendAllocation">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  延期
                </button>
                <button class="btn btn-danger" @click="releaseAllocation">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                  释放
                </button>
              </div>
            </div>
          </div>

          <div v-else class="card available-card">
            <div class="available-content">
              <div class="available-icon">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h4>当前可用</h4>
              <p>此 Testbed 当前未被分配</p>
            </div>
          </div>
        </div>
      </div>
    </template>

    <div v-if="showExtendModal" class="modal-overlay" @click.self="showExtendModal = false">
      <div class="modal modal-small">
        <div class="modal-header">
          <div class="modal-title-row">
            <div class="modal-icon modal-icon-warning">
              <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <h3 class="modal-title">延期 Testbed</h3>
          </div>
          <button class="modal-close" @click="showExtendModal = false">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="extend-info">
            <div class="info-row">
              <span class="info-label">当前过期时间</span>
              <span class="info-value">{{ formatTime(currentAllocation?.expires_at) }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">可延期时长</span>
              <span class="info-value highlight">{{ getMaxExtendTime() }} 分钟</span>
            </div>
          </div>
          <div class="form-group">
            <label for="extendTime">延期时长 (分钟)</label>
            <input
              id="extendTime"
              v-model.number="extendMinutes"
              type="number"
              min="1"
              :max="getMaxExtendTime()"
              class="form-control"
              required
            />
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-secondary" @click="showExtendModal = false">
            取消
          </button>
          <button type="button" class="btn btn-primary" :disabled="extending" @click="confirmExtend">
            {{ extending ? '延期中...' : '确认延期' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { externalAPI } from '../../api/resourcePool'
import { useDialog } from '../../composables/useDialog'

export default {
  name: 'TestbedDetail',
  setup() {
    const dialog = useDialog()
    const route = useRoute()
    const testbed = ref(null)
    const currentAllocation = ref(null)
    const history = ref([])
    const loading = ref(true)
    const extending = ref(false)
    const showExtendModal = ref(false)
    const showPassword = ref(false)
    const showSSHPassword = ref(false)
    const extendMinutes = ref(30)

    const fetchTestbed = async () => {
      loading.value = true
      try {
        const uuid = route.params.uuid
        const data = await externalAPI.getTestbed(uuid)

        testbed.value = data.data || data.testbed || data

        if (testbed.value?.current_allocation) {
          currentAllocation.value = testbed.value.current_allocation
        } else {
          try {
            const allocData = await externalAPI.getMyAllocations({ testbed_uuid: uuid, status: 'active' })
            const allocations = allocData.data || allocData.allocations || []
            if (allocations.length > 0) {
              currentAllocation.value = allocations[0]
            }
          } catch (e) {
            // Ignore error
          }
        }
      } catch (error) {
        console.error('Failed to fetch testbed:', error)
        dialog.alertError('获取 Testbed 详情失败: ' + error.message)
      } finally {
        loading.value = false
      }
    }

    const extendAllocation = async () => {
      showExtendModal.value = true
    }

    const confirmExtend = async () => {
      extending.value = true
      try {
        await externalAPI.extendAllocation(currentAllocation.value.uuid, extendMinutes.value * 60)
        dialog.alertSuccess('延期成功')
        showExtendModal.value = false
        fetchTestbed()
      } catch (error) {
        dialog.alertError('延期失败: ' + error.message)
      } finally {
        extending.value = false
      }
    }

    const releaseAllocation = async () => {
      let testbedName = '此 Testbed'
      if (testbed.value) {
        testbedName = testbed.value.name || testbed.value.testbed_name || testbed.value.uuid?.substring(0, 8) || '此 Testbed'
      }
      const confirmed = await dialog.confirm(`确定要释放 "${testbedName}" 吗？`)
      if (!confirmed) return
      try {
        await externalAPI.releaseTestbed(currentAllocation.value.uuid)
        dialog.alertSuccess('释放成功')
        currentAllocation.value = null
        fetchTestbed()
      } catch (error) {
        dialog.alertError('释放失败: ' + error.message)
      }
    }

    const isOwner = (allocation) => {
      const username = localStorage.getItem('username')
      return allocation.allocated_to === username
    }

    const getMaxExtendTime = () => {
      if (currentAllocation.value?.can_extend_until) {
        const until = new Date(currentAllocation.value.can_extend_until)
        const now = new Date()
        const diff = Math.floor((until - now) / 1000 / 60)
        return Math.max(0, diff)
      }
      return 30
    }

    const getStatusClass = (status) => {
      const classes = {
        'available': 'status-success',
        'allocated': 'status-info',
        'in_use': 'status-warning',
        'releasing': 'status-secondary',
        'deleted': 'status-danger'
      }
      return classes[status] || 'status-secondary'
    }

    const getStatusLabel = (status) => {
      const labels = {
        'available': '可用',
        'allocated': '已分配',
        'in_use': '使用中',
        'releasing': '释放中',
        'deleted': '已删除'
      }
      return labels[status] || status
    }

    const getServiceClass = (serviceTarget) => {
      const classes = {
        'robot': 'service-robot',
        'normal': 'service-normal'
      }
      return classes[serviceTarget] || ''
    }

    const formatTime = (time) => {
      if (!time) return '-'
      return new Date(time).toLocaleString('zh-CN')
    }

    const formatServiceTarget = (serviceTarget) => {
      const labels = {
        'robot': 'Robot',
        'normal': '普通用户'
      }
      return labels[serviceTarget] || '-'
    }

    const getInitial = (username) => {
      if (!username) return '?'
      return username.charAt(0).toUpperCase()
    }

    const goBack = () => {
      window.history.back()
    }

    onMounted(fetchTestbed)

    return {
      testbed,
      currentAllocation,
      history,
      loading,
      extending,
      showExtendModal,
      showPassword,
      showSSHPassword,
      extendMinutes,
      extendAllocation,
      confirmExtend,
      releaseAllocation,
      isOwner,
      getMaxExtendTime,
      getStatusClass,
      getStatusLabel,
      getServiceClass,
      formatTime,
      formatServiceTarget,
      getInitial,
      goBack
    }
  }
}
</script>

<style scoped>
.testbed-detail-page {
  max-width: 1200px;
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

.header-status {
  margin-left: auto;
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
  background: var(--bg-card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  border: 1px solid var(--border);
}

.empty-icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: rgba(239, 68, 68, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.empty-icon svg {
  width: 40px;
  height: 40px;
  color: #EF4444;
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

.content-grid {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 1.5rem;
}

.main-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.side-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.card {
  background: var(--bg-card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  border: 1px solid var(--border);
  overflow: hidden;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.header-icon-small {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.1), rgba(124, 58, 237, 0.1));
  display: flex;
  align-items: center;
  justify-content: center;
  color: #8B5CF6;
}

.header-icon-small svg {
  width: 18px;
  height: 18px;
}

.header-icon-warning {
  background: rgba(245, 158, 11, 0.1);
  color: #F59E0B;
}

.card-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  padding: 1.5rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-item.full-width {
  grid-column: 1 / -1;
}

.info-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.info-label svg {
  width: 14px;
  height: 14px;
}

.info-value {
  font-size: 0.9375rem;
  color: var(--text-primary);
  font-weight: 500;
}

.info-value.uuid {
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.info-value.host {
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  color: #8B5CF6;
}

.info-value.port {
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  background: rgba(139, 92, 246, 0.1);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  display: inline-block;
}

.password-field {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: var(--bg-main);
  padding: 0.625rem 0.875rem;
  border-radius: var(--radius-sm);
}

.password-value {
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  flex: 1;
}

.btn-toggle {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.375rem 0.625rem;
  font-size: 0.75rem;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-secondary);
}

.btn-toggle svg {
  width: 14px;
  height: 14px;
}

.btn-toggle:hover {
  background: var(--bg-secondary);
  border-color: #8B5CF6;
  color: #8B5CF6;
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

.status-success {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
}

.status-success .status-dot {
  background: #10B981;
}

.status-info {
  background: rgba(59, 130, 246, 0.1);
  color: #3B82F6;
}

.status-info .status-dot {
  background: #3B82F6;
}

.status-warning {
  background: rgba(245, 158, 11, 0.1);
  color: #F59E0B;
}

.status-warning .status-dot {
  background: #F59E0B;
}

.status-secondary {
  background: rgba(107, 114, 128, 0.1);
  color: #6B7280;
}

.status-secondary .status-dot {
  background: #6B7280;
}

.status-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #EF4444;
}

.status-danger .status-dot {
  background: #EF4444;
}

.service-badge {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
}

.service-robot {
  background: rgba(139, 92, 246, 0.1);
  color: #8B5CF6;
}

.service-normal {
  background: rgba(59, 130, 246, 0.1);
  color: #3B82F6;
}

.type-badge {
  display: inline-block;
  padding: 0.25rem 0.625rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  background: rgba(139, 92, 246, 0.1);
  color: #8B5CF6;
}

.empty-state-small {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.empty-state-small svg {
  width: 40px;
  height: 40px;
  margin-bottom: 0.5rem;
  opacity: 0.5;
}

.empty-state-small p {
  margin: 0;
  font-size: 0.875rem;
}

.history-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.history-table th,
.history-table td {
  padding: 0.875rem 1.25rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

.history-table td {
  /* Apply text overflow to all table cells */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Ensure inline elements in table cells also truncate */
.history-table td > *,
.history-table td > span,
.history-table td > a {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

.history-table th {
  background: var(--bg-main);
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
}

.history-table tbody tr:last-child td {
  border-bottom: none;
}

.time-value {
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.action-badge {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  background: rgba(139, 92, 246, 0.1);
  color: #8B5CF6;
}

.allocation-card {
  position: sticky;
  top: 1rem;
}

.allocation-content {
  padding: 1.5rem;
}

.allocation-user {
  display: flex;
  align-items: center;
  gap: 0.875rem;
  margin-bottom: 1.25rem;
  padding-bottom: 1.25rem;
  border-bottom: 1px solid var(--border);
}

.user-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #8B5CF6, #7C3AED);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 1.125rem;
  font-weight: 600;
}

.user-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.user-name {
  font-weight: 600;
  color: var(--text-primary);
}

.user-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.allocation-details {
  display: flex;
  flex-direction: column;
  gap: 0.875rem;
  margin-bottom: 1.25rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.detail-value {
  font-size: 0.875rem;
  color: var(--text-primary);
  font-weight: 500;
}

.detail-value.uuid {
  font-family: 'SF Mono', 'Monaco', 'Consolas', monospace;
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 400;
}

.detail-value.expiry {
  color: #F59E0B;
}

.allocation-actions {
  display: flex;
  gap: 0.75rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
}

.available-card {
  position: sticky;
  top: 1rem;
}

.available-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  text-align: center;
}

.available-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: rgba(16, 185, 129, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.available-icon svg {
  width: 32px;
  height: 32px;
  color: #10B981;
}

.available-content h4 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 0.25rem 0;
}

.available-content p {
  font-size: 0.875rem;
  color: var(--text-secondary);
  margin: 0;
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

.btn-secondary {
  background: var(--bg-main);
  color: var(--text-primary);
  border: 1px solid var(--border);
}

.btn-secondary:hover {
  background: var(--bg-secondary);
  border-color: var(--text-secondary);
}

.btn-warning {
  background: linear-gradient(135deg, #F59E0B, #D97706);
  color: white;
  flex: 1;
}

.btn-warning:hover {
  background: linear-gradient(135deg, #D97706, #B45309);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(245, 158, 11, 0.3);
}

.btn-danger {
  background: linear-gradient(135deg, #EF4444, #DC2626);
  color: white;
  flex: 1;
}

.btn-danger:hover {
  background: linear-gradient(135deg, #DC2626, #B91C1C);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
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
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-small {
  max-width: 420px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border);
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

.modal-icon-warning {
  background: rgba(245, 158, 11, 0.1);
  color: #F59E0B;
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
}

.extend-info {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
  padding: 1rem;
  background: var(--bg-main);
  border-radius: var(--radius-sm);
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-row .info-label {
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.info-row .info-value {
  font-size: 0.875rem;
  color: var(--text-primary);
  font-weight: 500;
}

.info-row .info-value.highlight {
  color: #F59E0B;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: var(--text-primary);
  font-size: 0.875rem;
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

@media (max-width: 1024px) {
  .content-grid {
    grid-template-columns: 1fr;
  }

  .side-content {
    order: -1;
  }

  .allocation-card,
  .available-card {
    position: static;
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-wrap: wrap;
  }

  .header-status {
    margin-left: 0;
    order: 3;
    width: 100%;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .allocation-actions {
    flex-direction: column;
  }
}
</style>
