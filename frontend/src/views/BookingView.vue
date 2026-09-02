<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRooms, getSeatMap } from '../api/room'
import { createReservation, autoAllocate } from '../api/reservation'
import { joinWaitlist } from '../api/waitlist'

// ---- 条件区状态 ----
const rooms = ref([])
const roomId = ref(null)
const date = ref(new Date().toISOString().slice(0, 10))
const start = ref('09:00')
const end = ref('12:00')

// 时间选项(30 分钟粒度)
const timeOptions = []
for (let h = 7; h <= 22; h++) {
  timeOptions.push(`${String(h).padStart(2, '0')}:00`)
  timeOptions.push(`${String(h).padStart(2, '0')}:30`)
}

// ---- 平面图 ----
const room = ref(null)
const seats = ref([])
const loading = ref(false)
const selected = ref(null)

const zoneName = { quiet: '静音区', regular: '普通区', discussion: '研讨区', computer: '机房区' }

async function loadRooms() {
  const resp = await getRooms()
  rooms.value = resp.data || []
  if (rooms.value.length && !roomId.value) {
    roomId.value = rooms.value[0].id
    loadSeatMap()
  }
}

async function loadSeatMap() {
  if (!roomId.value || !date.value || !start.value || !end.value) return
  if (start.value >= end.value) {
    ElMessage.warning('开始时间需早于结束时间')
    return
  }
  loading.value = true
  selected.value = null
  try {
    const resp = await getSeatMap(roomId.value, date.value, start.value, end.value)
    room.value = resp.data.room
    seats.value = resp.data.seats || []
  } finally {
    loading.value = false
  }
}

// 行列网格
const gridStyle = computed(() => ({
  display: 'grid',
  gridTemplateColumns: `repeat(${room.value?.seat_cols || 8}, 1fr)`,
  gap: '8px'
}))

function seatClass(s) {
  if (s.status !== 'available') return 'seat seat-disabled'
  if (s.occupied) return 'seat seat-occupied'
  if (selected.value?.id === s.id) return 'seat seat-selected'
  return 'seat seat-free'
}

function seatTip(s) {
  return `${s.seat_no} · ${zoneName[s.zone] || s.zone}${s.has_power ? ' · 电源' : ''}${s.near_window ? ' · 靠窗' : ''}${s.status !== 'available' ? ' · ' + (s.status === 'maintenance' ? '维护中' : '停用') : ''}`
}

function onSeatClick(s) {
  if (s.status !== 'available') {
    ElMessage.info(s.status === 'maintenance' ? '该座位维护中' : '该座位已停用')
    return
  }
  if (s.occupied) {
    ElMessage.info('该座位在所选时段已被占用')
    return
  }
  selected.value = s
}

async function confirmBooking() {
  const s = selected.value
  if (!s) return
  try {
    await ElMessageBox.confirm(
      `确认预约 ${room.value.name} ${s.seat_no} 座位？\n日期：${date.value}　时段：${start.value} - ${end.value}`,
      '预约确认',
      { confirmButtonText: '确认预约', cancelButtonText: '再想想' }
    )
  } catch {
    return
  }
  const resp = await createReservation({
    seat_id: s.id, date: date.value, start_time: start.value, end_time: end.value
  })
  ElMessage.success(`预约成功：${resp.data.room_name} ${resp.data.seat_no}`)
  selected.value = null
  loadSeatMap()
}

// ---- 智能分配 ----
const allocVisible = ref(false)
const allocForm = ref({ zone: '', has_power: false, near_window: false, need_power: false, need_window: false })
const recommendations = ref([])
const allocLoading = ref(false)

function openAlloc() {
  recommendations.value = []
  allocVisible.value = true
}

async function runAllocate(autoBook = false) {
  allocLoading.value = true
  try {
    const resp = await autoAllocate({
      room_id: roomId.value, date: date.value, start_time: start.value, end_time: end.value,
      zone: allocForm.value.zone || undefined,
      has_power: allocForm.value.need_power || undefined,
      near_window: allocForm.value.need_window || undefined,
      auto_book: autoBook
    })
    if (autoBook) {
      const r = resp.data.reservation
      ElMessage.success(`已为您分配：${r.room_name} ${r.seat_no}，请按时签到`)
      allocVisible.value = false
      loadSeatMap()
    } else {
      recommendations.value = resp.data.recommendations || []
      if (!recommendations.value.length) ElMessage.info('没有满足条件的推荐')
    }
  } catch (e) {
    // 404 无空位 → 引导候补
    if (String(e.message).includes('没有满足条件')) {
      try {
        await ElMessageBox.confirm('当前时段已满座，是否加入候补队列？', '满座提示', { type: 'info' })
        await joinWaitlist({
          room_id: roomId.value, date: date.value, start_time: start.value, end_time: end.value,
          zone: allocForm.value.zone || undefined
        })
        ElMessage.success('已加入候补队列，空出座位将自动递补并通知您')
        allocVisible.value = false
      } catch { /* 取消 */ }
    }
  } finally {
    allocLoading.value = false
  }
}

async function pickRecommended(rec) {
  const resp = await createReservation({
    seat_id: rec.seat.id, date: date.value, start_time: start.value, end_time: end.value
  })
  ElMessage.success(`预约成功：${resp.data.room_name} ${resp.data.seat_no}`)
  allocVisible.value = false
  loadSeatMap()
}

onMounted(loadRooms)
</script>

<template>
  <div class="page-card">
    <h2 style="margin-top: 0">座位预约</h2>

    <!-- 条件选择 -->
    <div class="filters">
      <span class="filter-label">自习室</span>
      <el-select v-model="roomId" style="width: 200px" @change="loadSeatMap">
        <el-option v-for="r in rooms" :key="r.id" :label="r.name" :value="r.id" />
      </el-select>

      <span class="filter-label">日期</span>
      <el-date-picker v-model="date" type="date" value-format="YYYY-MM-DD"
        :disabled-date="(d) => d.getTime() < Date.now() - 86400000" style="width: 150px" @change="loadSeatMap" />

      <span class="filter-label">开始</span>
      <el-select v-model="start" style="width: 100px" @change="loadSeatMap">
        <el-option v-for="t in timeOptions" :key="t" :label="t" :value="t" />
      </el-select>
      <span class="filter-label">结束</span>
      <el-select v-model="end" style="width: 100px" @change="loadSeatMap">
        <el-option v-for="t in timeOptions" :key="t" :label="t" :value="t" />
      </el-select>

      <el-button type="primary" plain @click="openAlloc">✨ 智能分配</el-button>
    </div>

    <div v-if="room" class="room-info">
      {{ room.name }} · {{ room.location }} · 开放 {{ room.open_time.slice(0, 5) }}-{{
        room.close_time.slice(0, 5)
      }}
      <el-tag size="small" type="info">绿=空闲</el-tag>
      <el-tag size="small" type="danger">红=占用</el-tag>
      <el-tag size="small" type="warning">橙=维护</el-tag>
    </div>

    <!-- 座位平面图 -->
    <div v-loading="loading" class="seat-grid-wrap">
      <div :style="gridStyle">
        <el-tooltip v-for="s in seats" :key="s.id" :content="seatTip(s)" placement="top">
          <div :class="seatClass(s)" @click="onSeatClick(s)">
            {{ s.seat_no }}
          </div>
        </el-tooltip>
      </div>
    </div>

    <!-- 已选座位操作 -->
    <div v-if="selected" class="selected-bar">
      已选 <b>{{ selected.seat_no }}</b>（{{ zoneName[selected.zone] }}{{ selected.has_power ? ' · 电源' : ''
      }}{{ selected.near_window ? ' · 靠窗' : '' }}）
      <el-button type="primary" @click="confirmBooking">提交预约</el-button>
    </div>
  </div>

  <!-- 智能分配对话框 -->
  <el-dialog v-model="allocVisible" title="智能自动分配" width="560px">
    <el-form label-width="90px">
      <el-form-item label="区域偏好">
        <el-select v-model="allocForm.zone" placeholder="不限" clearable style="width: 160px">
          <el-option label="静音区" value="quiet" />
          <el-option label="普通区" value="regular" />
          <el-option label="研讨区" value="discussion" />
          <el-option label="机房区" value="computer" />
        </el-select>
      </el-form-item>
      <el-form-item label="需要电源">
        <el-switch v-model="allocForm.need_power" />
      </el-form-item>
      <el-form-item label="需要靠窗">
        <el-switch v-model="allocForm.need_window" />
      </el-form-item>
      <el-form-item>
        <el-button :loading="allocLoading" @click="runAllocate(false)">获取推荐</el-button>
        <el-button type="primary" :loading="allocLoading" @click="runAllocate(true)">一键分配最优座位</el-button>
      </el-form-item>
    </el-form>

    <div v-if="recommendations.length">
      <el-divider content-position="left">推荐结果（按匹配度排序）</el-divider>
      <div v-for="(rec, i) in recommendations" :key="rec.seat.id" class="rec-item">
        <b>#{{ i + 1 }} {{ rec.seat.seat_no }}</b>
        <el-tag size="small" type="success">评分 {{ rec.score.toFixed(1) }}</el-tag>
        <span class="rec-reasons">{{ rec.reasons.join('；') }}</span>
        <el-button size="small" type="primary" plain @click="pickRecommended(rec)">选它</el-button>
      </div>
    </div>
  </el-dialog>
</template>

<style scoped>
.filters {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.filter-label {
  color: #606266;
  font-size: 14px;
}
.room-info {
  margin: 14px 0 10px;
  color: #606266;
  display: flex;
  align-items: center;
  gap: 8px;
}
.seat-grid-wrap {
  max-height: 480px;
  overflow: auto;
  padding: 12px;
  background: #fafbfc;
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
}
.seat {
  height: 40px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  cursor: pointer;
  user-select: none;
  border: 1px solid transparent;
  transition: transform 0.1s;
}
.seat:hover {
  transform: scale(1.06);
}
.seat-free {
  background: #f0f9eb;
  color: #67c23a;
  border-color: #c2e7b0;
}
.seat-occupied {
  background: #fef0f0;
  color: #f56c6c;
  border-color: #fbc4c4;
  cursor: not-allowed;
}
.seat-disabled {
  background: #fdf6ec;
  color: #e6a23c;
  border-color: #f5dab1;
  cursor: not-allowed;
}
.seat-selected {
  background: #409eff;
  color: #fff;
  border-color: #409eff;
}
.selected-bar {
  margin-top: 14px;
  display: flex;
  align-items: center;
  gap: 12px;
}
.rec-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  border-bottom: 1px solid #f0f0f0;
}
.rec-reasons {
  color: #909399;
  font-size: 12px;
  flex: 1;
}
</style>
