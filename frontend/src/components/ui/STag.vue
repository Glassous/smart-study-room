<script setup>
import { computed } from 'vue'

const props = defineProps({
  type: { type: String, default: '' },
  dot: { type: Boolean, default: false },
  line: { type: Boolean, default: false }
})

const typeClass = computed(() => {
  const map = {
    primary: 'pill-primary',
    success: 'pill-success',
    warning: 'pill-warning',
    danger: 'pill-danger',
    info: 'pill-info',
    violet: 'pill-violet',
    '': 'pill-neutral'
  }
  return map[props.type] || 'pill-neutral'
})
</script>

<template>
  <span class="pill" :class="[typeClass, { 'pill-plain': !dot, 'pill-line': line }]"><slot /></span>
</template>
