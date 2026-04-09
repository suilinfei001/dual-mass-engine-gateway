// 认证工具 - 真正调用后端 API
// API 基础路径 - 使用相对路径，由 Vite 代理到 auth-service
const API_BASE = '/api'

// 默认管理员凭证（用于提示）
const DEFAULT_CREDENTIALS = {
  username: 'admin',
  password: 'admin123'
}

export const AuthUtils = {
  // 登录 API 调用
  async login(username, password) {
    try {
      const response = await fetch(`${API_BASE}/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include', // 发送和接收 cookies
        body: JSON.stringify({ username, password })
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error('Login error:', error)
      return {
        success: false,
        message: '网络错误，请稍后重试'
      }
    }
  },

  // 登出 API 调用
  async logout() {
    try {
      const response = await fetch(`${API_BASE}/logout`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        credentials: 'include' // 发送和接收 cookies
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error('Logout error:', error)
      return {
        success: false,
        message: '网络错误，请稍后重试'
      }
    }
  },

  // 检查登录状态 API 调用
  async checkLogin() {
    try {
      const response = await fetch(`${API_BASE}/check-login`, {
        method: 'GET',
        credentials: 'include' // 发送和接收 cookies
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const data = await response.json()
      return data
    } catch (error) {
      console.error('Check login error:', error)
      return {
        is_logged_in: false,
        username: null
      }
    }
  },

  // 获取默认凭证（用于提示）
  getDefaultCredentials() {
    return DEFAULT_CREDENTIALS
  }
}

export default AuthUtils
