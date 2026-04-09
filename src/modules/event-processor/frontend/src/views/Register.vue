<template>
  <div class="auth-page">
    <div class="auth-container">
      <div class="auth-branding">
        <div class="branding-content">
          <svg xmlns="http://www.w3.org/2000/svg" class="branding-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <h1>双引擎质量网关</h1>
          <p>企业级质量检查与资源管理平台</p>
          <div class="branding-features">
            <div class="feature-item">
              <svg xmlns="http://www.w3.org/2000/svg" class="feature-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              <span>事件驱动</span>
            </div>
            <div class="feature-item">
              <svg xmlns="http://www.w3.org/2000/svg" class="feature-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
              </svg>
              <span>智能检测</span>
            </div>
            <div class="feature-item">
              <svg xmlns="http://www.w3.org/2000/svg" class="feature-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v2a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 13a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H5a1 1 0 01-1-1v-6zM16 13a1 1 0 011-1h2a1 1 0 011 1v6a1 1 0 01-1 1h-2a1 1 0 01-1-1v-6z" />
              </svg>
              <span>资源管理</span>
            </div>
          </div>
        </div>
      </div>
      
      <div class="auth-form-section">
        <div class="auth-card">
          <div class="auth-header">
            <h2>创建账号</h2>
            <p>开始您的质量检查之旅</p>
          </div>

          <div v-if="error" class="alert alert-danger">
            <svg xmlns="http://www.w3.org/2000/svg" class="alert-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ error }}
          </div>

          <div v-if="success" class="alert alert-success">
            <svg xmlns="http://www.w3.org/2000/svg" class="alert-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            {{ success }}
          </div>

          <form @submit.prevent="handleRegister" class="auth-form">
            <div class="form-group">
              <label for="email">邮箱</label>
              <div class="input-wrapper">
                <svg xmlns="http://www.w3.org/2000/svg" class="input-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                </svg>
                <input
                  type="email"
                  id="email"
                  v-model="form.email"
                  class="form-control"
                  required
                  placeholder="请输入邮箱（必须以 @aishu.cn 结尾）"
                />
              </div>
              <small class="form-hint">邮箱必须以 @aishu.cn 结尾</small>
            </div>

            <div class="form-group">
              <label for="password">密码</label>
              <div class="input-wrapper">
                <svg xmlns="http://www.w3.org/2000/svg" class="input-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <input
                  type="password"
                  id="password"
                  v-model="form.password"
                  class="form-control"
                  required
                  placeholder="请输入密码（至少 6 个字符）"
                />
              </div>
              <small class="form-hint">至少 6 个字符</small>
            </div>

            <div class="form-group">
              <label for="confirmPassword">确认密码</label>
              <div class="input-wrapper">
                <svg xmlns="http://www.w3.org/2000/svg" class="input-icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <input
                  type="password"
                  id="confirmPassword"
                  v-model="form.confirmPassword"
                  class="form-control"
                  required
                  placeholder="请再次输入密码"
                />
              </div>
            </div>

            <div class="form-actions">
              <button type="submit" class="btn btn-primary btn-block" :disabled="loading">
                <span v-if="loading" class="spinner"></span>
                {{ loading ? '注册中...' : '注册' }}
              </button>
            </div>
          </form>

          <div class="auth-footer">
            <p>已有账号？<router-link to="/login">立即登录</router-link></p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

export default {
  name: 'Register',
  setup() {
    const router = useRouter()
    const form = ref({
      email: '',
      password: '',
      confirmPassword: ''
    })
    const loading = ref(false)
    const error = ref('')
    const success = ref('')

    const handleRegister = async () => {
      error.value = ''
      success.value = ''

      if (!form.value.email.endsWith('@aishu.cn')) {
        error.value = '邮箱必须以 @aishu.cn 结尾'
        return
      }

      if (form.value.password.length < 6) {
        error.value = '密码至少需要 6 个字符'
        return
      }

      if (form.value.password !== form.value.confirmPassword) {
        error.value = '两次输入的密码不一致'
        return
      }

      loading.value = true

      try {
        const response = await fetch('/api/auth/register', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          credentials: 'include',
          body: JSON.stringify({
            email: form.value.email,
            password: form.value.password
          })
        })

        const data = await response.json()

        if (data.success) {
          success.value = '注册成功！您现在可以使用用户名登录：' + data.user.username
          setTimeout(() => {
            router.push('/login')
          }, 2000)
        } else {
          error.value = data.message || '注册失败'
        }
      } catch (err) {
        error.value = '发生错误，请重试。'
      } finally {
        loading.value = false
      }
    }

    return {
      form,
      loading,
      error,
      success,
      handleRegister
    }
  }
}
</script>

<style scoped>
.auth-page {
  min-height: calc(100vh - 120px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background: linear-gradient(135deg, #F0F9FF 0%, #E0F2FE 50%, #F0F9FF 100%);
}

.auth-container {
  display: flex;
  max-width: 1000px;
  width: 100%;
  background: white;
  border-radius: 1.5rem;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.15);
  overflow: hidden;
}

.auth-branding {
  flex: 1;
  background: linear-gradient(135deg, #0C4A6E 0%, #0369A1 50%, #0EA5E9 100%);
  padding: 3rem;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
}

.auth-branding::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255,255,255,0.1) 0%, transparent 50%);
  animation: shimmer 15s infinite linear;
}

@keyframes shimmer {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

.branding-content {
  position: relative;
  z-index: 1;
  text-align: center;
  color: white;
}

.branding-icon {
  width: 80px;
  height: 80px;
  color: #38BDF8;
  margin-bottom: 1.5rem;
}

.branding-content h1 {
  font-size: 1.75rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.branding-content p {
  font-size: 1rem;
  opacity: 0.9;
  margin-bottom: 2rem;
}

.branding-features {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.feature-item {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  font-size: 0.95rem;
  opacity: 0.9;
}

.feature-icon {
  width: 20px;
  height: 20px;
  color: #38BDF8;
}

.auth-form-section {
  flex: 1;
  padding: 3rem;
  display: flex;
  align-items: center;
}

.auth-card {
  width: 100%;
  max-width: 360px;
  margin: 0 auto;
}

.auth-header {
  text-align: center;
  margin-bottom: 2rem;
}

.auth-header h2 {
  font-size: 1.75rem;
  font-weight: 700;
  color: #0C4A6E;
  margin-bottom: 0.5rem;
}

.auth-header p {
  color: #64748B;
  font-size: 0.95rem;
}

.alert {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
  font-size: 0.875rem;
}

.alert-danger {
  background: #FEF2F2;
  color: #DC2626;
  border: 1px solid #FECACA;
}

.alert-success {
  background: #F0FDF4;
  color: #16A34A;
  border: 1px solid #BBF7D0;
}

.alert-icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.auth-form .form-group {
  margin-bottom: 1.25rem;
}

.auth-form label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #374151;
  margin-bottom: 0.5rem;
}

.form-hint {
  display: block;
  margin-top: 0.375rem;
  color: #6B7280;
  font-size: 0.8125rem;
}

.input-wrapper {
  position: relative;
}

.input-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  color: #9CA3AF;
  pointer-events: none;
}

.form-control {
  width: 100%;
  padding: 0.75rem 1rem 0.75rem 2.75rem;
  border: 1px solid #D1D5DB;
  border-radius: 0.5rem;
  font-size: 0.95rem;
  transition: all 0.2s ease;
  background: #F9FAFB;
}

.form-control:focus {
  outline: none;
  border-color: #0EA5E9;
  box-shadow: 0 0 0 3px rgba(14, 165, 233, 0.15);
  background: white;
}

.form-control::placeholder {
  color: #9CA3AF;
}

.form-actions {
  margin-top: 1.5rem;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  font-size: 0.95rem;
  font-weight: 500;
  border-radius: 0.5rem;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
}

.btn-block {
  width: 100%;
}

.btn-primary {
  background: linear-gradient(135deg, #0EA5E9, #0284C7);
  color: white;
  box-shadow: 0 4px 14px 0 rgba(14, 165, 233, 0.35);
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px 0 rgba(14, 165, 233, 0.25);
}

.btn-primary:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.auth-footer {
  text-align: center;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid #E5E7EB;
}

.auth-footer p {
  color: #6B7280;
  font-size: 0.875rem;
}

.auth-footer a {
  color: #0EA5E9;
  text-decoration: none;
  font-weight: 500;
}

.auth-footer a:hover {
  text-decoration: underline;
}

@media (max-width: 768px) {
  .auth-container {
    flex-direction: column;
    max-width: 400px;
  }
  
  .auth-branding {
    padding: 2rem;
  }
  
  .branding-features {
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: center;
  }
  
  .auth-form-section {
    padding: 2rem;
  }
}
</style>
