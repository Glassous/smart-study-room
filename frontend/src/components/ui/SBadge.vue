<script setup>
import { computed } from 'vue'

const props = defineProps({
  value: { type: [Number, String], default: '' },
  max: { type: Number, default: 99 },
  hidden: { type: Boolean, default: false },
  isDot: { type: Boolean, default: false }
})

const display = computed(() => {
  if (props.isDot) return ''
  const n = Number(props.value)
  if (Number.isFinite(n) && n > props.max) return `${props.max}+`
  return String(props.value)
})

const isHidden = computed(() => props.hidden || (!props.isDot && (props.value === 0 || props.value === '' || props.value === undefined || props.value === null)))
</script>

<template>
  <span v-if="!isHidden" class="s-badge" :class="{ 'is-dot': isDot }">{{ display }}</span>
</template>

<style scoped>
.s-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: var(--r-full);
  background: var(--red);
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
  font-variant-numeric: tabular-nums;
  box-shadow: 0 0 0 2px var(--surface);
  pointer-events: none;
}

.s-badge.is-dot {
  min-width: 8px;
  width: 8px;
  height: 8px;
  padding: 0;
}
</style>
