<script setup>
import { computed } from 'vue'

const props = defineProps({
  variant: { type: String, default: 'secondary' },
  size: { type: String, default: 'md' },
  type: { type: String, default: 'button' },
  loading: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  block: { type: Boolean, default: false }
})

const classes = computed(() => [
  `s-btn--${props.variant}`,
  `s-btn--${props.size}`,
  { 's-btn--block': props.block, 'is-loading': props.loading }
])

const isDisabled = computed(() => props.disabled || props.loading)
</script>

<template>
  <button
    :type="type"
    class="s-btn"
    :class="classes"
    :disabled="isDisabled"
    :aria-busy="loading || null"
  >
    <span v-if="loading" class="s-btn__spinner" aria-hidden="true" />
    <span class="s-btn__content"><slot /></span>
  </button>
</template>

<style scoped>
.s-btn {
  --btn-height: 34px;
  --btn-padding: 14px;
  --btn-font: 13.5px;

  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  height: var(--btn-height);
  padding: 0 var(--btn-padding);
  border-radius: var(--r-md);
  border: 1px solid transparent;
  font-size: var(--btn-font);
  font-weight: 500;
  line-height: 1;
  white-space: nowrap;
  user-select: none;
  transition:
    background var(--dur-1) var(--ease),
    border-color var(--dur-1) var(--ease),
    color var(--dur-1) var(--ease),
    box-shadow var(--dur-1) var(--ease),
    transform var(--dur-1) var(--ease);
}

.s-btn:active:not(:disabled) {
  transform: translateY(0.5px);
}

.s-btn:disabled {
  opacity: .55;
  pointer-events: none;
}

.s-btn__content {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-width: 0;
}

.s-btn__content :deep(.app-icon) {
  flex: 0 0 auto;
}

.s-btn__spinner {
  width: 14px;
  height: 14px;
  flex: 0 0 auto;
  border-radius: 50%;
  border: 2px solid currentColor;
  border-top-color: transparent;
  opacity: .9;
  animation: s-btn-spin .65s linear infinite;
}

@keyframes s-btn-spin {
  to { transform: rotate(360deg); }
}

/* ---------- 尺寸 ---------- */
.s-btn--sm { --btn-height: 28px; --btn-padding: 9px; --btn-font: 12.5px; border-radius: var(--r-sm); }
.s-btn--md { --btn-height: 34px; }
.s-btn--lg { --btn-height: 42px; --btn-padding: 18px; --btn-font: 14.5px; }

.s-btn--block {
  width: 100%;
}

/* ---------- 变体 ---------- */

/* 主按钮 */
.s-btn--primary {
  background: var(--primary);
  color: var(--text-inverse);
  box-shadow: var(--shadow-1);
}
.s-btn--primary:hover:not(:disabled) {
  background: var(--primary-hover);
  box-shadow: 0 2px 8px rgba(59, 102, 218, .28);
}
.s-btn--primary:active:not(:disabled) {
  background: var(--primary-active);
}

/* 次按钮 */
.s-btn--secondary {
  background: var(--surface);
  border-color: var(--border);
  color: var(--text-2);
}
.s-btn--secondary:hover:not(:disabled) {
  border-color: var(--border-strong);
  background: var(--surface-hover);
  color: var(--text-1);
}

/* 柔和按钮（表格行操作） */
.s-btn--soft {
  background: var(--surface-3);
  color: var(--text-2);
}
.s-btn--soft:hover:not(:disabled) {
  background: #e7ecf4;
  color: var(--text-1);
}

.s-btn--soft-success {
  background: var(--green-weak);
  color: var(--green-strong);
}
.s-btn--soft-success:hover:not(:disabled) {
  background: #d3eedf;
}

.s-btn--soft-danger {
  background: var(--red-weak);
  color: var(--red-strong);
}
.s-btn--soft-danger:hover:not(:disabled) {
  background: #f6dcda;
}

.s-btn--soft-warn {
  background: var(--amber-weak);
  color: var(--amber-strong);
}
.s-btn--soft-warn:hover:not(:disabled) {
  background: #f6e5c4;
}

/* 幽灵按钮 */
.s-btn--ghost {
  background: transparent;
  color: var(--text-3);
}
.s-btn--ghost:hover:not(:disabled) {
  background: var(--surface-3);
  color: var(--text-1);
}

/* 危险实心 */
.s-btn--danger {
  background: var(--red);
  color: var(--text-inverse);
}
.s-btn--danger:hover:not(:disabled) {
  background: var(--red-strong);
}
</style>
