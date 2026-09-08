import { reactive } from 'vue'

/**
 * 轻量反馈中枢：Toast 消息 + 确认/输入弹窗
 * 由 FeedbackHost.vue 渲染，业务代码仅调用本模块导出的 API。
 * - message.success / warning / error / info  → 替代 ElMessage
 * - confirmDialog(...) / promptDialog(...)    → 替代 ElMessageBox.confirm / prompt
 */

let toastSeed = 0
let boxSeed = 0

export const feedbackState = reactive({
  toasts: [],
  box: null
})

function pushToast(type, content, options = {}) {
  if (content === undefined || content === null || content === '') return
  const id = ++toastSeed
  feedbackState.toasts.push({
    id,
    type,
    content: String(content),
    duration: options.duration ?? 2600
  })
  if ((options.duration ?? 2600) > 0) {
    setTimeout(() => dismissToast(id), options.duration ?? 2600)
  }
  return id
}

export function dismissToast(id) {
  const idx = feedbackState.toasts.findIndex((t) => t.id === id)
  if (idx > -1) feedbackState.toasts.splice(idx, 1)
}

export const message = {
  success: (content, options) => pushToast('success', content, options),
  warning: (content, options) => pushToast('warning', content, options),
  error: (content, options) => pushToast('error', content, options),
  info: (content, options) => pushToast('info', content, options)
}

function openBox(mode, content, title, options = {}) {
  return new Promise((resolve, reject) => {
    feedbackState.box = {
      id: ++boxSeed,
      mode,
      content: String(content ?? ''),
      title: title || '提示',
      type: options.type || 'info',
      confirmText: options.confirmButtonText || '确定',
      cancelText: options.cancelButtonText || '取消',
      inputValue: options.inputValue ?? '',
      inputPlaceholder: options.inputPlaceholder || '',
      inputPattern: options.inputPattern || null,
      inputErrorMessage: options.inputErrorMessage || '输入格式不正确',
      resolve,
      reject
    }
  })
}

export function confirmDialog(content, title, options) {
  return openBox('confirm', content, title, options)
}

export function promptDialog(content, title, options) {
  return openBox('prompt', content, title, options)
}

export function resolveBox(box, value) {
  box.resolve(value)
  feedbackState.box = null
}

export function rejectBox(box) {
  box.reject('cancel')
  feedbackState.box = null
}
