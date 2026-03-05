<template>
  <div id="app">
    <nav class="navbar">
      <div class="nav-brand">
        <router-link to="/">Event Processor</router-link>
      </div>
      <div class="nav-links">
        <router-link to="/events">Events</router-link>
        <router-link to="/resources">Resources</router-link>
        <router-link v-if="isAdmin" to="/console">Console</router-link>
      </div>
      <div class="nav-auth">
        <template v-if="isLoggedIn">
          <span class="username">{{ username }}</span>
          <button @click="logout" class="btn-logout">Logout</button>
        </template>
        <template v-else>
          <router-link to="/login" class="btn-login">Login</router-link>
          <router-link to="/register" class="btn-register">Register</router-link>
        </template>
      </div>
    </nav>
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

export default {
  name: 'App',
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
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 2rem;
  height: 64px;
  background: var(--bg-card);
  color: var(--text-primary);
  box-shadow: var(--shadow);
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-brand a {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--primary);
  text-decoration: none;
  letter-spacing: -0.025em;
}

.nav-brand a span {
  color: var(--accent);
}

.nav-links {
  display: flex;
  gap: 0.5rem;
}

.nav-links a {
  color: var(--text-secondary);
  text-decoration: none;
  padding: 0.5rem 1rem;
  border-radius: var(--radius-sm);
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.nav-links a:hover {
  color: var(--primary);
  background-color: var(--bg-main);
}

.nav-links a.router-link-active {
  color: var(--primary);
  background-color: #FEF3C7;
}

.nav-auth {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.username {
  font-weight: 500;
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.btn-login,
.btn-register,
.btn-logout {
  padding: 0.5rem 1.25rem;
  border-radius: var(--radius-sm);
  text-decoration: none;
  cursor: pointer;
  border: none;
  font-size: 0.875rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn-login:focus,
.btn-register:focus,
.btn-logout:focus {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.btn-login,
.btn-register {
  background-color: var(--primary);
  color: white;
}

.btn-login:hover,
.btn-register:hover {
  background-color: var(--primary-hover);
}

.btn-logout {
  background-color: transparent;
  color: var(--text-secondary);
  border: 1px solid var(--border);
}

.btn-logout:hover {
  background-color: var(--danger-bg);
  color: var(--danger);
  border-color: var(--danger);
}

.main-content {
  flex: 1;
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
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
}

th, td {
  padding: 0.875rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--border);
}

th {
  background-color: var(--bg-main);
  font-weight: 600;
  font-size: 0.8125rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-secondary);
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
}
</style>
