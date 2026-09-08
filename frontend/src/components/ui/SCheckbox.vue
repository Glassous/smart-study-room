<script setup>
defineProps({
  modelValue: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false }
})

defineEmits(['update:modelValue'])
</script>

<template>
  <label class="s-check" :class="{ 'is-disabled': disabled }">
    <input
      class="s-check__native"
      type="checkbox"
      :checked="modelValue"
      :disabled="disabled"
      @change="$emit('update:modelValue', $event.target.checked)"
    />
    <span class="s-check__box" aria-hidden="true">
      <svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="3.2" stroke-linecap="round" stroke-linejoin="round"><path d="m5 12.5 4.2 4.2L19 7" /></svg>
    </span>
    <span class="s-check__label"><slot /></span>
  </label>
</template>

<style scoped>
.s-check {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.s-check.is-disabled {
  opacity: .55;
  pointer-events: none;
}

.s-check__native {
  position: absolute;
  opacity: 0;
  width: 0;
  height: 0;
}

.s-check__box {
  display: grid;
  place-items: center;
  width: 17px;
  height: 17px;
  border-radius: 5px;
  border: 1.5px solid var(--border-strong);
  background: var(--surface);
  color: transparent;
  flex: 0 0 auto;
  transition: background var(--dur-1) var(--ease), border-color var(--dur-1) var(--ease), color var(--dur-1) var(--ease);
}

.s-check:hover .s-check__box {
  border-color: var(--primary-hover);
}

.s-check__native:checked + .s-check__box {
  background: var(--primary);
  border-color: var(--primary);
  color: #fff;
}

.s-check__native:focus-visible + .s-check__box {
  box-shadow: var(--ring);
}

.s-check__label {
  font-size: var(--fs-body-sm);
  color: var(--text-2);
  line-height: 1.5;
}
</style>
