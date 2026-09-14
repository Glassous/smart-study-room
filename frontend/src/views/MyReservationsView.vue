<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  listMyReservations, cancelReservation, checkinReservation,
  leaveReservation, returnReservation, checkoutReservation
} from '../api/reservation'

const list = ref([])
const loading = ref(false)

const statusMeta = {
  pending: { text: '待签到', type: 'warning' },
  checked_in: { text: '使用中', type: 'success' },
  temp_leave: { text: '临时离开', type: 'info' },
  completed: { text: '已完成', type: '' },
  cancelled: { text: '已取消', type: 'info' },
  violation: { text: '已违约', type: 'danger' }
}
const sourceMeta = { manual: '手动选座', auto: '智能分配', waitlist: '候补递补' }

async function load() {
  loading.value = true
  try {
    const resp = await listMyReservations()
    list.value = resp.data || []
  } finally {
    loading.value = false
  }
}

// 各状态可执行的操作
function actionsOf(r) {
  switch (r.status) {
    case 'pending':
      return [
        { key: 'checkin', label: '签到', type: 'primary' },
        { key: 'cancel', label: '取消预约', type: 'danger', plain: true }
      ]
    case 'checked_in':
      return [
        { key: 'leave', label: '临时离开', type: 'warning', plain: true },
        { key: 'checkout', label: '签退', type: 'primary' }
      ]
    case 'temp_leave':
      return [
        { key: 'return', label: '返回座位', type: 'success' },
        { key: 'checkout', label: '提前结束', type: 'info', plain: true }
      ]
    default:
      return []
  }
}

const apiMap = {
  checkin: checkinReservation,
  cancel: cancelReservation,
  leave: leaveReservation,
  return: returnReservation,
  checkout: checkoutReservation
}
const actionText = {
  checkin: '签到', cancel: '取消', leave: '临时离开', return: '返回', checkout: '签退'
}

async function doAction(r, key) {
  if (key === 'cancel') {
    try {
      await ElMessageBox.confirm(
        '距开始不足 30 分钟的取消将扣除 2 信用分，确认取消？',
        '取消预约', { confirmButtonText: '确认取消', type: 'warning' }
      )
    } catch { return }
  }
  await apiMap[key](r.id)
  ElMessage.success(actionText[key] + '成功')
  load()
}

onMounted(load)
</script>

<template>
  <div class="page-card">
    <div class="head">
      <h2 style="margin: 0">我的预约</h2>
      <el-button @click="load">刷新</el-button>
    </div>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column label="日期" width="110">
        <template #default="{ row }">{{ row.res_date }}</template>
      </el-table-column>
      <el-table-column label="时段" width="130">
        <template #default="{ row }">{{ row.start_time.slice(0, 5) }} - {{ row.end_time.slice(0, 5) }}</template>
      </el-table-column>
      <el-table-column label="自习室" prop="room_name" min-width="130" />
      <el-table-column label="座位" width="80">
        <template #default="{ row }">
          <b>{{ row.seat_no }}</b>
        </template>
      </el-table-column>
      <el-table-column label="来源" width="90">
        <template #default="{ row }">
          <el-tag size="small" effect="plain">{{ sourceMeta[row.source] || row.source }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="statusMeta[row.status]?.type">
            {{ statusMeta[row.status]?.text || row.status }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" min-width="220">
        <template #default="{ row }">
          <el-button
            v-for="a in actionsOf(row)"
            :key="a.key"
            size="small"
            :type="a.type"
            :plain="a.plain"
            @click="doAction(row, a.key)"
          >
            {{ a.label }}
          </el-button>
          <span v-if="!actionsOf(row).length" class="dim">—</span>
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
