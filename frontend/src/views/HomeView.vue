<script setup>
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
</script>

<template>
  <div class="page-card">
    <h2>欢迎使用智能共享自习室预约系统</h2>
    <p v-if="auth.isLoggedIn">
      你好，{{ auth.user?.real_name || auth.user?.username }}！
      更多功能（座位预约、热力图等）正在开发中，敬请期待。
    </p>
  </div>
</template>
