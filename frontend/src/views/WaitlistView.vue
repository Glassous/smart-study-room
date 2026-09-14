<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listMyWaitlist, cancelWaitlist } from '../api/waitlist'

const list = ref([])
const loading = ref(false)

const statusMeta = {
  waiting: { text: '排队中', type: 'warning' },
  promoted: { text: '已递补', type: 'success' },
  cancelled: { text: '已取消', type: 'info' },
  expired: { text: '已过期', type: 'info' }
}
const zoneName = { quiet: '静音区', regular: '普通区', discussion: '研讨区', computer: '机房区' }

async function load() {
  loading.value = true
  try {
    const resp = await listMyWaitlist()
    list.value = resp.data || []
  } finally {
    loading.value = false
  }
}

async function doCancel(row) {
  try {
    await ElMessageBox.confirm('确认退出该场次候补队列？', '取消候补', { type: 'warning' })
  } catch { return }
  await cancelWaitlist(row.id)
  ElMessage.success('已退出候补')
  load()
}

function prefText(row) {
  const ps = []
  if (row.zone) ps.push(zoneName[row.zone] || row.zone)
  if (row.has_power) ps.push('电源')
  if (row.near_window) ps.push('靠窗')
  return ps.length ? ps.join(' / ') : '不限'
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <div class="head">
      <h2 style="margin: 0">我的候补</h2>
      <el-button @click="load">刷新</el-button>
    </div>
    <el-alert type="info" :closable="false" style="margin-bottom: 12px"
      title="满座时段提交候补后，系统会在空位释放时自动按排队顺序递补，并通过站内消息通知您。" />

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column label="日期" width="110">
        <template #default="{ row }">{{ row.res_date }}</template>
      </el-table-column>
      <el-table-column label="时段" width="130">
        <template #default="{ row }">{{ row.start_time.slice(0, 5) }} - {{ row.end_time.slice(0, 5) }}</template>
      </el-table-column>
      <el-table-column label="自习室" prop="room_name" min-width="130" />
      <el-table-column label="偏好" min-width="120">
        <template #default="{ row }">{{ prefText(row) }}</template>
      </el-table-column>
      <el-table-column label="排队位次" width="90">
        <template #default="{ row }">
          <b v-if="row.status === 'waiting'">第 {{ row.position }} 位</b>
          <span v-else>—</span>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="statusMeta[row.status]?.type">
            {{ statusMeta[row.status]?.text || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button v-if="row.status === 'waiting'" size="small" type="danger" plain @click="doCancel(row)">
            退出
          </el-button>
          <span v-else class="dim">—</span>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}
.dim {
  color: #c0c4cc;
}
</style>
