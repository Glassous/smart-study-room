/**
 * v-loading 指令：为宿主元素叠加加载遮罩
 * 用法与 Element Plus 的 v-loading 一致：<div v-loading="loading">…</div>
 */

function ensureMask(el) {
  if (!el.__loadingMask) {
    const mask = document.createElement('div')
    mask.className = 'v-loading-mask'
    const spinner = document.createElement('span')
    spinner.className = 'v-loading-spinner'
    mask.appendChild(spinner)
    el.__loadingMask = mask
  }
  return el.__loadingMask
}

function toggle(el, value) {
  const mask = ensureMask(el)
  if (value && !mask.parentNode) {
    const pos = getComputedStyle(el).position
    if (pos === 'static' || pos === '') el.classList.add('v-loading-parent')
    el.appendChild(mask)
  } else if (!value && mask.parentNode) {
    mask.parentNode.removeChild(mask)
    el.classList.remove('v-loading-parent')
  }
}

export const loadingDirective = {
  mounted(el, binding) {
    toggle(el, binding.value)
  },
  updated(el, binding) {
    if (binding.value !== binding.oldValue) toggle(el, binding.value)
  },
  unmounted(el) {
    toggle(el, false)
    el.__loadingMask = null
  }
}
