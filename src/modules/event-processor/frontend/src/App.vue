<template>
  <div id="app">
    <!-- Dialog Manager for global alerts and confirms -->
    <DialogManager />
    <nav class="navbar">
      <div class="nav-container">
        <router-link to="/" class="nav-brand">
          <svg xmlns="http://www.w3.org/2000/svg" class="brand-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <span>质量网关</span>
        </router-link>
        
        <div class="nav-links">
          <router-link to="/events" class="nav-link">
            <svg xmlns="http://www.w3.org/2000/svg" class="nav-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
            </svg>
            事件处理
          </router-link>
          <router-link to="/resources" class="nav-link">
            <svg xmlns="http://www.w3.org/2000/svg" class="nav-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
            </svg>
            可执行资源
          </router-link>
          <router-link v-if="isAdmin" to="/console" class="nav-link">
            <svg xmlns="http://www.w3.org/2000/svg" class="nav-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            控制台
          </router-link>
          <router-link v-if="isLoggedIn" to="/resource-pool" class="nav-link">
            <svg xmlns="http://www.w3.org/2000/svg" class="nav-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            资源池管理
          </router-link>
        </div>
        
        <div class="nav-auth">
          <template v-if="isLoggedIn">
            <div class="user-info">
              <svg xmlns="http://www.w3.org/2000/svg" class="user-avatar" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
              <span class="username">{{ username }}</span>
            </div>
            <button @click="logout" class="btn-logout">
              <svg xmlns="http://www.w3.org/2000/svg" class="btn-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
              登出
            </button>
          </template>
          <template v-else>
            <router-link to="/login" class="btn-login">登录</router-link>
          </template>
        </div>
      </div>
    </nav>
    <main class="main-content">
      <router-view />
    </main>
    <footer class="app-footer">
      <div class="footer-container">
        <div class="footer-brand">
          <svg xmlns="http://www.w3.org/2000/svg" class="footer-logo" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <div class="brand-text">
            <h3>双引擎质量网关</h3>
            <p>企业级质量检查与资源管理平台</p>
          </div>
        </div>
        <div class="footer-links">
          <div class="footer-column">
            <h4>功能模块</h4>
            <ul>
              <li><router-link to="/events">事件处理</router-link></li>
              <li><router-link to="/resources">可执行资源</router-link></li>
              <li><router-link v-if="isLoggedIn" to="/resource-pool">资源池管理</router-link></li>
            </ul>
          </div>
          <div class="footer-column">
            <h4>快速链接</h4>
            <ul>
              <li><router-link to="/documentation">使用文档</router-link></li>
              <li><router-link to="/api-reference">API 参考</router-link></li>
            </ul>
          </div>
        </div>
      </div>
      <div class="footer-bottom">
        <p>&copy; 2026 双引擎质量网关. All rights reserved.</p>
      </div>
    </footer>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import DialogManager from './components/DialogManager.vue'

export default {
  name: 'App',
  components: {
    DialogManager
  },
  setup() {
    const router = useRouter()
    const isLoggedIn = ref(false)
    const username = ref('')
    const userRole = ref('')

    const isAdmin = computed(() => userRole.value === 'admin')

    const checkAuth = async () => {
      try {
        const response = await fetch('/api/auth/status', {
          credentials: 'include'
        })
        const data = await response.json()
        isLoggedIn.value = data.loggedIn
        if (data.loggedIn && data.user) {
          username.value = data.user.username
          userRole.value = data.user.role
          localStorage.setItem('isLoggedIn', 'true')
          localStorage.setItem('userRole', data.user.role)
          localStorage.setItem('username', data.user.username)
        } else {
          localStorage.removeItem('isLoggedIn')
          localStorage.removeItem('userRole')
          localStorage.removeItem('username')
        }
      } catch (error) {
        console.error('Failed to check auth status:', error)
      }
    }

    const logout = async () => {
      try {
        await fetch('/api/auth/logout', {
          method: 'POST',
          credentials: 'include'
        })
        localStorage.removeItem('isLoggedIn')
        localStorage.removeItem('userRole')
        localStorage.removeItem('username')
        window.location.href = '/'
      } catch (error) {
        console.error('Failed to logout:', error)
      }
    }

    onMounted(() => {
      checkAuth()
      isLoggedIn.value = localStorage.getItem('isLoggedIn') === 'true'
      username.value = localStorage.getItem('username') || ''
      userRole.value = localStorage.getItem('userRole') || ''
    })

    return {
      isLoggedIn,
      username,
      userRole,
      isAdmin,
      logout
    }
  }
}
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

:root {
  --primary: #171717;
  --primary-hover: #404040;
  --accent: #D4AF37;
  --accent-hover: #B8962F;
  --bg-main: #F8FAFC;
  --bg-card: #FFFFFF;
  --text-primary: #171717;
  --text-secondary: #64748B;
  --border: #E2E8F0;
  --success: #10B981;
  --success-bg: #D1FAE5;
  --danger: #EF4444;
  --danger-bg: #FEE2E2;
  --warning: #F59E0B;
  --warning-bg: #FEF3C7;
  --info: #3B82F6;
  --info-bg: #DBEAFE;
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px -1px rgba(0, 0, 0, 0.1);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -2px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -4px rgba(0, 0, 0, 0.1);
  --radius: 12px;
  --radius-sm: 8px;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif;
  background-color: var(--bg-main);
  color: var(--text-primary);
  line-height: 1.6;
}

#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  background: linear-gradient(135deg, #0C4A6E 0%, #0369A1 100%);
  box-shadow: 0 4px 20px -2px rgba(0, 0, 0, 0.15);
}

.nav-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 1.5rem;
  height: 64px;
}

.nav-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 1.25rem;
  font-weight: 700;
  color: white;
  text-decoration: none;
  transition: opacity 0.2s ease;
}

.nav-brand:hover {
  opacity: 0.9;
}

.brand-icon {
  width: 32px;
  height: 32px;
  color: #38BDF8;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  color: rgba(255, 255, 255, 0.8);
  text-decoration: none;
  font-size: 0.9rem;
  font-weight: 500;
  border-radius: 0.5rem;
  transition: all 0.2s ease;
}

.nav-icon {
  width: 18px;
  height: 18px;
}

.nav-link:hover {
  color: white;
  background: rgba(255, 255, 255, 0.1);
}

.nav-link.router-link-active {
  color: white;
  background: rgba(56, 189, 248, 0.2);
}

.nav-auth {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: rgba(255, 255, 255, 0.9);
}

.user-avatar {
  width: 32px;
  height: 32px;
  color: #38BDF8;
}

.username {
  font-weight: 500;
  font-size: 0.9rem;
}

.btn-logout {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.5rem 1rem;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-logout:hover {
  background: rgba(255, 255, 255, 0.2);
}

.btn-icon {
  width: 16px;
  height: 16px;
}

.btn-login {
  padding: 0.5rem 1.25rem;
  background: #0EA5E9;
  color: white;
  text-decoration: none;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn-login:hover {
  background: #0284C7;
  transform: translateY(-1px);
}

.main-content {
  flex: 1;
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

/* Footer Styles */
.app-footer {
  background: linear-gradient(180deg, #F8FAFC 0%, #F1F5F9 100%);
  border-top: 1px solid #E2E8F0;
  margin-top: auto;
}

.footer-container {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  max-width: 1200px;
  margin: 0 auto;
  padding: 3rem 1.5rem;
  gap: 3rem;
}

.footer-brand {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.footer-logo {
  width: 48px;
  height: 48px;
  color: #0EA5E9;
  flex-shrink: 0;
}

.brand-text h3 {
  font-size: 1.125rem;
  font-weight: 700;
  color: #0C4A6E;
  margin: 0 0 0.375rem 0;
}

.brand-text p {
  font-size: 0.875rem;
  color: #64748B;
  margin: 0;
}

.footer-links {
  display: flex;
  gap: 4rem;
}

.footer-column h4 {
  font-size: 0.875rem;
  font-weight: 600;
  color: #0C4A6E;
  margin-bottom: 1rem;
}

.footer-column ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.footer-column ul li {
  margin-bottom: 0.625rem;
}

.footer-column a {
  font-size: 0.875rem;
  color: #64748B;
  text-decoration: none;
  transition: color 0.2s ease;
}

.footer-column a:hover {
  color: #0EA5E9;
}

.footer-bottom {
  padding: 1.25rem 1.5rem;
  border-top: 1px solid #E2E8F0;
  text-align: center;
  background: white;
}

.footer-bottom p {
  font-size: 0.8125rem;
  color: #94A3B8;
  margin: 0;
}

.card {
  background: var(--bg-card);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
  border: 1px solid var(--border);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border);
}

.card-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: -0.025em;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.625rem 1.25rem;
  border-radius: var(--radius-sm);
  border: none;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn:focus {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.btn-primary {
  background-color: var(--primary);
  color: white;
}

.btn-primary:hover {
  background-color: var(--primary-hover);
}

.btn-danger {
  background-color: var(--danger);
  color: white;
}

.btn-danger:hover {
  background-color: #DC2626;
}

.btn-danger:disabled {
  background-color: #FCA5A5;
  cursor: not-allowed;
  opacity: 0.65;
}

.btn-success {
  background-color: var(--success);
  color: white;
}

.btn-success:hover {
  background-color: #059669;
}

.btn-success:disabled {
  background-color: #6EE7B7;
  cursor: not-allowed;
  opacity: 0.65;
}

.btn-secondary {
  background-color: var(--text-secondary);
  color: white;
}

.btn-secondary:hover {
  background-color: #475569;
}

.btn-secondary:disabled {
  background-color: #94A3B8;
  cursor: not-allowed;
  opacity: 0.65;
}

.btn:disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.8125rem;
}

.form-group {
  margin-bottom: 1.25rem;
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
  padding: 0.75rem 1rem;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 0.9375rem;
  transition: all 0.2s ease;
  background-color: var(--bg-card);
}

.form-control:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px rgba(212, 175, 55, 0.15);
}

.form-control::placeholder {
  color: #94A3B8;
}

select.form-control {
  appearance: none;
  background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
  background-repeat: no-repeat;
  background-position: right 0.75rem center;
  background-size: 20px;
  padding-right: 2.5rem;
}

table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  table-layout: fixed;
}

th, td {
  padding: 0.875rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

/* Apply text overflow to all table cells */
td {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Ensure inline elements in table cells also truncate */
td > *,
td > span,
td > a {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: bottom;
}

th {
  background-color: var(--bg-main);
  font-weight: 600;
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
  white-space: nowrap;
}

th:first-child {
  border-top-left-radius: var(--radius-sm);
}

th:last-child {
  border-top-right-radius: var(--radius-sm);
}

tr:hover {
  background-color: var(--bg-main);
}

tr:last-child td {
  border-bottom: none;
}

.badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.025em;
}

.badge-success {
  background-color: var(--success-bg);
  color: #047857;
}

.badge-danger {
  background-color: var(--danger-bg);
  color: #B91C1C;
}

.badge-warning {
  background-color: var(--warning-bg);
  color: #B45309;
}

.badge-info {
  background-color: var(--info-bg);
  color: #1D4ED8;
}

.badge-secondary {
  background-color: #F1F5F9;
  color: #475569;
}

.tabs {
  display: flex;
  gap: 0.5rem;
  border-bottom: 2px solid var(--border);
  margin-bottom: 1.5rem;
}

.tab {
  padding: 0.75rem 1.5rem;
  cursor: pointer;
  border: none;
  background: none;
  font-size: 0.9375rem;
  font-weight: 500;
  color: var(--text-secondary);
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: all 0.2s ease;
}

.tab.active {
  color: var(--accent);
  border-bottom-color: var(--accent);
}

.tab:hover {
  color: var(--primary);
}

.pagination {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
  margin-top: 1.5rem;
}

.pagination button {
  padding: 0.5rem 1rem;
  border: 1px solid var(--border);
  background: var(--bg-card);
  cursor: pointer;
  border-radius: var(--radius-sm);
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-secondary);
  transition: all 0.2s ease;
}

.pagination button:focus {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination button.active {
  background-color: var(--primary);
  color: white;
  border-color: var(--primary);
}

.pagination button:not(:disabled):hover {
  border-color: var(--primary);
  color: var(--primary);
}

.alert {
  padding: 1rem 1.25rem;
  border-radius: var(--radius-sm);
  margin-bottom: 1.5rem;
  font-size: 0.9375rem;
}

.alert-success {
  background-color: var(--success-bg);
  color: #047857;
  border: 1px solid #A7F3D0;
}

.alert-danger {
  background-color: var(--danger-bg);
  color: #B91C1C;
  border: 1px solid #FECACA;
}

.alert-warning {
  background-color: var(--warning-bg);
  color: #B45309;
  border: 1px solid #FDE68A;
}

.alert-info {
  background-color: var(--info-bg);
  color: #1D4ED8;
  border: 1px solid #BFDBFE;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal {
  background: var(--bg-card);
  border-radius: var(--radius);
  padding: 1.5rem;
  max-width: 600px;
  width: 90%;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: var(--shadow-lg);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid var(--border);
  background-color: var(--bg-card);
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--text-secondary);
  line-height: 1;
  padding: 0.25rem;
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
}

.modal-close:hover {
  background-color: var(--bg-main);
  color: var(--primary);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border);
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 3rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--border);
  border-top: 3px solid var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.empty-state {
  text-align: center;
  padding: 3rem;
  color: var(--text-secondary);
}

.empty-state p {
  font-size: 0.9375rem;
}

.modal-large {
  max-width: 900px;
}

@media (max-width: 768px) {
  .navbar {
    padding: 0 1rem;
    height: 56px;
  }
  
  .nav-brand a {
    font-size: 1rem;
  }
  
  .nav-links {
    gap: 0.25rem;
  }
  
  .nav-links a {
    padding: 0.375rem 0.75rem;
    font-size: 0.8125rem;
  }
  
  .main-content {
    padding: 1rem;
  }
  
  .card {
    padding: 1rem;
    margin-bottom: 1rem;
  }
  
  th, td {
    padding: 0.625rem 0.75rem;
    font-size: 0.8125rem;
  }

  .footer-content {
    flex-direction: column;
    padding: 1.5rem;
    gap: 1.5rem;
  }

  .footer-bottom {
    padding: 0.75rem 1rem;
  }
}
</style>
