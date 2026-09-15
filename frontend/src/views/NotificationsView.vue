<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listNotifications, markRead, markAllRead } from '../api/notification'

const list = ref([])
const loading = ref(false)

const typeMeta = {
  reservation_success: { text: '预约成功', color: '#67c23a' },
  checkin_reminder: { text: '签到提醒', color: '#409eff' },
  violation: { text: '违约警告', color: '#f56c6c' },
  credit_change: { text: '信用变动', color: '#e6a23c' },
  waitlist_promoted: { text: '候补递补', color: '#9b59f6' },
  system: { text: '系统', color: '#909399' }
}

async function load() {
  loading.value = true
  try {
    const resp = await listNotifications()
    list.value = resp.data || []
  } finally {
    loading.value = false
  }
}

async function onMarkRead(n) {
  if (n.is_read) return
  await markRead(n.id)
  n.is_read = true
}

async function onMarkAll() {
  const resp = await markAllRead()
  ElMessage.success(`已标记 ${resp.data.marked} 条为已读`)
  load()
}

function fmtTime(t) {
  return new Date(t).toLocaleString('zh-CN', { hour12: false })
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <div class="head">
      <h2 style="margin: 0">消息中心</h2>
      <el-button @click="onMarkAll">全部已读</el-button>
    </div>

    <div v-loading="loading" class="list">
      <el-empty v-if="!list.length" description="暂无消息" />
      <div
        v-for="n in list"
        :key="n.id"
        class="msg-item"
        :class="{ unread: !n.is_read }"
        @click="onMarkRead(n)"
      >
        <el-tag size="small" :color="typeMeta[n.type]?.color" effect="dark" style="border: none">
          {{ typeMeta[n.type]?.text || n.type }}
        </el-tag>
        <div class="msg-body">
          <div class="msg-title">
            {{ n.title }}
            <el-badge v-if="!n.is_read" is-dot class="dot" />
          </div>
          <div class="msg-content">{{ n.content }}</div>
        </div>
        <span class="msg-time">{{ fmtTime(n.created_at) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}
.msg-item {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  padding: 12px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
  border-radius: 6px;
}
.msg-item:hover {
  background: #f5f7fa;
}
.msg-item.unread {
  background: #ecf5ff40;
}
.msg-body {
  flex: 1;
}
.msg-title {
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 6px;
}
.msg-content {
  color: #606266;
  font-size: 13px;
  margin-top: 4px;
}
.msg-time {
  color: #c0c4cc;
  font-size: 12px;
  white-space: nowrap;
}
</style>
