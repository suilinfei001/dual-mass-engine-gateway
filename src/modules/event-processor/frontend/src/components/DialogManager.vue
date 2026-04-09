<template>
  <div class="dialog-manager">
    <!-- Alert Dialog -->
    <Transition name="dialog-fade">
      <div
        v-if="alertState.show"
        class="dialog-overlay"
        :aria-hidden="!alertState.show"
        @click.self="closeAlert"
        @keydown.esc="closeAlert"
      >
        <div
          class="dialog dialog-alert"
          :class="['dialog-' + alertState.type, 'dialog-' + alertState.size]"
          role="alertdialog"
          :aria-labelledby="'alert-title-' + alertState.id"
          :aria-describedby="'alert-desc-' + alertState.id"
          tabindex="-1"
        >
          <div class="dialog-icon-wrapper" :class="'icon-' + alertState.type">
            <svg v-if="alertState.type === 'success'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <svg v-else-if="alertState.type === 'error'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <svg v-else-if="alertState.type === 'warning'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div class="dialog-content">
            <h3 :id="'alert-title-' + alertState.id" class="dialog-title">{{ alertState.title || defaultTitles[alertState.type] }}</h3>
            <p :id="'alert-desc-' + alertState.id" class="dialog-message">{{ alertState.message }}</p>
          </div>
          <div class="dialog-actions">
            <button
              class="btn-dialog"
              :class="'btn-' + (alertState.type === 'error' ? 'danger' : alertState.type === 'success' ? 'success' : 'primary')"
              @click="closeAlert"
              ref="alertButton"
            >
              {{ alertState.confirmText || '确定' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Confirm Dialog -->
    <Transition name="dialog-fade">
      <div
        v-if="confirmState.show"
        class="dialog-overlay"
        :aria-hidden="!confirmState.show"
        @click.self="closeConfirm(false)"
        @keydown.esc="closeConfirm(false)"
      >
        <div
          class="dialog dialog-confirm"
          :class="['dialog-' + confirmState.type, 'dialog-' + confirmState.size]"
          role="alertdialog"
          :aria-labelledby="'confirm-title-' + confirmState.id"
          :aria-describedby="'confirm-desc-' + confirmState.id"
          tabindex="-1"
        >
          <div class="dialog-icon-wrapper" :class="'icon-' + confirmState.type">
            <svg v-if="confirmState.type === 'danger'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <svg v-else-if="confirmState.type === 'warning'" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
          <div class="dialog-content">
            <h3 :id="'confirm-title-' + confirmState.id" class="dialog-title">{{ confirmState.title || defaultTitles[confirmState.type] }}</h3>
            <p :id="'confirm-desc-' + confirmState.id" class="dialog-message">{{ confirmState.message }}</p>
          </div>
          <div class="dialog-actions">
            <button
              class="btn-dialog btn-secondary"
              @click="closeConfirm(false)"
              ref="cancelButton"
            >
              {{ confirmState.cancelText || '取消' }}
            </button>
            <button
              class="btn-dialog"
              :class="'btn-' + confirmState.type"
              @click="closeConfirm(true)"
              ref="confirmButton"
            >
              {{ confirmState.confirmText || '确定' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script>
import { reactive, ref, nextTick, watch } from 'vue'

let dialogIdCounter = 0

const alertState = reactive({
  show: false,
  title: '',
  message: '',
  type: 'info',
  confirmText: '确定',
  size: 'medium',
  resolve: null,
  id: 0
})

const confirmState = reactive({
  show: false,
  title: '',
  message: '',
  type: 'info',
  confirmText: '确定',
  cancelText: '取消',
  size: 'medium',
  resolve: null,
  id: 0
})

const defaultTitles = {
  success: '操作成功',
  error: '操作失败',
  warning: '注意',
  info: '提示',
  danger: '危险操作'
}

const alertButton = ref(null)
const cancelButton = ref(null)
const confirmButton = ref(null)

function focusElement(el) {
  nextTick(() => {
    if (el && el.value) {
      el.value.focus()
    }
  })
}

function showAlertInternal(options) {
  if (typeof options === 'string') {
    options = { message: options }
  }

  alertState.title = options.title || ''
  alertState.message = options.message || ''
  alertState.type = options.type || 'info'
  alertState.confirmText = options.confirmText || '确定'
  alertState.size = options.size || 'medium'
  alertState.id = ++dialogIdCounter
  alertState.show = true

  focusElement(alertButton)

  return new Promise((resolve) => {
    alertState.resolve = resolve
  })
}

function showConfirmInternal(options) {
  if (typeof options === 'string') {
    options = { message: options }
  }

  confirmState.title = options.title || ''
  confirmState.message = options.message || ''
  confirmState.type = options.type || 'info'
  confirmState.confirmText = options.confirmText || '确定'
  confirmState.cancelText = options.cancelText || '取消'
  confirmState.size = options.size || 'medium'
  confirmState.id = ++dialogIdCounter
  confirmState.show = true

  focusElement(confirmButton)

  return new Promise((resolve) => {
    confirmState.resolve = resolve
  })
}

function closeAlert() {
  alertState.show = false
  if (alertState.resolve) {
    alertState.resolve(true)
    alertState.resolve = null
  }
}

function closeConfirm(result) {
  confirmState.show = false
  if (confirmState.resolve) {
    confirmState.resolve(result)
    confirmState.resolve = null
  }
}

export function showAlert(options) {
  return showAlertInternal(options)
}

export function showConfirm(options) {
  return showConfirmInternal(options)
}

export default {
  name: 'DialogManager',
  setup() {
    return {
      alertState,
      confirmState,
      defaultTitles,
      closeAlert,
      closeConfirm,
      alertButton,
      cancelButton,
      confirmButton
    }
  }
}
</script>

<style scoped>
.dialog-manager {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 9999;
}

.dialog-overlay {
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
  padding: 1rem;
  pointer-events: auto;
}

.dialog {
  background: white;
  border-radius: 16px;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04), 0 0 0 1px rgba(0, 0, 0, 0.05);
  width: 100%;
  pointer-events: auto;
  transform: translateY(0) scale(1);
  opacity: 1;
}

.dialog-small {
  max-width: 360px;
}

.dialog-medium {
  max-width: 420px;
}

.dialog-large {
  max-width: 520px;
}

/* Alert specific styles */
.dialog-alert {
  padding: 0;
  overflow: hidden;
}

.dialog-alert .dialog-icon-wrapper {
  width: 100%;
  height: 72px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dialog-alert .dialog-icon-wrapper svg {
  width: 40px;
  height: 40px;
}

.dialog-alert .icon-success {
  background: linear-gradient(135deg, #D1FAE5 0%, #A7F3D0 100%);
}

.dialog-alert .icon-success svg {
  color: #059669;
}

.dialog-alert .icon-error {
  background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
}

.dialog-alert .icon-error svg {
  color: #DC2626;
}

.dialog-alert .icon-warning {
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
}

.dialog-alert .icon-warning svg {
  color: #D97706;
}

.dialog-alert .icon-info {
  background: linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%);
}

.dialog-alert .icon-info svg {
  color: #2563EB;
}

.dialog-alert .dialog-content {
  padding: 1.25rem 1.5rem 0.5rem;
  text-align: center;
}

.dialog-alert .dialog-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 0.5rem 0;
  line-height: 1.4;
}

.dialog-alert .dialog-message {
  font-size: 0.9375rem;
  color: #64748B;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.5;
}

.dialog-alert .dialog-actions {
  padding: 1rem 1.5rem 1.5rem;
  display: flex;
  justify-content: center;
}

/* Confirm specific styles */
.dialog-confirm {
  padding: 1.5rem;
}

.dialog-confirm .dialog-icon-wrapper {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1rem;
}

.dialog-confirm .dialog-icon-wrapper svg {
  width: 28px;
  height: 28px;
}

.dialog-confirm .icon-danger {
  background: linear-gradient(135deg, #FEE2E2 0%, #FECACA 100%);
}

.dialog-confirm .icon-danger svg {
  color: #DC2626;
}

.dialog-confirm .icon-warning {
  background: linear-gradient(135deg, #FEF3C7 0%, #FDE68A 100%);
}

.dialog-confirm .icon-warning svg {
  color: #D97706;
}

.dialog-confirm .icon-info {
  background: linear-gradient(135deg, #DBEAFE 0%, #BFDBFE 100%);
}

.dialog-confirm .icon-info svg {
  color: #2563EB;
}

.dialog-confirm .dialog-content {
  text-align: center;
}

.dialog-confirm .dialog-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 0.75rem 0;
  line-height: 1.4;
}

.dialog-confirm .dialog-message {
  font-size: 0.9375rem;
  color: #64748B;
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.5;
}

.dialog-confirm .dialog-actions {
  margin-top: 1.5rem;
  display: flex;
  gap: 0.75rem;
  justify-content: center;
}

/* Button styles */
.btn-dialog {
  min-width: 100px;
  min-height: 44px;
  padding: 0.625rem 1.5rem;
  border-radius: 10px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: inherit;
}

.btn-dialog:focus-visible {
  outline: 2px solid currentColor;
  outline-offset: 2px;
}

.btn-dialog:active {
  transform: scale(0.97);
}

/* Primary button */
.btn-primary {
  background: linear-gradient(135deg, #3B82F6 0%, #2563EB 100%);
  color: white;
}

.btn-primary:hover {
  background: linear-gradient(135deg, #2563EB 0%, #1D4ED8 100%);
}

.btn-primary:focus-visible {
  outline-color: #3B82F6;
}

/* Secondary button */
.btn-secondary {
  background: #F1F5F9;
  color: #475569;
}

.btn-secondary:hover {
  background: #E2E8F0;
}

.btn-secondary:focus-visible {
  outline-color: #64748B;
}

/* Danger button */
.btn-danger {
  background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%);
  color: white;
}

.btn-danger:hover {
  background: linear-gradient(135deg, #DC2626 0%, #B91C1C 100%);
}

.btn-danger:focus-visible {
  outline-color: #EF4444;
}

/* Warning button */
.btn-warning {
  background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%);
  color: white;
}

.btn-warning:hover {
  background: linear-gradient(135deg, #D97706 0%, #B45309 100%);
}

.btn-warning:focus-visible {
  outline-color: #F59E0B;
}

/* Success button */
.btn-success {
  background: linear-gradient(135deg, #10B981 0%, #059669 100%);
  color: white;
}

.btn-success:hover {
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
}

.btn-success:focus-visible {
  outline-color: #10B981;
}

/* Info button */
.btn-info {
  background: linear-gradient(135deg, #3B82F6 0%, #2563EB 100%);
  color: white;
}

.btn-info:hover {
  background: linear-gradient(135deg, #2563EB 0%, #1D4ED8 100%);
}

.btn-info:focus-visible {
  outline-color: #3B82F6;
}

/* Transition animations */
.dialog-fade-enter-active,
.dialog-fade-leave-active {
  transition: opacity 0.2s ease;
}

.dialog-fade-enter-from,
.dialog-fade-leave-to {
  opacity: 0;
}

.dialog-fade-enter-active .dialog,
.dialog-fade-leave-active .dialog {
  transition: transform 0.25s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.2s ease;
}

.dialog-fade-enter-from .dialog,
.dialog-fade-leave-to .dialog {
  transform: translateY(-20px) scale(0.95);
  opacity: 0;
}

/* Reduced motion support */
@media (prefers-reduced-motion: reduce) {
  .dialog-fade-enter-active,
  .dialog-fade-leave-active,
  .dialog-fade-enter-active .dialog,
  .dialog-fade-leave-active .dialog {
    transition: opacity 0.15s ease;
  }

  .dialog-fade-enter-from .dialog,
  .dialog-fade-leave-to .dialog {
    transform: none;
  }

  .btn-dialog {
    transition: none;
  }

  .btn-dialog:active {
    transform: none;
  }
}

/* Responsive */
@media (max-width: 480px) {
  .dialog {
    max-width: calc(100% - 1rem);
  }

  .dialog-confirm .dialog-actions {
    flex-direction: column-reverse;
  }

  .btn-dialog {
    width: 100%;
  }
}
</style>
