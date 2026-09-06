<script setup>
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login, register } from '../api/auth'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const mode = ref('login') // login | register
const loading = ref(false)

const loginForm = reactive({ username: 'stu01', password: '123456' })
const regForm = reactive({ username: '', password: '', confirm: '', real_name: '', student_no: '' })

async function onLogin() {
  if (!loginForm.username || !loginForm.password) {
    ElMessage.warning('请输入用户名与口令')
    return
  }
  loading.value = true
  try {
    const resp = await login(loginForm)
    auth.setLogin(resp.data)
    ElMessage.success('登录成功')
    router.push(route.query.redirect || '/')
  } finally {
    loading.value = false
  }
}

async function onRegister() {
  const f = regForm
  if (!f.username || !f.password || !f.real_name) {
    ElMessage.warning('请完整填写必填项')
    return
  }
  if (f.password.length < 6) {
    ElMessage.warning('口令至少 6 位')
    return
  }
  if (f.password !== f.confirm) {
    ElMessage.warning('两次输入的口令不一致')
    return
  }
  loading.value = true
  try {
    await register({
      username: f.username,
      password: f.password,
      real_name: f.real_name,
      student_no: f.student_no
    })
    ElMessage.success('注册成功，请登录')
    loginForm.username = f.username
    mode.value = 'login'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-wrap">
    <div class="login-card">
      <h1 class="title">📚 智能共享自习室</h1>
      <p class="subtitle">座位预约 · 智能分配 · 信用治理</p>

      <el-tabs v-model="mode" stretch>
        <el-tab-pane label="登录" name="login">
          <el-form @submit.prevent>
            <el-form-item>
              <el-input v-model="loginForm.username" placeholder="用户名" size="large" />
            </el-form-item>
            <el-form-item>
              <el-input
                v-model="loginForm.password"
                type="password"
                placeholder="口令"
                size="large"
                show-password
                @keyup.enter="onLogin"
              />
            </el-form-item>
            <el-button type="primary" size="large" class="submit" :loading="loading" @click="onLogin">
              登 录
            </el-button>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="注册" name="register">
          <el-form @submit.prevent>
            <el-form-item>
              <el-input v-model="regForm.username" placeholder="用户名(≥3位)" size="large" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="regForm.real_name" placeholder="真实姓名" size="large" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="regForm.student_no" placeholder="学号(选填)" size="large" />
            </el-form-item>
            <el-form-item>
              <el-input v-model="regForm.password" type="password" placeholder="口令(≥6位)" size="large" show-password />
            </el-form-item>
            <el-form-item>
              <el-input v-model="regForm.confirm" type="password" placeholder="确认口令" size="large" show-password />
            </el-form-item>
            <el-button type="primary" size="large" class="submit" :loading="loading" @click="onRegister">
              注 册
            </el-button>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <p class="tip">演示账号：stu01 / 123456 · 管理员：admin / admin123</p>
    </div>
  </div>
</template>

<style scoped>
.login-wrap {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 400px;
  background: #fff;
  border-radius: 12px;
  padding: 32px 36px;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.18);
}
.title {
  text-align: center;
  margin: 0 0 4px;
  color: #303133;
}
.subtitle {
  text-align: center;
  color: #909399;
  margin: 0 0 18px;
  font-size: 13px;
}
.submit {
  width: 100%;
}
.tip {
  margin-top: 14px;
  font-size: 12px;
  color: #909399;
  text-align: center;
}
</style>
