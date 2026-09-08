<script setup>
import { computed } from 'vue'

const props = defineProps({
  value: { type: Number, default: 0 },
  size: { type: Number, default: 152 },
  stroke: { type: Number, default: 11 },
  color: { type: String, default: 'var(--primary)' }
})

const SWEEP = 0.75 // 270° 圆弧

const radius = computed(() => (props.size - props.stroke) / 2)
const center = computed(() => props.size / 2)
const circumference = computed(() => 2 * Math.PI * radius.value)
const arcLen = computed(() => circumference.value * SWEEP)
const progress = computed(() => Math.max(0, Math.min(100, props.value)))
const dash = computed(() => `${arcLen.value * (progress.value / 100)} ${circumference.value}`)
const trackDash = computed(() => `${arcLen.value} ${circumference.value}`)
</script>

<template>
  <div class="s-gauge" :style="{ width: `${size}px`, height: `${size}px` }">
    <svg :width="size" :height="size" :viewBox="`0 0 ${size} ${size}`" aria-hidden="true">
      <g :transform="`rotate(135 ${center} ${center})`">
        <circle
          :cx="center" :cy="center" :r="radius"
          fill="none" stroke="var(--surface-3)" :stroke-width="stroke" stroke-linecap="round"
          :stroke-dasharray="trackDash"
        />
        <circle
          :cx="center" :cy="center" :r="radius"
          fill="none" :stroke="color" :stroke-width="stroke" stroke-linecap="round"
          :stroke-dasharray="dash"
          class="s-gauge__bar"
        />
      </g>
    </svg>
    <div class="s-gauge__center"><slot /></div>
  </div>
</template>

<style scoped>
.s-gauge {
  position: relative;
  flex: 0 0 auto;
}

.s-gauge__bar {
  transition: stroke-dasharray .6s var(--ease-out);
}

.s-gauge__center {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding-bottom: 4px;
}
</style>
