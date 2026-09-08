<script setup>
defineProps({
  name: { type: String, required: true },
  size: { type: [Number, String], default: 20 }
})
</script>

<template>
  <svg
    class="app-icon"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
  >
    <!-- 通用线性图标：无填充纯描边，颜色一律 currentColor / 语义 CSS 变量 -->
    <g v-if="name === 'home'">
      <path d="M4.2 10.2 12 3.7l7.8 6.5v9.3a1.5 1.5 0 0 1-1.5 1.5H5.7a1.5 1.5 0 0 1-1.5-1.5v-9.3Z" stroke="currentColor" stroke-width="1.7" />
      <path d="M9.3 20.9v-6h5.4v6" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'booking'">
      <rect x="3.2" y="4.5" width="17.6" height="15.5" rx="2.4" stroke="currentColor" stroke-width="1.7" />
      <path d="M7.9 2.4v4.2M16.1 2.4v4.2M3.3 9.7h17.4" stroke="currentColor" stroke-width="1.7" />
      <path d="m8.7 15.3 2.2 2.2 4.6-4.7" stroke="currentColor" stroke-width="1.9" />
    </g>
    <g v-else-if="name === 'reservations'">
      <path d="M14.6 2.9H6a2.05 2.05 0 0 0-2.05 2.05v14.1c0 1.13.92 2.05 2.05 2.05h12a2.05 2.05 0 0 0 2.05-2.05V7.2l-4.45-4.3Z" stroke="currentColor" stroke-width="1.7" />
      <path d="M14.6 2.9v4.4h4.5" stroke="currentColor" stroke-width="1.7" />
      <path d="M8.8 12.7h6.5M8.8 16.1h4.3" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'waitlist'">
      <circle cx="12" cy="12" r="8.6" stroke="currentColor" stroke-width="1.7" />
      <path d="M12 7.4V12l3.4 1.9" stroke="currentColor" stroke-width="1.8" />
    </g>
    <g v-else-if="name === 'notification' || name === 'notification-empty'">
      <path d="M18 8a6 6 0 1 0-12 0c0 7-3 9-3 9h18s-3-2-3-9" stroke="currentColor" stroke-width="1.7" />
      <path d="M13.8 21a2 2 0 0 1-3.6 0" stroke="currentColor" stroke-width="1.7" />
      <circle v-if="name === 'notification'" class="notif-dot" cx="15.7" cy="5.7" r="1.9" />
    </g>
    <g v-else-if="name === 'analytics'">
      <path d="M3.2 3.2v16a1.8 1.8 0 0 0 1.8 1.8h16" stroke="currentColor" stroke-width="1.7" />
      <path d="M7 16.4v-4.4M12 16.4V8M17 16.4v-6.8" stroke="currentColor" stroke-width="1.9" />
    </g>
    <g v-else-if="name === 'admin'">
      <path d="M12 3.4 19.4 6v5.2c0 4.5-2.9 7.9-7.4 9.6-4.5-1.7-7.4-5.1-7.4-9.6V6l7.4-2.6Z" stroke="currentColor" stroke-width="1.7" />
      <path d="m9.5 12.1 1.8 1.8 3.5-3.6" stroke="currentColor" stroke-width="1.9" />
    </g>
    <g v-else-if="name === 'profile'">
      <circle cx="12" cy="7.5" r="3.7" stroke="currentColor" stroke-width="1.7" />
      <path d="M5.1 20.4c.6-3.6 3.1-5.6 6.9-5.6s6.3 2 6.9 5.6" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'sidebar-collapse'">
      <rect x="2.75" y="2.75" width="18.5" height="18.5" rx="2.5" stroke="currentColor" stroke-width="1.6" />
      <path d="M9 2.75v18.5" stroke="currentColor" stroke-width="1.6" />
      <path d="m15.7 9.2-2.8 2.8 2.8 2.8" stroke="currentColor" stroke-width="1.9" />
    </g>
    <g v-else-if="name === 'sidebar-expand'">
      <rect x="2.75" y="2.75" width="18.5" height="18.5" rx="2.5" stroke="currentColor" stroke-width="1.6" />
      <path d="M9 2.75v18.5" stroke="currentColor" stroke-width="1.6" />
      <path d="m12.3 9.2 2.8 2.8-2.8 2.8" stroke="currentColor" stroke-width="1.9" />
    </g>
    <g v-else-if="name === 'more'">
      <circle cx="5.5" cy="12" r="1.8" fill="currentColor" />
      <circle cx="12" cy="12" r="1.8" fill="currentColor" />
      <circle cx="18.5" cy="12" r="1.8" fill="currentColor" />
    </g>
    <g v-else-if="name === 'refresh'">
      <path d="M3 12a9 9 0 0 1 15-6.7L21 8" stroke="currentColor" stroke-width="1.8" />
      <path d="M21 3v5h-5" stroke="currentColor" stroke-width="1.8" />
      <path d="M21 12a9 9 0 0 1-15 6.7L3 16" stroke="currentColor" stroke-width="1.8" />
      <path d="M3 21v-5h5" stroke="currentColor" stroke-width="1.8" />
    </g>
    <g v-else-if="name === 'sparkles'">
      <path d="m12 3.2 1.9 5.7 5.7 1.9-5.7 1.9-1.9 5.7-1.9-5.7-5.7-1.9 5.7-1.9 1.9-5.7Z" stroke="currentColor" stroke-width="1.6" />
      <path d="M18.8 3.6v3.2M17.2 5.2h3.2" stroke="currentColor" stroke-width="1.6" />
    </g>
    <g v-else-if="name === 'seat'">
      <path d="M19 9.5V6.3a2.3 2.3 0 0 0-2.3-2.3H7.3A2.3 2.3 0 0 0 5 6.3v3.2" stroke="currentColor" stroke-width="1.7" />
      <path d="M3.2 16.4a1.8 1.8 0 0 0 1.8 1.8h14a1.8 1.8 0 0 0 1.8-1.8v-4.9a2 2 0 0 0-4 0v1.6H7.2v-1.6a2 2 0 0 0-4 0v4.9Z" stroke="currentColor" stroke-width="1.7" />
      <path d="M5 18.2v2.4M19 18.2v2.4" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'desk-book'">
      <rect x="7.5" y="4.5" width="9" height="14" rx="1.2" stroke="currentColor" stroke-width="1.7" />
      <path d="M12 4.5v14" stroke="currentColor" stroke-width="1.5" />
    </g>
    <g v-else-if="name === 'desk-power'">
      <rect x="7.5" y="3.5" width="9" height="12" rx="2" stroke="currentColor" stroke-width="1.7" />
      <path d="M10.2 6.2v2.1M13.8 6.2v2.1" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
      <path d="M13.6 10.2 11.2 14h2l-.9 3 2.9-4.3h-2l.4-2.5Z" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" />
    </g>
    <g v-else-if="name === 'desk-pc'">
      <rect x="3.8" y="4" width="16.4" height="10.8" rx="1.7" stroke="currentColor" stroke-width="1.7" />
      <path d="M12 14.8v2.2M8.3 19.3h7.4" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" />
      <path d="M7.2 12.3l2.2-2.2 1.8 1.8 3.2-3.2" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" />
    </g>
    <g v-else-if="name === 'eye'">
      <path d="M2 12s3.7-7.4 10-7.4S22 12 22 12s-3.7 7.4-10 7.4S2 12 2 12Z" stroke="currentColor" stroke-width="1.6" />
      <circle cx="12" cy="12" r="2.8" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'inbox'">
      <path d="M22 12h-6l-2 3h-4l-2-3H2" stroke="currentColor" stroke-width="1.7" />
      <path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11Z" stroke="currentColor" stroke-width="1.6" />
    </g>
    <g v-else-if="name === 'circle-success'">
      <circle class="st-success" cx="12" cy="12" r="8.2" stroke-width="1.7" />
      <path class="st-success" d="m8.7 12.2 2.2 2.2 4.6-4.7" stroke-width="1.9" />
    </g>
    <g v-else-if="name === 'circle-wait'">
      <circle class="st-wait" cx="12" cy="12" r="8.2" stroke-width="1.7" />
      <path class="st-wait" d="M12 7.7V12l3.1 1.8" stroke-width="1.8" />
    </g>
    <g v-else-if="name === 'circle-muted'">
      <circle class="st-muted" cx="12" cy="12" r="8.2" stroke-width="1.7" />
      <path class="st-muted" d="M8.5 12h7" stroke-width="1.9" />
    </g>
    <g v-else-if="name === 'warning'">
      <path class="st-warn" d="M12 3.6 20.6 19a1 1 0 0 1-.87 1.5H4.27A1 1 0 0 1 3.4 19L12 3.6Z" stroke-width="1.7" />
      <path class="st-warn" d="M12 8.6v5M12 16.5v.2" stroke-width="2" />
    </g>
    <g v-else-if="name === 'credit'">
      <rect x="2.8" y="5" width="18.4" height="14" rx="2.2" stroke="currentColor" stroke-width="1.7" />
      <path d="M2.8 9.6h18.4" stroke="currentColor" stroke-width="1.7" />
      <path d="M6.8 15h4.4" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'ticket'">
      <path d="M3.6 6.5h16.8v3.3a2.3 2.3 0 0 0 0 4.4v3.3H3.6v-3.3a2.3 2.3 0 0 0 0-4.4V6.5Z" stroke="currentColor" stroke-width="1.7" />
      <path d="M12 8.2v1.4M12 11.3v1.4M12 14.4v1.4" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'announcement'">
      <path d="m3 11 18-5v12L3 14v-3Z" stroke="currentColor" stroke-width="1.7" />
      <path d="M11.6 16.8a3 3 0 1 1-5.8-1.6" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'building'">
      <path d="M6.2 21V4.4a2 2 0 0 1 2-2h7.6a2 2 0 0 1 2 2V21" stroke="currentColor" stroke-width="1.7" />
      <path d="M6.2 12.2H3.8v8.8M17.8 9h2.4v12" stroke="currentColor" stroke-width="1.7" />
      <path d="M10 6.5h4M10 9.5h4M10 12.5h4" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'users'">
      <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke="currentColor" stroke-width="1.7" />
      <circle cx="9" cy="7" r="4" stroke="currentColor" stroke-width="1.7" />
      <path d="M23 21v-2a4 4 0 0 0-3-3.87" stroke="currentColor" stroke-width="1.7" />
      <path d="M16 3.13a4 4 0 0 1 0 7.75" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'add'">
      <circle cx="12" cy="12" r="8.5" stroke="currentColor" stroke-width="1.7" />
      <path d="M12 8v8M8 12h8" stroke="currentColor" stroke-width="1.8" />
    </g>
    <g v-else-if="name === 'walk'">
      <circle cx="12" cy="5" r="1.9" stroke="currentColor" stroke-width="1.6" />
      <path d="m9.2 20.2 2.8-6 2.8 6" stroke="currentColor" stroke-width="1.7" />
      <path d="m6.4 8.4 5.6 1.8 5.6-1.8" stroke="currentColor" stroke-width="1.7" />
      <path d="M12 10.2v4" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'book'">
      <path d="M12 6.6C10 4.9 7.2 4.1 4.3 4.6c-1.1.2-1.9 1.1-1.9 2.2V19c0 1 .9 1.8 1.9 1.7 2.6-.4 5 .2 7.7 2.2 2.7-2 5.1-2.6 7.7-2.2 1 .1 1.9-.7 1.9-1.7V6.8c0-1.1-.8-2-1.9-2.2-2.9-.5-5.7.3-7.7 2Z" stroke="currentColor" stroke-width="1.6" />
      <path d="M12 6.6v16.3" stroke="currentColor" stroke-width="1.6" />
    </g>
    <g v-else-if="name === 'chevron-right'">
      <path d="m9 5 7 7-7 7" stroke="currentColor" stroke-width="2" />
    </g>
    <g v-else-if="name === 'check'">
      <path d="m5 12.5 4.2 4.2L19 7" stroke="currentColor" stroke-width="2.2" />
    </g>
    <g v-else-if="name === 'info'">
      <circle cx="12" cy="12" r="8.5" stroke="currentColor" stroke-width="1.7" />
      <path d="M12 10.6v5.1M12 7.4v.1" stroke="currentColor" stroke-width="2" />
    </g>
    <g v-else-if="name === 'sun'">
      <circle cx="12" cy="12" r="4.1" stroke="currentColor" stroke-width="1.7" />
      <path d="M12 2.2v2.3M12 19.5v2.3M4.5 4.5l1.6 1.6M17.9 17.9l1.6 1.6M2.2 12h2.3M19.5 12h2.3M4.5 19.5l1.6-1.6M17.9 6.1l1.6-1.6" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'moon'">
      <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'system'">
      <rect x="2.6" y="3.4" width="18.8" height="13.4" rx="2" stroke="currentColor" stroke-width="1.7" />
      <path d="M8.4 20.6h7.2M12 16.8v3.8" stroke="currentColor" stroke-width="1.7" />
    </g>
    <g v-else-if="name === 'logout'">
      <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" stroke="currentColor" stroke-width="1.8" />
      <path d="m16 17 5-5-5-5M21 12H9" stroke="currentColor" stroke-width="1.8" />
    </g>
  </svg>
</template>

<style scoped>
.app-icon {
  display: block;
  color: inherit;
  overflow: visible;
}
.app-icon * {
  stroke-linecap: round;
  stroke-linejoin: round;
}
/* 语义状态色：引用主题令牌，随深浅色自动切换 */
.app-icon .notif-dot {
  fill: var(--red);
  stroke: none;
}
.app-icon .st-success {
  stroke: var(--green-strong);
}
.app-icon .st-wait {
  stroke: var(--amber-strong);
}
.app-icon .st-muted {
  stroke: var(--text-4);
}
.app-icon .st-warn {
  stroke: var(--amber-strong);
}
</style>
