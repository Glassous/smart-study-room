<script setup>
import AppIcon from '../AppIcon.vue'

defineProps({
  description: { type: String, default: '暂无数据' },
  icon: { type: String, default: 'seat' }
})
</script>

<template>
  <div class="empty-state">
    <div class="empty-figure" aria-hidden="true">
      <AppIcon :name="icon" :size="40" />
    </div>
    <div class="empty-title">{{ description }}</div>
    <div v-if="$slots.default" class="empty-desc"><slot /></div>
  </div>
</template>

<style scoped>
.empty-figure {
  display: grid;
  place-items: center;
  width: 72px;
  height: 72px;
  border-radius: var(--r-2xl);
  background: var(--surface-3);
  color: var(--text-4);
  margin-bottom: 4px;
}
</style>
