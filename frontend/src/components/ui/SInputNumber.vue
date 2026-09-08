<script setup>
import { computed } from 'vue'

const props = defineProps({
  modelValue: { type: Number, default: 0 },
  min: { type: Number, default: -Infinity },
  max: { type: Number, default: Infinity },
  step: { type: Number, default: 1 },
  disabled: { type: Boolean, default: false }
})

const emit = defineEmits(['update:modelValue', 'change'])

const display = computed(() => String(props.modelValue))

function clamp(v) {
  return Math.min(props.max, Math.max(props.min, v))
}

function commit(e) {
  const parsed = Number(e.target.value)
  if (!Number.isFinite(parsed)) {
    e.target.value = display.value
    return
  }
  const next = clamp(Math.trunc(parsed))
  e.target.value = String(next)
  if (next !== props.modelValue) {
    emit('update:modelValue', next)
    emit('change', next)
  }
}

function stepBy(delta) {
  if (props.disabled) return
  const base = Number.isFinite(props.modelValue) ? props.modelValue : props.min
  const next = clamp(base + delta)
  if (next !== props.modelValue) {
    emit('update:modelValue', next)
    emit('change', next)
  }
}

function onKeydown(e) {
  if (e.key === 'ArrowUp') { e.preventDefault(); stepBy(props.step) }
  else if (e.key === 'ArrowDown') { e.preventDefault(); stepBy(-props.step) }
}
</script>

<template>
  <div class="s-number" :class="{ 'is-disabled': disabled }">
    <input
      class="s-number__field"
      type="text"
      inputmode="numeric"
      :value="display"
      :disabled="disabled"
      @change="commit"
      @keydown="onKeydown"
    />
    <span class="s-number__controls">
      <button type="button" class="s-number__btn" :disabled="disabled || modelValue >= max" aria-label="增加" @click="stepBy(step)">
        <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="m6 14 6-6 6 6" /></svg>
      </button>
      <button type="button" class="s-number__btn" :disabled="disabled || modelValue <= min" aria-label="减少" @click="stepBy(-step)">
        <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="m6 10 6 6 6-6" /></svg>
      </button>
    </span>
  </div>
</template>

<style scoped>
.s-number {
  display: inline-flex;
  align-items: stretch;
  width: 116px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-md);
  overflow: hidden;
  transition: border-color var(--dur-1) var(--ease), box-shadow var(--dur-1) var(--ease);
}

.s-number:focus-within {
  border-color: var(--primary);
  box-shadow: var(--ring);
}

.s-number.is-disabled {
  background: var(--surface-3);
  opacity: .7;
  pointer-events: none;
}

.s-number__field {
  flex: 1;
  min-width: 0;
  border: 0;
  outline: none;
  background: transparent;
  padding: 0 10px;
  height: 32px;
  font-size: var(--fs-body-sm);
  color: var(--text-1);
  font-variant-numeric: tabular-nums;
}

.s-number__controls {
  display: flex;
  flex-direction: column;
  border-left: 1px solid var(--hairline);
  background: var(--surface-2);
}

.s-number__btn {
  all: unset;
  display: grid;
  place-items: center;
  flex: 1;
  min-width: 24px;
  color: var(--text-3);
  cursor: pointer;
  transition: background var(--dur-1) var(--ease), color var(--dur-1) var(--ease);
}

.s-number__btn:hover:not(:disabled) {
  background: var(--surface-hover);
  color: var(--text-1);
}

.s-number__btn:disabled {
  opacity: .35;
  cursor: not-allowed;
}

.s-number__btn + .s-number__btn {
  border-top: 1px solid var(--hairline);
}
</style>
