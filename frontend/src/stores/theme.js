import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const savedMode = typeof localStorage !== 'undefined' ? localStorage.getItem('studyroom-theme-mode') : null
  const mode = ref(savedMode === 'light' || savedMode === 'dark' || savedMode === 'system' ? savedMode : 'system')

  const systemIsDark = ref(
    typeof window !== 'undefined' && window.matchMedia
      ? window.matchMedia('(prefers-color-scheme: dark)').matches
      : false
  )

  const resolvedTheme = computed(() => {
    if (mode.value === 'system') {
      return systemIsDark.value ? 'dark' : 'light'
    }
    return mode.value
  })

  const isDark = computed(() => resolvedTheme.value === 'dark')

  function applyTheme() {
    if (typeof document === 'undefined') return
    const theme = resolvedTheme.value
    document.documentElement.setAttribute('data-theme', theme)
    if (theme === 'dark') {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  function setMode(newMode) {
    if (['light', 'dark', 'system'].includes(newMode)) {
      mode.value = newMode
      if (typeof localStorage !== 'undefined') {
        localStorage.setItem('studyroom-theme-mode', newMode)
      }
      applyTheme()
    }
  }

  let mediaQueryListenerAttached = false

  function initTheme() {
    if (typeof window !== 'undefined' && window.matchMedia) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      systemIsDark.value = mediaQuery.matches

      if (!mediaQueryListenerAttached) {
        mediaQuery.addEventListener('change', (e) => {
          systemIsDark.value = e.matches
          if (mode.value === 'system') {
            applyTheme()
          }
        })
        mediaQueryListenerAttached = true
      }
    }
    applyTheme()
  }

  return {
    mode,
    systemIsDark,
    resolvedTheme,
    isDark,
    setMode,
    initTheme,
    applyTheme
  }
})
