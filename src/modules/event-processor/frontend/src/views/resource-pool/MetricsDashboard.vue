<template>
  <div class="metrics-dashboard-page">
    <div class="page-header">
      <div class="header-icon">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
        </svg>
      </div>
      <div class="header-content">
        <h1>监控仪表盘</h1>
        <p>资源池使用情况与统计分析</p>
      </div>
      <div class="header-actions">
        <div class="time-range-selector">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" class="selector-icon">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <select v-model="timeRange" class="time-select" @change="fetchMetrics">
            <option value="24h">过去 24 小时</option>
            <option value="7d">过去 7 天</option>
            <option value="30d">过去 30 天</option>
          </select>
        </div>
      </div>
    </div>

    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon stat-icon-primary">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ metrics.total_testbeds || 0 }}</div>
          <div class="stat-label">总 Testbed 数</div>
        </div>
        <div class="stat-trend trend-up">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
          </svg>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon stat-icon-success">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ metrics.available_testbeds || 0 }}</div>
          <div class="stat-label">可用 Testbed</div>
        </div>
        <div class="stat-percentage">
          <span class="percentage-value">{{ availablePercentage }}%</span>
          <span class="percentage-label">可用率</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon stat-icon-warning">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ metrics.active_allocations || 0 }}</div>
          <div class="stat-label">使用中分配</div>
        </div>
        <div class="stat-indicator">
          <div class="pulse-dot"></div>
          <span>活跃</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon stat-icon-info">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ metrics.total_users || 0 }}</div>
          <div class="stat-label">活跃用户</div>
        </div>
        <div class="stat-avatars">
          <div class="avatar avatar-1"></div>
          <div class="avatar avatar-2"></div>
          <div class="avatar avatar-3"></div>
        </div>
      </div>
    </div>

    <div class="charts-row">
      <div class="card chart-card">
        <div class="card-header">
          <div class="header-left">
            <h3 class="card-title">资源利用率</h3>
            <span class="card-subtitle">实时监控</span>
          </div>
          <div class="utilization-badge" :class="getUtilizationClass()">
            {{ utilizationRate }}%
          </div>
        </div>
        <div class="chart-container">
          <div class="donut-chart">
            <svg viewBox="0 0 100 100">
              <defs>
                <linearGradient id="utilizationGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" :style="`stop-color: ${getUtilizationGradientStart()}`" />
                  <stop offset="100%" :style="`stop-color: ${getUtilizationGradientEnd()}`" />
                </linearGradient>
              </defs>
              <circle
                cx="50"
                cy="50"
                r="40"
                fill="none"
                stroke="#E2E8F0"
                stroke-width="12"
              />
              <circle
                cx="50"
                cy="50"
                r="40"
                fill="none"
                stroke="url(#utilizationGradient)"
                stroke-width="12"
                stroke-dasharray="251.2"
                :stroke-dashoffset="251.2 * (1 - utilizationRate / 100)"
                transform="rotate(-90 50 50)"
                class="utilization-circle"
              />
            </svg>
            <div class="donut-center">
              <span class="donut-value">{{ utilizationRate }}%</span>
              <span class="donut-label">利用率</span>
            </div>
          </div>
          <div class="utilization-legend">
            <div class="legend-item">
              <div class="legend-dot legend-used"></div>
              <span>已使用: {{ metrics.total_testbeds - metrics.available_testbeds || 0 }}</span>
            </div>
            <div class="legend-item">
              <div class="legend-dot legend-available"></div>
              <span>可用: {{ metrics.available_testbeds || 0 }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="card chart-card">
        <div class="card-header">
          <div class="header-left">
            <h3 class="card-title">类别分布</h3>
            <span class="card-subtitle">资源统计</span>
          </div>
        </div>
        <div class="category-list">
          <div
            v-for="(cat, index) in categoryDistribution"
            :key="cat.uuid"
            class="category-item"
          >
            <div class="category-header">
              <div class="category-icon" :style="{ background: getCategoryColor(index) }">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                </svg>
              </div>
              <div class="category-info">
                <span class="category-name">{{ cat.name }}</span>
                <span class="category-count">{{ cat.available }} / {{ cat.total }} 可用</span>
              </div>
            </div>
            <div class="progress-container">
              <div class="progress-bar">
                <div
                  class="progress-fill"
                  :style="{ 
                    width: (cat.total > 0 ? cat.available / cat.total * 100 : 0) + '%',
                    background: getCategoryColor(index)
                  }"
                ></div>
              </div>
              <span class="progress-text">{{ cat.total > 0 ? Math.round(cat.available / cat.total * 100) : 0 }}%</span>
            </div>
          </div>
          <div v-if="categoryDistribution.length === 0" class="empty-categories">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
            </svg>
            <p>暂无类别数据</p>
          </div>
        </div>
      </div>
    </div>

    <div class="content-row">
      <div class="card table-card">
        <div class="card-header">
          <div class="header-left">
            <h3 class="card-title">用户使用排行</h3>
            <span class="card-subtitle">TOP {{ userStats.length }}</span>
          </div>
        </div>
        <div v-if="loadingStats" class="loading">
          <div class="spinner"></div>
          <span>加载中...</span>
        </div>
        <div v-else-if="userStats.length === 0" class="empty-state">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
          </svg>
          <p>暂无用户数据</p>
        </div>
        <table v-else class="stats-table">
          <thead>
            <tr>
              <th>排名</th>
              <th>用户</th>
              <th>当前分配</th>
              <th>历史分配</th>
              <th>总使用时长</th>
              <th>最后使用</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(user, index) in userStats" :key="user.username">
              <td :title="index + 1">
                <span class="rank-badge" :class="getRankClass(index)">
                  <span v-if="index < 3" class="rank-icon">
                    <svg v-if="index === 0" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" />
                    </svg>
                    <svg v-else-if="index === 1" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2L15.09 8.26L22 9.27L17 14.14L18.18 21.02L12 17.77L5.82 21.02L7 14.14L2 9.27L8.91 8.26L12 2Z" />
                    </svg>
                  </span>
                  {{ index + 1 }}
                </span>
              </td>
              <td :title="user.username">
                <div class="user-cell">
                  <div class="user-avatar">{{ getInitial(user.username) }}</div>
                  <span class="user-name">{{ user.username }}</span>
                </div>
              </td>
              <td :title="user.current_allocations">
                <span class="badge" :class="user.current_allocations > 0 ? 'badge-warning' : 'badge-secondary'">
                  {{ user.current_allocations }}
                </span>
              </td>
              <td :title="user.total_allocations">
                <span class="number-value">{{ user.total_allocations }}</span>
              </td>
              <td :title="formatDuration(user.total_duration_seconds)">
                <span class="duration-value">{{ formatDuration(user.total_duration_seconds) }}</span>
              </td>
              <td :title="formatTime(user.last_used_at)">
                <span class="time-value">{{ formatTime(user.last_used_at) }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="card activity-card">
        <div class="card-header">
          <div class="header-left">
            <h3 class="card-title">最近活动</h3>
            <span class="card-subtitle">实时动态</span>
          </div>
        </div>
        <div v-if="recentActivity.length === 0" class="empty-state">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p>暂无活动记录</p>
        </div>
        <div v-else class="activity-list">
          <div
            v-for="activity in recentActivity"
            :key="activity.id"
            class="activity-item"
          >
            <div class="activity-icon" :class="getActivityIconClass(activity.action)">
              <svg v-if="activity.action === 'acquired'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <svg v-else-if="activity.action === 'released'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div class="activity-content">
              <div class="activity-title">{{ activity.title }}</div>
              <div class="activity-meta">
                <span class="activity-user">
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                  {{ activity.user }}
                </span>
                <span class="activity-time">{{ formatRelativeTime(activity.time) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { adminAPI } from '../../api/resourcePool'

export default {
  name: 'MetricsDashboard',
  setup() {
    const metrics = ref({})
    const categoryDistribution = ref([])
    const userStats = ref([])
    const recentActivity = ref([])
    const loadingStats = ref(false)
    const timeRange = ref('24h')

    const utilizationRate = computed(() => {
      if (!metrics.value.total_testbeds) return 0
      const used = (metrics.value.total_testbeds || 0) - (metrics.value.available_testbeds || 0)
      return Math.round((used / metrics.value.total_testbeds) * 100)
    })

    const availablePercentage = computed(() => {
      if (!metrics.value.total_testbeds) return 0
      return Math.round((metrics.value.available_testbeds / metrics.value.total_testbeds) * 100)
    })

    const fetchMetrics = async () => {
      try {
        const data = await adminAPI.getMetrics()
        metrics.value = data.data || data.metrics || data
      } catch (error) {
        console.error('Failed to fetch metrics:', error)
      }
    }

    const fetchUsageStats = async () => {
      loadingStats.value = true
      try {
        const params = {}
        if (timeRange.value) {
          params.time_range = timeRange.value
        }

        const data = await adminAPI.getUsageStats(params)

        categoryDistribution.value = data.data?.categories || data.categories || []
        userStats.value = data.data?.users || data.users || []
        recentActivity.value = data.data?.recent_activity || data.recent_activity || []
      } catch (error) {
        console.error('Failed to fetch usage stats:', error)
      } finally {
        loadingStats.value = false
      }
    }

    const getUtilizationClass = () => {
      const rate = utilizationRate.value
      if (rate > 80) return 'utilization-danger'
      if (rate > 60) return 'utilization-warning'
      return 'utilization-success'
    }

    const getUtilizationGradientStart = () => {
      const rate = utilizationRate.value
      if (rate > 80) return '#EF4444'
      if (rate > 60) return '#F59E0B'
      return '#10B981'
    }

    const getUtilizationGradientEnd = () => {
      const rate = utilizationRate.value
      if (rate > 80) return '#DC2626'
      if (rate > 60) return '#D97706'
      return '#059669'
    }

    const getCategoryColor = (index) => {
      const colors = [
        'linear-gradient(135deg, #8B5CF6, #7C3AED)',
        'linear-gradient(135deg, #3B82F6, #1D4ED8)',
        'linear-gradient(135deg, #10B981, #059669)',
        'linear-gradient(135deg, #F59E0B, #D97706)',
        'linear-gradient(135deg, #EF4444, #DC2626)',
        'linear-gradient(135deg, #EC4899, #DB2777)'
      ]
      return colors[index % colors.length]
    }

    const getRankClass = (index) => {
      if (index === 0) return 'rank-gold'
      if (index === 1) return 'rank-silver'
      if (index === 2) return 'rank-bronze'
      return ''
    }

    const getActivityIconClass = (action) => {
      const classes = {
        'acquired': 'icon-success',
        'released': 'icon-info',
        'extended': 'icon-warning'
      }
      return classes[action] || 'icon-secondary'
    }

    const getInitial = (username) => {
      if (!username) return '?'
      return username.charAt(0).toUpperCase()
    }

    const formatDuration = (seconds) => {
      if (!seconds) return '-'
      const hours = Math.floor(seconds / 3600)
      if (hours >= 1) {
        return `${hours} 小时`
      }
      return `${Math.floor(seconds / 60)} 分钟`
    }

    const formatTime = (time) => {
      if (!time) return '-'
      return new Date(time).toLocaleString('zh-CN')
    }

    const formatRelativeTime = (time) => {
      if (!time) return '-'
      const diff = Date.now() - new Date(time).getTime()
      const minutes = Math.floor(diff / 60000)
      const hours = Math.floor(diff / 3600000)
      const days = Math.floor(diff / 86400000)

      if (days > 0) return `${days} 天前`
      if (hours > 0) return `${hours} 小时前`
      if (minutes > 0) return `${minutes} 分钟前`
      return '刚刚'
    }

    onMounted(() => {
      fetchMetrics()
      fetchUsageStats()
    })

    return {
      metrics,
      categoryDistribution,
      userStats,
      recentActivity,
      loadingStats,
      timeRange,
      utilizationRate,
      availablePercentage,
      fetchMetrics,
      fetchUsageStats,
      getUtilizationClass,
      getUtilizationGradientStart,
      getUtilizationGradientEnd,
      getCategoryColor,
      getRankClass,
      getActivityIconClass,
      getInitial,
      formatDuration,
      formatTime,
      formatRelativeTime
    }
  }
}
</script>

<style scoped>
.metrics-dashboard-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 1.25rem;
  margin-bottom: 2rem;
  padding: 1.5rem;
  background: var(--bg-card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  border: 1px solid var(--border);
}

.header-icon {
  width: 56px;
  height: 56px;
  border-radius: 12px;
  background: linear-gradient(135deg, #8B5CF6, #7C3AED);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.header-icon svg {
  width: 28px;
  height: 28px;
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

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.time-range-selector {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: var(--bg-main);
  border-radius: var(--radius);
  border: 1px solid var(--border);
}

.selector-icon {
  width: 18px;
  height: 18px;
  color: var(--text-secondary);
}

.time-select {
  border: none;
  background: transparent;
  font-size: 0.875rem;
  color: var(--text-primary);
  cursor: pointer;
  outline: none;
  padding-right: 0.5rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.25rem;
  margin-bottom: 1.5rem;
}

.stat-card {
  background: var(--bg-card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  border: 1px solid var(--border);
  padding: 1.5rem;
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, var(--accent), transparent);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.stat-card:hover {
  box-shadow: var(--shadow-lg);
  transform: translateY(-2px);
}

.stat-card:hover::before {
  opacity: 1;
}

.stat-icon {
  width: 52px;
  height: 52px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.stat-icon svg {
  width: 26px;
  height: 26px;
}

.stat-icon-primary {
  background: linear-gradient(135deg, #8B5CF6, #7C3AED);
}

.stat-icon-success {
  background: linear-gradient(135deg, #10B981, #059669);
}

.stat-icon-warning {
  background: linear-gradient(135deg, #F59E0B, #D97706);
}

.stat-icon-info {
  background: linear-gradient(135deg, #3B82F6, #1D4ED8);
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
}

.stat-label {
  font-size: 0.8125rem;
  color: var(--text-secondary);
  margin-top: 0.5rem;
}

.stat-trend {
  display: flex;
  align-items: center;
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 600;
}

.trend-up {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
}

.trend-up svg {
  width: 14px;
  height: 14px;
}

.stat-percentage {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.125rem;
}

.percentage-value {
  font-size: 1.125rem;
  font-weight: 700;
  color: #10B981;
}

.percentage-label {
  font-size: 0.6875rem;
  color: var(--text-secondary);
}

.stat-indicator {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  background: rgba(245, 158, 11, 0.1);
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  color: #F59E0B;
}

.pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #F59E0B;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.stat-avatars {
  display: flex;
  align-items: center;
}

.avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: 2px solid var(--bg-card);
  margin-left: -8px;
}

.avatar-1 { background: linear-gradient(135deg, #8B5CF6, #7C3AED); }
.avatar-2 { background: linear-gradient(135deg, #3B82F6, #1D4ED8); }
.avatar-3 { background: linear-gradient(135deg, #10B981, #059669); }

.charts-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.25rem;
  margin-bottom: 1.5rem;
}

.chart-card {
  min-height: 380px;
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
  flex-direction: column;
  gap: 0.125rem;
}

.card-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.card-subtitle {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.utilization-badge {
  padding: 0.375rem 0.875rem;
  border-radius: 20px;
  font-size: 0.875rem;
  font-weight: 700;
}

.utilization-success {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
}

.utilization-warning {
  background: rgba(245, 158, 11, 0.1);
  color: #F59E0B;
}

.utilization-danger {
  background: rgba(239, 68, 68, 0.1);
  color: #EF4444;
}

.chart-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2rem;
  gap: 1.5rem;
}

.donut-chart {
  position: relative;
  width: 180px;
  height: 180px;
}

.donut-chart svg {
  width: 100%;
  height: 100%;
}

.utilization-circle {
  transition: stroke-dashoffset 1s ease;
}

.donut-center {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
}

.donut-value {
  display: block;
  font-size: 2rem;
  font-weight: 700;
  color: var(--text-primary);
}

.donut-label {
  display: block;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.utilization-legend {
  display: flex;
  gap: 1.5rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.legend-used {
  background: linear-gradient(135deg, #8B5CF6, #7C3AED);
}

.legend-available {
  background: #E2E8F0;
}

.category-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.5rem;
}

.category-item {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.category-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.category-icon svg {
  width: 18px;
  height: 18px;
}

.category-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.category-name {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.category-count {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.progress-container {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.progress-bar {
  flex: 1;
  height: 8px;
  background: var(--bg-main);
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.5s ease;
}

.progress-text {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
  min-width: 36px;
  text-align: right;
}

.empty-categories {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.empty-categories svg {
  width: 48px;
  height: 48px;
  margin-bottom: 0.75rem;
  opacity: 0.5;
}

.empty-categories p {
  margin: 0;
  font-size: 0.875rem;
}

.content-row {
  display: grid;
  grid-template-columns: 1.5fr 1fr;
  gap: 1.25rem;
}

.table-card {
  overflow: hidden;
}

.activity-card {
  max-height: 500px;
  display: flex;
  flex-direction: column;
}

.activity-card .card-header {
  flex-shrink: 0;
}

.loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  gap: 1rem;
  color: var(--text-secondary);
}

.spinner {
  width: 32px;
  height: 32px;
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
  padding: 3rem;
  color: var(--text-secondary);
}

.empty-state svg {
  width: 48px;
  height: 48px;
  margin-bottom: 0.75rem;
  opacity: 0.5;
}

.empty-state p {
  margin: 0;
  font-size: 0.875rem;
}

.stats-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.stats-table th,
.stats-table td {
  padding: 1rem 1.25rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

.stats-table td {
  /* Apply text overflow to all table cells */
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Ensure inline elements in table cells also truncate */
.stats-table td > *,
.stats-table td > span,
.stats-table td > a {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

.stats-table th {
  background: var(--bg-main);
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
}

.stats-table tbody tr {
  transition: background 0.2s ease;
}

.stats-table tbody tr:hover {
  background: var(--bg-secondary);
}

.stats-table tbody tr:last-child td {
  border-bottom: none;
}

.rank-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  font-weight: 700;
  font-size: 0.875rem;
  background: var(--bg-main);
  color: var(--text-secondary);
}

.rank-icon {
  display: flex;
}

.rank-icon svg {
  width: 12px;
  height: 12px;
}

.rank-gold {
  background: linear-gradient(135deg, #FCD34D, #F59E0B);
  color: white;
}

.rank-silver {
  background: linear-gradient(135deg, #E5E7EB, #9CA3AF);
  color: white;
}

.rank-bronze {
  background: linear-gradient(135deg, #D97706, #B45309);
  color: white;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #8B5CF6, #7C3AED);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 0.8125rem;
  font-weight: 600;
}

.user-name {
  font-weight: 500;
  color: var(--text-primary);
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
}

.badge-warning {
  background: rgba(245, 158, 11, 0.1);
  color: #F59E0B;
}

.badge-secondary {
  background: var(--bg-main);
  color: var(--text-secondary);
}

.number-value {
  font-weight: 600;
  color: var(--text-primary);
}

.duration-value {
  color: var(--text-primary);
  font-size: 0.875rem;
}

.time-value {
  color: var(--text-secondary);
  font-size: 0.8125rem;
}

.activity-list {
  flex: 1;
  overflow-y: auto;
  padding: 1rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.activity-item {
  display: flex;
  align-items: flex-start;
  gap: 0.875rem;
  padding: 0.875rem;
  background: var(--bg-main);
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
}

.activity-item:hover {
  background: var(--bg-secondary);
}

.activity-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.activity-icon svg {
  width: 18px;
  height: 18px;
}

.icon-success {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
}

.icon-info {
  background: rgba(59, 130, 246, 0.1);
  color: #3B82F6;
}

.icon-warning {
  background: rgba(245, 158, 11, 0.1);
  color: #F59E0B;
}

.icon-secondary {
  background: var(--bg-secondary);
  color: var(--text-secondary);
}

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-title {
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
}

.activity-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.activity-user {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.activity-user svg {
  width: 12px;
  height: 12px;
}

@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .content-row {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .header-actions {
    width: 100%;
  }

  .time-range-selector {
    flex: 1;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .charts-row {
    grid-template-columns: 1fr;
  }

  .stat-card {
    padding: 1.25rem;
  }

  .stat-value {
    font-size: 1.5rem;
  }
}
</style>
