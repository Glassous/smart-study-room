<script setup>
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'
import { unreadCount } from '../api/notification'

const router = useRouter()
const auth = useAuthStore()

// 未读消息角标(60s 轮询)
const unread = ref(0)
let timer = null

async function refreshUnread() {
  if (!auth.isLoggedIn) return
  try {
    const resp = await unreadCount()
    unread.value = resp.data?.count || 0
  } catch { /* 忽略轮询失败 */ }
}

// 菜单随迭代补充: 预约/我的预约/热力图/消息/个人中心/管理端
const menus = computed(() => {
  const items = [
    { index: '/booking', title: '座位预约', icon: 'Seat' },
    { index: '/mine', title: '我的预约', icon: 'Tickets' },
    { index: '/waitlist', title: '我的候补', icon: 'Clock' },
    { index: '/analytics', title: '热力图统计', icon: 'DataAnalysis' },
    { index: '/notifications', title: '消息中心', icon: 'Bell' },
    { index: '/profile', title: '个人中心', icon: 'User' },
    { index: '/', title: '首页', icon: 'HomeFilled' }
  ]
  if (auth.isAdmin) {
    items.push({ index: '/admin', title: '管理端', icon: 'Setting' })
  }
  return items
})

function handleSelect(index) {
  router.push(index)
}

function onLogout() {
  auth.logout()
  ElMessage.success('已退出登录')
  router.push('/login')
}

onMounted(() => {
  refreshUnread()
  timer = setInterval(refreshUnread, 60000)
})
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <el-container class="layout">
    <el-aside width="220px" class="aside">
      <div class="logo">📚 智能自习室</div>
      <el-menu :default-active="$route.path" router @select="handleSelect">
        <el-menu-item v-for="m in menus" :key="m.index" :index="m.index">
          <span>{{ m.title }}</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <div class="header-title">智能共享自习室预约系统</div>
        <div class="header-right">
          <el-badge :value="unread" :hidden="unread === 0" :max="99" class="bell-badge">
            <el-button text @click="router.push('/notifications')">🔔</el-button>
          </el-badge>
          <el-dropdown v-if="auth.isLoggedIn" @command="(cmd) => cmd === 'logout' && onLogout()">
            <span class="user-chip">
              {{ auth.user?.real_name || auth.user?.username }}
              <el-tag size="small" :type="auth.isAdmin ? 'danger' : 'primary'">
                {{ auth.isAdmin ? '管理员' : '学生' }}
              </el-tag>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout {
  height: 100%;
}
.aside {
  background: #fff;
  border-right: 1px solid #e4e7ed;
}
.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 600;
  color: #409eff;
  border-bottom: 1px solid #e4e7ed;
}
.header {
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.header-title {
  font-size: 16px;
  font-weight: 600;
}
.header-right {
  display: flex;
  align-items: center;
  gap: 18px;
}
.bell-badge {
  margin-right: 4px;
}
.user-chip {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
}
.main {
  padding: 16px;
}
</style>
