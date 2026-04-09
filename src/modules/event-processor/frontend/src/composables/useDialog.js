import { showAlert as showAlertInternal, showConfirm as showConfirmInternal } from '../components/DialogManager.vue'

/**
 * 显示一个简单的提示对话框
 * @param {string|object} message - 消息内容或配置对象
 * @param {object} options - 配置选项
 * @returns {Promise<boolean>}
 */
export function alert(message, options = {}) {
  if (typeof message === 'object') {
    options = message
    message = options.message
  }
  return showAlertInternal({
    message,
    ...options
  })
}

/**
 * 显示一个成功提示对话框
 * @param {string} message - 消息内容
 * @param {object} options - 配置选项
 * @returns {Promise<boolean>}
 */
export function alertSuccess(message, options = {}) {
  return showAlertInternal({
    message,
    type: 'success',
    ...options
  })
}

/**
 * 显示一个错误提示对话框
 * @param {string} message - 消息内容
 * @param {object} options - 配置选项
 * @returns {Promise<boolean>}
 */
export function alertError(message, options = {}) {
  return showAlertInternal({
    message,
    type: 'error',
    ...options
  })
}

/**
 * 显示一个警告提示对话框
 * @param {string} message - 消息内容
 * @param {object} options - 配置选项
 * @returns {Promise<boolean>}
 */
export function alertWarning(message, options = {}) {
  return showAlertInternal({
    message,
    type: 'warning',
    ...options
  })
}

/**
 * 显示一个确认对话框
 * @param {string|object} message - 消息内容或配置对象
 * @param {object} options - 配置选项
 * @returns {Promise<boolean>} - 用户选择结果
 */
export function confirm(message, options = {}) {
  if (typeof message === 'object') {
    options = message
    message = options.message
  }
  return showConfirmInternal({
    message,
    ...options
  })
}

/**
 * 显示一个危险操作的确认对话框
 * @param {string} message - 消息内容
 * @param {object} options - 配置选项
 * @returns {Promise<boolean>}
 */
export function confirmDanger(message, options = {}) {
  return showConfirmInternal({
    message,
    type: 'danger',
    confirmText: '确定',
    cancelText: '取消',
    ...options
  })
}

/**
 * 显示一个警告操作的确认对话框
 * @param {string} message - 消息内容
 * @param {object} options - 配置选项
 * @returns {Promise<boolean>}
 */
export function confirmWarning(message, options = {}) {
  return showConfirmInternal({
    message,
    type: 'warning',
    confirmText: '确定',
    cancelText: '取消',
    ...options
  })
}

/**
 * 对话框 composable
 * 在 Vue 组件中使用 useDialog() 获取对话框方法
 *
 * @example
 * import { useDialog } from '@/composables/useDialog'
 *
 * export default {
 *   setup() {
 *     const dialog = useDialog()
 *
 *     const handleDelete = async () => {
 *       const confirmed = await dialog.confirmDanger('确定要删除吗？')
 *       if (confirmed) {
 *         // 执行删除操作
 *       }
 *     }
 *
 *     return { handleDelete }
 *   }
 * }
 */
export function useDialog() {
  return {
    alert,
    alertSuccess,
    alertError,
    alertWarning,
    confirm,
    confirmDanger,
    confirmWarning
  }
}

export default useDialog
