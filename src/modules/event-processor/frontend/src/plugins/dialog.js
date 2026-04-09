import { alert, alertSuccess, alertError, alertWarning, confirm, confirmDanger, confirmWarning } from '../composables/useDialog.js'

export default {
  install(app) {
    // Add global dialog methods
    app.config.globalProperties.$alert = alert
    app.config.globalProperties.$alertSuccess = alertSuccess
    app.config.globalProperties.$alertError = alertError
    app.config.globalProperties.$alertWarning = alertWarning
    app.config.globalProperties.$confirm = confirm
    app.config.globalProperties.$confirmDanger = confirmDanger
    app.config.globalProperties.$confirmWarning = confirmWarning

    // Also provide as provide/inject for composition API
    app.provide('dialog', {
      alert,
      alertSuccess,
      alertError,
      alertWarning,
      confirm,
      confirmDanger,
      confirmWarning
    })
  }
}
