<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()

// 菜单随迭代补充: 预约/我的预约/热力图/消息/个人中心/管理端
const menus = computed(() => {
  const items = [
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
