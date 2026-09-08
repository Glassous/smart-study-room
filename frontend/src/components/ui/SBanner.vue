<script setup>
import { computed } from 'vue'

const props = defineProps({
  type: { type: String, default: 'info' },
  title: { type: String, default: '' }
})

const icon = computed(() => {
  switch (props.type) {
    case 'success':
      return 'M8.7 12.1l2.1 2.1 4.7-4.8'
    case 'warning':
      return 'M12 8.2v5.5M12 16.7v.1'
    case 'error':
      return 'M9 9l6 6M15 9l-6 6'
    default:
      return 'M12 10.5v5M12 7.4v.1'
  }
})
</script>

<template>
  <div class="s-banner" :class="`s-banner--${type}`" role="status">
    <svg class="s-banner__icon" viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
      <circle v-if="type === 'info'" cx="12" cy="12" r="8.5" :stroke-width="1.7" />
      <path v-else-if="type === 'warning'" d="M12 3.5 21 19H3L12 3.5Z" :stroke-width="1.7" />
      <path v-else-if="type === 'success'" d="M4.5 12.5a7.5 7.5 0 1 0 15 0 7.5 7.5 0 1 0-15 0Z" :stroke-width="1.7" />
      <path v-else d="M4.5 12.5a7.5 7.5 0 1 0 15 0 7.5 7.5 0 1 0-15 0Z" :stroke-width="1.7" />
      <path :d="icon" />
    </svg>
    <div class="s-banner__body">
      <div v-if="title" class="s-banner__title">{{ title }}</div>
      <div v-if="$slots.default" class="s-banner__content"><slot /></div>
    </div>
  </div>
</template>

<style scoped>
.s-banner {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 11px 14px;
  border-radius: var(--r-lg);
  border: 1px solid transparent;
  font-size: var(--fs-body-sm);
  line-height: 1.6;
}

.s-banner__icon {
  flex: 0 0 auto;
  margin-top: 3px;
}

.s-banner__title {
  color: var(--text-1);
  font-weight: 550;
}

.s-banner__content {
  color: var(--text-2);
}

.s-banner--info {
  background: var(--primary-faint);
  border-color: #dfe8fb;
  color: var(--primary-active);
}

.s-banner--info .s-banner__title {
  color: var(--primary-active);
}

.s-banner--success {
  background: var(--green-faint);
  border-color: #d5ecdf;
  color: var(--green-strong);
}

.s-banner--success .s-banner__title {
  color: var(--green-strong);
}

.s-banner--warning {
  background: var(--amber-faint);
  border-color: #f2e3c2;
  color: var(--amber-strong);
}

.s-banner--warning .s-banner__title {
  color: var(--amber-strong);
}

.s-banner--error {
  background: var(--red-faint);
  border-color: #f2d6d3;
  color: var(--red-strong);
}

.s-banner--error .s-banner__title {
  color: var(--red-strong);
}
</style>
