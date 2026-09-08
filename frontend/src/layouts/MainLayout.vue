<script setup>
import { computed, ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { gsap } from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'
import { message } from '../components/ui/feedback'
import { useAuthStore } from '../stores/auth'
import { useNotificationStore } from '../stores/notification'
import { logout as logoutApi } from '../api/auth'
import AppIcon from '../components/AppIcon.vue'
import MobileTopbar from '../components/MobileTopbar.vue'
import AIAssistant from '../components/AIAssistant.vue'
import SBadge from '../components/ui/SBadge.vue'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const notifStore = useNotificationStore()
const narrowMediaQuery = window.matchMedia('(max-width: 1024px)')
const isNarrowScreen = ref(narrowMediaQuery.matches)
const isSidebarCollapsed = ref(narrowMediaQuery.matches)
const isAtTop = ref(true)
const mobileTitleProgress = ref(0)
const mobilePageTitle = ref('')
const contentRef = ref(null)
const isDesktopCollapsed = computed(() => !isNarrowScreen.value && isSidebarCollapsed.value)
let titleMedia = null

// 用户菜单弹出层（Teleport 到 body，避免被侧栏 overflow 裁剪）
const userMenuOpen = ref(false)
const userAreaRef = ref(null)
const userMenuStyle = ref({})

function updateUserMenuPosition() {
  const el = userAreaRef.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const collapsed = isDesktopCollapsed.value
  if (collapsed) {
    userMenuStyle.value = {
      position: 'fixed',
      left: `${Math.round(rect.right + 10)}px`,
      bottom: `${Math.round(window.innerHeight - rect.bottom)}px`,
      width: '150px'
    }
  } else {
    userMenuStyle.value = {
      position: 'fixed',
      left: `${Math.round(rect.left)}px`,
      bottom: `${Math.round(window.innerHeight - rect.top + 8)}px`,
      width: `${Math.round(rect.width)}px`
    }
  }
}

function toggleUserMenu() {
  userMenuOpen.value = !userMenuOpen.value
  if (userMenuOpen.value) updateUserMenuPosition()
}

function onDocPointerDown(e) {
  if (userMenuOpen.value && !userAreaRef.value?.contains(e.target) && !e.target.closest('.user-menu')) {
    userMenuOpen.value = false
  }
}

function onUiReposition() {
  if (userMenuOpen.value) updateUserMenuPosition()
}

// 未读消息角标(通过 Pinia Store 集中管理与 60s 轮询)
const unread = computed(() => notifStore.unread)
let timer = null

function refreshUnread() {
  notifStore.refresh()
}

// 菜单：首页 / 预约 / 我的预约 / 候补 / 统计 / 消息 / 个人中心 / 管理端
const menus = computed(() => {
  const common = [
    { index: '/', title: '首页', icon: 'home' }
  ]
  const roleMenus = auth.isAdmin
    ? [
        { index: '/analytics', title: '热力图统计', icon: 'analytics' },
        { index: '/admin', title: '管理端', icon: 'admin' }
      ]
    : [
        { index: '/booking', title: '座位预约', icon: 'booking' },
        { index: '/mine', title: '我的预约', icon: 'reservations' },
        { index: '/waitlist', title: '我的候补', icon: 'waitlist' },
        { index: '/notifications', title: '消息中心', icon: unread.value > 0 ? 'notification' : 'notification-empty' }
      ]
  return [...common, ...roleMenus, { index: '/profile', title: '个人中心', icon: 'profile' }]
})

async function handleSelect(index) {
  userMenuOpen.value = false
  if (route.path !== index) await router.push(index)
  if (isNarrowScreen.value) closeSidebar()
}

function onLogout() {
  const token = auth.token
  if (token) void logoutApi(token).catch(() => {})
  auth.logout()
  message.success('已退出登录')
  router.push('/login')
}

function handleAccountCommand(command) {
  userMenuOpen.value = false
  if (command === 'profile') {
    router.push('/profile')
    if (isNarrowScreen.value) closeSidebar()
  }
  if (command === 'logout') onLogout()
}

function openSidebar() {
  isSidebarCollapsed.value = false
}

function closeSidebar() {
  isSidebarCollapsed.value = true
}

function toggleSidebar() {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

function onNarrowChange(event) {
  isNarrowScreen.value = event.matches
  isSidebarCollapsed.value = event.matches
  isAtTop.value = event.matches ? (contentRef.value?.scrollTop || 0) <= 1 : true
  nextTick(setupMobileTitleAnimation)
}

function onContentScroll(event) {
  if (!isNarrowScreen.value) return
  isAtTop.value = event.currentTarget.scrollTop <= 1
}

function onKeydown(event) {
  if (event.key === 'Escape' && isNarrowScreen.value && !isSidebarCollapsed.value) {
    closeSidebar()
  }
}

function cleanupMobileTitleAnimation() {
  titleMedia?.revert()
  titleMedia = null
  mobileTitleProgress.value = 0
}

function setupMobileTitleAnimation() {
  cleanupMobileTitleAnimation()
  const scroller = contentRef.value
  const heading = scroller?.querySelector('[data-page-title]')
  if (!scroller || !heading) return
  mobilePageTitle.value = heading.querySelector('h1')?.textContent?.trim() || heading.textContent?.trim() || route.meta?.title || ''

  titleMedia = gsap.matchMedia()
  titleMedia.add(
    {
      isNarrow: '(max-width: 1024px)',
      reduceMotion: '(prefers-reduced-motion: reduce)'
    },
    (context) => {
      const { isNarrow, reduceMotion } = context.conditions
      if (!isNarrow) return

      if (reduceMotion) {
        const trigger = ScrollTrigger.create({
          trigger: heading,
          scroller,
          start: 'top 52px',
          onEnter: () => {
            gsap.set(heading, { autoAlpha: 0, y: 0 })
            mobileTitleProgress.value = 1
          },
          onLeaveBack: () => {
            gsap.set(heading, { autoAlpha: 1, y: 0 })
            mobileTitleProgress.value = 0
          }
        })
        return () => trigger.kill()
      }

      const tween = gsap.fromTo(
        heading,
        { autoAlpha: 1, y: 0 },
        {
          autoAlpha: 0,
          y: -12,
          ease: 'none',
          scrollTrigger: {
            trigger: heading,
            scroller,
            start: 'top 72px',
            end: 'top 32px',
            scrub: 0.18,
            invalidateOnRefresh: true,
            onUpdate: (self) => {
              mobileTitleProgress.value = self.progress
            }
          }
        }
      )
      return () => tween.kill()
    },
    scroller
  )
  ScrollTrigger.refresh()
}

watch(() => route.fullPath, async () => {
  cleanupMobileTitleAnimation()
  mobilePageTitle.value = ''
  userMenuOpen.value = false
  if (contentRef.value) contentRef.value.scrollTop = 0
  await nextTick()
  setupMobileTitleAnimation()
})

onMounted(() => {
  narrowMediaQuery.addEventListener('change', onNarrowChange)
  window.addEventListener('keydown', onKeydown)
  window.addEventListener('scroll', onUiReposition, true)
  window.addEventListener('resize', onUiReposition)
  document.addEventListener('pointerdown', onDocPointerDown)
  if (auth.isStudent) {
    refreshUnread()
    timer = setInterval(refreshUnread, 60000)
  }
  nextTick(setupMobileTitleAnimation)
})
onUnmounted(() => {
  cleanupMobileTitleAnimation()
  narrowMediaQuery.removeEventListener('change', onNarrowChange)
  window.removeEventListener('keydown', onKeydown)
  window.removeEventListener('scroll', onUiReposition, true)
  window.removeEventListener('resize', onUiReposition)
  document.removeEventListener('pointerdown', onDocPointerDown)
  if (timer) clearInterval(timer)
})
</script>

<template>
  <!-- 第一层双栏：外壳 = 左 Sidebar | 右 Content Column -->
  <div
    class="shell"
    :class="{
      'sidebar-collapsed': isSidebarCollapsed,
      'narrow-screen': isNarrowScreen,
      'mobile-sidebar-open': isNarrowScreen && !isSidebarCollapsed
    }"
  >
    <MobileTopbar
      v-if="isNarrowScreen"
      :sidebar-collapsed="isSidebarCollapsed"
      :at-top="isAtTop"
      :menus="menus"
      :current-path="route.path"
      :page-title="mobilePageTitle"
      :title-progress="mobileTitleProgress"
      @open-sidebar="openSidebar"
      @navigate="handleSelect"
    />

    <div
      v-if="isNarrowScreen && !isSidebarCollapsed"
      class="sidebar-backdrop"
      aria-hidden="true"
      @click="closeSidebar"
      @wheel.prevent
      @touchmove.prevent
    />

    <!-- ============== 左栏（导航 / 品牌 / 用户） ============== -->
    <aside class="sidebar" :class="{ collapsed: isSidebarCollapsed }">
      <!-- 品牌（折叠/展开按钮固定在品牌行右侧，位置不随折叠状态改变） -->
      <div class="brand">
        <div class="brand-mark" aria-hidden="true">
          <AppIcon name="book" :size="18" />
        </div>
        <div v-show="!isDesktopCollapsed" class="brand-copy">
          <div class="brand-name">智能自习室</div>
          <div class="brand-sub">Smart Study Room</div>
        </div>
        <button
          class="sidebar-toggle"
          type="button"
          :title="isDesktopCollapsed ? '展开侧栏' : '折叠侧栏'"
          :aria-label="isDesktopCollapsed ? '展开侧栏' : '折叠侧栏'"
          :aria-expanded="!isSidebarCollapsed"
          @click="toggleSidebar"
        >
          <AppIcon :name="isDesktopCollapsed ? 'sidebar-expand' : 'sidebar-collapse'" :size="18" />
        </button>
      </div>

      <!-- 菜单 -->
      <nav class="menu" aria-label="主导航">
        <button
          v-for="m in menus"
          :key="m.index"
          class="menu-item"
          :class="{ active: route.path === m.index }"
          :title="isDesktopCollapsed ? m.title : undefined"
          :aria-label="m.title"
          @click="handleSelect(m.index)"
        >
          <span class="menu-icon"><AppIcon :name="m.icon" :size="19" /></span>
          <span class="menu-text">{{ m.title }}</span>
          <SBadge
            v-if="m.index === '/notifications' && unread > 0"
            :value="unread"
            :max="99"
            class="menu-badge"
          />
        </button>
      </nav>

      <div class="sidebar-spacer" />

      <!-- 底部：账户入口 -->
      <div v-if="auth.isLoggedIn" ref="userAreaRef" class="sidebar-footer">
        <button
          class="user-card"
          :class="{ open: userMenuOpen }"
          :title="isDesktopCollapsed ? (auth.user?.real_name || auth.user?.username) : undefined"
          :aria-expanded="userMenuOpen"
          @click="toggleUserMenu"
        >
          <div class="avatar-slot">
            <div class="avatar">
              {{ (auth.user?.real_name || auth.user?.username || '?').slice(0, 1) }}
            </div>
          </div>
          <div v-show="!isDesktopCollapsed" class="user-meta">
            <div class="user-name">
              {{ auth.user?.real_name || auth.user?.username }}
            </div>
            <span class="role-pill" :class="auth.isAdmin ? 'role-pill--admin' : 'role-pill--student'">
              {{ auth.isAdmin ? '管理员' : '学生' }}
            </span>
          </div>
          <span v-show="!isDesktopCollapsed" class="user-chev" :class="{ open: userMenuOpen }" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m8 10 4 4 4-4" /></svg>
          </span>
        </button>
      </div>
    </aside>

    <!-- ============== 右栏（页面正文；标题由各页面自己维护） ============== -->
    <section ref="contentRef" class="content" @scroll="onContentScroll">
      <!-- 每个页面自行使用 .split 实现"页面级双栏" -->
      <main class="page-body">
        <router-view />
      </main>
    </section>
    <AIAssistant v-if="auth.isLoggedIn && auth.isStudent" />

    <!-- 用户弹出菜单（Teleport 避免侧栏 overflow 裁剪） -->
    <Teleport to="body">
      <Transition name="user-menu">
        <div v-if="userMenuOpen && auth.isLoggedIn" class="user-menu" :style="userMenuStyle" role="menu">
          <button class="user-menu-item" role="menuitem" @click="handleAccountCommand('profile')">
            <AppIcon name="profile" :size="16" />个人中心
          </button>
          <div class="user-menu-divider" />
          <button class="user-menu-item user-menu-item--danger" role="menuitem" @click="handleAccountCommand('logout')">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M15 4h3.5A1.5 1.5 0 0 1 20 5.5v13a1.5 1.5 0 0 1-1.5 1.5H15M10 8l-4 4 4 4M6 12h10" transform="rotate(180 12 12)" /></svg>
            退出登录
          </button>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
/* ---- 外壳级双栏（全局） ---- */
.shell {
  height: 100%;
  display: grid;
  grid-template-columns: 256px 1fr;
  background: var(--canvas);
  transition: grid-template-columns var(--dur-3) var(--ease);
}
.shell:not(.narrow-screen).sidebar-collapsed {
  grid-template-columns: 72px 1fr;
}
.sidebar-backdrop {
  position: fixed;
  z-index: 40;
  inset: 0;
  background: rgba(23, 32, 54, .38);
  backdrop-filter: blur(1px);
}

/* ============== 左栏 Sidebar ============== */
.sidebar {
  display: flex;
  flex-direction: column;
  background: var(--surface);
  border-right: 1px solid var(--border);
  padding: 18px 12px 12px;
  min-height: 0;
  overflow: hidden;
}

/* 折叠态：收窄内边距以容纳 logo + 固定位置的折叠按钮 */
.shell:not(.narrow-screen) .sidebar.collapsed {
  padding: 18px 4px 12px;
}

.brand {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 2px 6px 14px 8px;
  min-height: 50px;
  white-space: nowrap;
}
.shell:not(.narrow-screen) .sidebar.collapsed .brand {
  padding: 2px 4px 14px;
  gap: 0;
}
.brand-mark {
  width: 32px;
  height: 32px;
  flex: 0 0 32px;
  border-radius: var(--r-lg);
  background: var(--primary);
  color: #fff;
  display: grid;
  place-items: center;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, .18), 0 2px 6px rgba(59, 102, 218, .3);
}
.brand-copy {
  min-width: 0;
  flex: 1;
}
.brand-name {
  font-size: 15px;
  font-weight: 650;
  color: var(--text-1);
  letter-spacing: -.01em;
}
.brand-sub {
  font-size: 10px;
  color: var(--text-4);
  margin-top: 1px;
  letter-spacing: .07em;
  text-transform: uppercase;
}
.sidebar-toggle {
  appearance: none;
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-4);
  width: 30px;
  height: 30px;
  padding: 0;
  border-radius: var(--r-md);
  display: grid;
  place-items: center;
  flex: 0 0 30px;
  transition: color var(--dur-1) var(--ease), background var(--dur-1) var(--ease);
}
.sidebar-toggle:hover {
  color: var(--text-2);
  background: var(--surface-3);
}
.sidebar-toggle:focus-visible {
  outline: 2px solid var(--primary);
  outline-offset: 1px;
}

/* 菜单 */
.menu {
  display: flex;
  flex-direction: column;
  gap: 2px;
  overflow-x: hidden;
  overflow-y: auto;
  padding: 4px 0;
  scrollbar-width: none;
}
.menu::-webkit-scrollbar {
  display: none;
}
.menu-item {
  all: unset;
  display: flex;
  align-items: center;
  gap: 0;
  padding: 0;
  border-radius: var(--r-md);
  cursor: pointer;
  color: var(--text-2);
  font-size: var(--fs-body-sm);
  font-weight: 500;
  transition: background var(--dur-1) var(--ease), color var(--dur-1) var(--ease);
  position: relative;
  width: 100%;
  min-height: 38px;
  box-sizing: border-box;
}
.menu-item:hover {
  background: var(--surface-hover);
  color: var(--text-1);
}
.menu-item.active {
  background: var(--primary-weak);
  color: var(--primary-active);
  font-weight: 600;
}
.menu-item.active::before {
  content: '';
  position: absolute;
  left: -12px;
  top: 8px;
  bottom: 8px;
  width: 3px;
  border-radius: 0 3px 3px 0;
  background: var(--primary);
}
.shell:not(.narrow-screen) .sidebar.collapsed .menu-item.active::before {
  left: -4px;
}
.menu-icon {
  width: 42px;
  height: 20px;
  display: grid;
  place-items: center;
  flex: 0 0 42px;
  color: var(--text-4);
  transition: color var(--dur-1) var(--ease);
}
.menu-item:hover .menu-icon {
  color: var(--text-2);
}
.menu-item.active .menu-icon {
  color: var(--primary);
}
.menu-text {
  flex: 1;
  min-width: 0;
  height: 20px;
  display: flex;
  align-items: center;
  line-height: 20px;
  white-space: nowrap;
  overflow: hidden;
}
.menu-badge {
  margin-right: 8px;
}
.shell:not(.narrow-screen) .sidebar.collapsed .menu-item {
  justify-content: center;
}
.shell:not(.narrow-screen) .sidebar.collapsed .menu-icon {
  flex: 0 0 auto;
  width: auto;
}
.shell:not(.narrow-screen) .sidebar.collapsed .menu-text {
  display: none;
}
/* 折叠态徽标锚定在图标右上角，避免孤悬 */
.shell:not(.narrow-screen) .sidebar.collapsed .menu-badge {
  position: absolute;
  top: 1px;
  left: calc(50% + 8px);
  right: auto;
  margin: 0;
  transform: scale(.86);
  transform-origin: top left;
}

.sidebar-spacer {
  flex: 1;
  min-height: 8px;
}

/* 底部用户卡 */
.sidebar-footer {
  padding-top: 10px;
}

.user-card {
  all: unset;
  display: flex;
  align-items: center;
  gap: 0;
  width: 100%;
  min-height: 56px;
  padding: 8px 0;
  border-radius: var(--r-lg);
  background: transparent;
  cursor: pointer;
  box-sizing: border-box;
  transition: background var(--dur-1) var(--ease);
}
.user-card:hover,
.user-card.open {
  background: var(--surface-hover);
}
.user-card:focus-visible {
  outline: 2px solid var(--primary);
  outline-offset: -2px;
}
.avatar-slot {
  width: 44px;
  flex: 0 0 44px;
  display: grid;
  place-items: center;
}
.avatar {
  width: 34px;
  height: 34px;
  border-radius: var(--r-md);
  background: var(--primary-weak);
  color: var(--primary-active);
  display: grid;
  place-items: center;
  font-weight: 700;
  font-size: 14px;
}
.shell:not(.narrow-screen) .sidebar.collapsed .avatar-slot {
  width: 100%;
  flex: 0 0 auto;
}
.user-meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 3px;
  padding-left: 2px;
}
.user-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-1);
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.role-pill {
  font-size: 11px;
  font-weight: 500;
  line-height: 1;
  padding: 3px 8px;
  border-radius: var(--r-full);
}
.role-pill--student {
  background: var(--primary-weak);
  color: var(--primary-active);
}
.role-pill--admin {
  background: var(--red-weak);
  color: var(--red-strong);
}
.user-chev {
  color: var(--text-4);
  flex: 0 0 auto;
  margin-right: 10px;
  display: grid;
  place-items: center;
  transition: transform var(--dur-2) var(--ease);
}
.user-chev.open {
  transform: rotate(180deg);
}

/* 用户弹出菜单（Teleport 到 body，fixed 定位由 JS 计算） */
.user-menu {
  z-index: calc(var(--z-sidebar) + 10);
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--r-lg);
  box-shadow: var(--shadow-3);
  padding: 5px;
  box-sizing: border-box;
}
.user-menu-item {
  all: unset;
  display: flex;
  align-items: center;
  gap: 9px;
  width: 100%;
  padding: 8px 10px;
  border-radius: var(--r-sm);
  font-size: var(--fs-body-sm);
  color: var(--text-2);
  cursor: pointer;
  box-sizing: border-box;
  transition: background var(--dur-1) var(--ease), color var(--dur-1) var(--ease);
}
.user-menu-item:hover {
  background: var(--surface-hover);
  color: var(--text-1);
}
.user-menu-item--danger {
  color: var(--red-strong);
}
.user-menu-item--danger:hover {
  background: var(--red-weak);
  color: var(--red-strong);
}
.user-menu-divider {
  height: 1px;
  background: var(--hairline);
  margin: 4px 6px;
}
.user-menu-enter-active,
.user-menu-leave-active {
  transition: opacity var(--dur-1) var(--ease), transform var(--dur-2) var(--ease-out);
}
.user-menu-enter-from,
.user-menu-leave-to {
  opacity: 0;
  transform: translateY(6px);
}

/* ============== 右栏 Content ============== */
.content {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
}
.page-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

/* ============== 窄屏：透明 Topbar + 覆盖式侧栏 ============== */
@media (max-width: 1024px) {
  .shell,
  .shell.sidebar-collapsed {
    display: block;
    position: relative;
    height: 100%;
  }

  .sidebar {
    position: fixed;
    z-index: 50;
    inset: 0 auto 0 0;
    width: 272px;
    min-height: 100%;
    transform: translateX(0);
    visibility: visible;
    pointer-events: auto;
    box-shadow: 12px 0 32px rgba(23, 32, 54, .18);
    transition: transform var(--dur-3) var(--ease), visibility 0s linear 0s;
  }
  .sidebar.collapsed {
    transform: translateX(-100%);
    visibility: hidden;
    pointer-events: none;
    box-shadow: none;
    transition: transform var(--dur-3) var(--ease), visibility 0s linear .24s;
  }

  .content {
    display: block;
    width: 100%;
    height: 100%;
    padding: 0;
    overflow-x: hidden;
    overflow-y: auto;
    overscroll-behavior: contain;
  }
  .mobile-sidebar-open .content {
    overflow: hidden;
  }
  .page-body {
    min-height: 0;
    overflow: visible;
  }
  /* 触控目标：移动端菜单项加高 */
  .menu-item {
    min-height: 46px;
  }
}
</style>
