<script setup>
import { ref, computed, onMounted } from 'vue'
import { message, confirmDialog } from '../components/ui/feedback'
import { getRooms, getSeatMap } from '../api/room'
import { createReservation, autoAllocate } from '../api/reservation'
import { joinWaitlist } from '../api/waitlist'
import AppIcon from '../components/AppIcon.vue'
import SButton from '../components/ui/SButton.vue'
import SSelect from '../components/ui/SSelect.vue'
import SDatePicker from '../components/ui/SDatePicker.vue'
import SCheckbox from '../components/ui/SCheckbox.vue'
import SDialog from '../components/ui/SDialog.vue'
import STag from '../components/ui/STag.vue'

// ---- 条件区状态 ----
const rooms = ref([])
const roomId = ref(null)
const date = ref(new Date().toISOString().slice(0, 10))
const start = ref('09:00')
const end = ref('12:00')

// 时间选项(30 分钟粒度)
const timeOptions = []
for (let h = 7; h <= 22; h++) {
  timeOptions.push({ label: `${String(h).padStart(2, '0')}:00`, value: `${String(h).padStart(2, '0')}:00` })
  timeOptions.push({ label: `${String(h).padStart(2, '0')}:30`, value: `${String(h).padStart(2, '0')}:30` })
}

// ---- 平面图 ----
const room = ref(null)
const seats = ref([])
const loading = ref(false)
const selected = ref(null)

const zoneName = { quiet: '静音区', regular: '普通区', discussion: '研讨区', computer: '机房区' }

// 桌型（由 zone + has_power 推导）：机房=电脑桌，有电源=插座桌，其余=普通桌
function seatType(s) {
  if (s.zone === 'computer') return '电脑桌'
  if (s.has_power) return '插座桌'
  return '普通桌'
}
function seatIcon(s) {
  if (s.zone === 'computer') return 'desk-pc'
  if (s.has_power) return 'desk-power'
  return 'desk-book'
}

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
    message.warning('开始时间需早于结束时间')
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

// 行列网格（座位区）
const gridStyle = computed(() => ({
  gridTemplateColumns: `repeat(${room.value?.seat_cols || 8}, minmax(56px, 1fr))`,
  gap: '8px'
}))

// 房间框架（墙/窗包裹座位区），最小宽度保证小容器可横向滚动
const frameStyle = computed(() => {
  const cols = room.value?.seat_cols || 8
  const contentMin = cols * 56 + (cols - 1) * 8 + 16 + 36
  return { minWidth: `${Math.max(480, contentMin)}px` }
})

// 靠窗方位：某条边只有当其全部座位连续靠窗时才判定为窗（避免角落座位误判两条边）
const windowSides = computed(() => {
  const sides = { top: false, bottom: false, left: false, right: false }
  if (!room.value || !seats.value.length) return sides
  const rows = room.value.seat_rows
  const cols = room.value.seat_cols

  const edge = { top: [], bottom: [], left: [], right: [] }
  for (const s of seats.value) {
    if (s.row_no === 1) edge.top.push(s)
    if (s.row_no === rows) edge.bottom.push(s)
    if (s.col_no === 1) edge.left.push(s)
    if (s.col_no === cols) edge.right.push(s)
  }
  // 顶/底边 = 一行（cols 个座位）；左/右边 = 一列（rows 个座位），且需整条边全部靠窗
  sides.top = edge.top.length === cols && edge.top.every((s) => s.near_window)
  sides.bottom = edge.bottom.length === cols && edge.bottom.every((s) => s.near_window)
  sides.left = edge.left.length === rows && edge.left.every((s) => s.near_window)
  sides.right = edge.right.length === rows && edge.right.every((s) => s.near_window)
  return sides
})

// 靠窗座位旁的窗条方向：仅在该侧边确认为窗时渲染
function nearWindowSide(s) {
  const sides = windowSides.value
  const rows = room.value?.seat_rows || 0
  const cols = room.value?.seat_cols || 0
  if (sides.right && s.col_no === cols) return 'seat-near-window-right'
  if (sides.left && s.col_no === 1) return 'seat-near-window-left'
  if (sides.top && s.row_no === 1) return 'seat-near-window-top'
  if (sides.bottom && s.row_no === rows) return 'seat-near-window-bottom'
  return ''
}

function seatClass(s) {
  let cls
  if (s.status !== 'available') cls = 'seat seat-disabled'
  else if (s.occupied) cls = 'seat seat-occupied'
  else if (selected.value?.id === s.id) cls = 'seat seat-selected'
  else cls = 'seat seat-free'
  if (s.near_window) {
    const side = nearWindowSide(s)
    if (side) cls += ' ' + side
  }
  return cls
}

function seatTip(s) {
  return `${s.seat_no} · ${seatType(s)} · ${zoneName[s.zone] || s.zone}${s.near_window ? ' · 靠窗' : ''}${s.status !== 'available' ? ' · ' + (s.status === 'maintenance' ? '维护中' : '停用') : ''}`
}

// 座位悬浮提示（共享单例，fixed 定位避免被滚动容器裁剪）
const hoverTip = ref(null)
function onGridOver(e) {
  const seatEl = e.target.closest('.seat')
  if (!seatEl) return
  const s = seats.value.find((x) => String(x.id) === seatEl.dataset.id)
  if (!s) return
  const rect = seatEl.getBoundingClientRect()
  hoverTip.value = {
    text: seatTip(s),
    x: Math.min(Math.max(rect.left + rect.width / 2, 90), window.innerWidth - 90),
    y: rect.top
  }
}
function onGridLeave() {
  hoverTip.value = null
}

function onSeatClick(s) {
  if (s.status !== 'available') {
    message.info(s.status === 'maintenance' ? '该座位维护中' : '该座位已停用')
    return
  }
  if (s.occupied) {
    message.info('该座位在所选时段已被占用')
    return
  }
  selected.value = s
}

async function confirmBooking() {
  const s = selected.value
  if (!s) return
  try {
    await confirmDialog(
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
  message.success(`预约成功：${resp.data.room_name} ${resp.data.seat_no}`)
  selected.value = null
  loadSeatMap()
}

// ---- 智能分配 ----
const allocVisible = ref(false)
const allocForm = ref({ zone: '', need_power: false, need_window: false })
const recommendations = ref([])
const allocLoading = ref(false)

const zoneOptions = [
  { label: '静音区', value: 'quiet' },
  { label: '普通区', value: 'regular' },
  { label: '研讨区', value: 'discussion' },
  { label: '机房区', value: 'computer' }
]

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
      message.success(`已为您分配：${r.room_name} ${r.seat_no}，请按时签到`)
      allocVisible.value = false
      loadSeatMap()
    } else {
      recommendations.value = resp.data.recommendations || []
      if (!recommendations.value.length) message.info('没有满足条件的推荐')
    }
  } catch (e) {
    if (String(e.message).includes('没有满足条件')) {
      try {
        await confirmDialog('当前时段已满座，是否加入候补队列？', '满座提示', { type: 'info' })
        await joinWaitlist({
          room_id: roomId.value, date: date.value, start_time: start.value, end_time: end.value,
          zone: allocForm.value.zone || undefined
        })
        message.success('已加入候补队列，空出座位将自动递补并通知您')
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
  message.success(`预约成功：${resp.data.room_name} ${resp.data.seat_no}`)
  allocVisible.value = false
  loadSeatMap()
}

// 统计右栏实时数据
const seatStats = computed(() => {
  const total = seats.value.length
  if (!total) return { total: 0, free: 0, occ: 0, down: 0 }
  let occ = 0, down = 0
  seats.value.forEach((s) => {
    if (s.status !== 'available') down++
    else if (s.occupied) occ++
  })
  return { total, free: total - occ - down, occ, down }
})

onMounted(loadRooms)
</script>

<template>
  <div class="page-view">
    <header class="view-heading" data-page-title>
      <h1>座位预约</h1>
      <p class="heading-sub">选择自习室与时段，点击平面图中的座位提交预约</p>
    </header>
    <!-- 座位预约：页面级双栏 = 左筛选 | 右座位图 -->
    <div class="split booking-split">
    <!-- 左栏：筛选 + 智能分配 + 统计 + 图例 -->
    <div class="split-left">
      <section class="card responsive-compact">
        <div class="card-title-row">
          <h3>预约条件</h3>
        </div>
        <div class="filter-group">
          <div class="filter-row">
            <label>自习室</label>
            <SSelect v-model="roomId" :options="rooms" label-key="name" value-key="id" @change="loadSeatMap" />
          </div>
          <div class="filter-row">
            <label>日期</label>
            <SDatePicker
              v-model="date"
              :disabled-date="(d) => d.getTime() < Date.now() - 86400000"
              @change="loadSeatMap"
            />
          </div>
          <div class="filter-row">
            <label>时段</label>
            <div class="time-pair">
              <SSelect v-model="start" :options="timeOptions" @change="loadSeatMap" />
              <span class="arrow">→</span>
              <SSelect v-model="end" :options="timeOptions" @change="loadSeatMap" />
            </div>
            <div class="filter-hint">默认 30 分钟粒度，开始需早于结束</div>
          </div>
        </div>
      </section>

      <section class="card responsive-compact">
        <div class="card-title-row">
          <h3>实时数据</h3>
        </div>
        <div class="kpi-grid">
          <div class="kpi">
            <div class="kpi-num">{{ seatStats.total }}</div>
            <div class="kpi-label">座位总数</div>
          </div>
          <div class="kpi kpi-ok">
            <div class="kpi-num">{{ seatStats.free }}</div>
            <div class="kpi-label">当前空闲</div>
          </div>
          <div class="kpi kpi-bad">
            <div class="kpi-num">{{ seatStats.occ }}</div>
            <div class="kpi-label">时段内占用</div>
          </div>
          <div class="kpi">
            <div class="kpi-num">{{ seatStats.down }}</div>
            <div class="kpi-label">维护/停用</div>
          </div>
        </div>
      </section>

      <section class="card responsive-compact">
        <div class="card-title-row">
          <h3>快捷操作</h3>
        </div>
        <div class="actions">
          <SButton variant="primary" block @click="openAlloc">
            <AppIcon name="sparkles" :size="16" />智能分配
          </SButton>
          <SButton variant="secondary" block @click="loadSeatMap">
            <AppIcon name="refresh" :size="16" />刷新座位图
          </SButton>
        </div>
        <div class="legend">
          <div class="legend-item"><span class="type-ico"><AppIcon name="desk-book" :size="16" /></span><span>普通桌</span></div>
          <div class="legend-item"><span class="type-ico"><AppIcon name="desk-power" :size="16" /></span><span>插座桌</span></div>
          <div class="legend-item"><span class="type-ico"><AppIcon name="desk-pc" :size="16" /></span><span>电脑桌</span></div>
          <div class="legend-item"><span class="type-ico type-ico--window" /><span>窗户</span></div>
        </div>
        <div class="legend legend--status">
          <div class="legend-item"><span class="dot dot-free" /><span>空闲（可预约）</span></div>
          <div class="legend-item"><span class="dot dot-selected" /><span>已选中</span></div>
          <div class="legend-item"><span class="dot dot-occupied" /><span>占用</span></div>
          <div class="legend-item"><span class="dot dot-disabled" /><span>维护/停用</span></div>
        </div>
      </section>
    </div>

    <!-- 右栏：座位平面图 + 已选条 -->
    <div class="split-right">
      <section class="card seat-card">
        <div class="card-title-row">
          <div>
            <h3 style="margin:0">座位平面图</h3>
            <div v-if="room" class="room-meta muted">
              {{ room.name }} · {{ room.location }} · 开放 {{ room.open_time?.slice(0, 5) }}–{{ room.close_time?.slice(0, 5) }} · {{ room.seat_rows }}×{{ room.seat_cols }}
            </div>
          </div>
        </div>

        <div class="responsive-scroll" tabindex="0" aria-label="座位平面图，可左右滑动">
        <div v-loading="loading" class="seat-grid-wrap" @mouseover="onGridOver" @mouseleave="onGridLeave">
          <div v-if="!room" class="empty-tip muted">请先在左侧选择自习室并设置预约条件</div>
          <div v-else :style="frameStyle" class="room-frame">
            <!-- 顶部：窗或墙 -->
            <div class="wall wall-top" :class="{ window: windowSides.top }" aria-hidden="true"></div>
            <!-- 左：窗或墙 -->
            <div class="wall wall-left" :class="{ window: windowSides.left }" aria-hidden="true"></div>
            <!-- 座位网格 -->
            <div :style="gridStyle" class="seat-grid">
              <div
                v-for="s in seats"
                :key="s.id"
                :class="seatClass(s)"
                :data-id="s.id"
                role="button"
                :aria-label="seatTip(s)"
                tabindex="-1"
                @click="onSeatClick(s)"
              >
                <span class="seat-icon"><AppIcon :name="seatIcon(s)" :size="28" /></span>
                <span class="seat-no">{{ s.seat_no }}</span>
              </div>
            </div>
            <!-- 右：窗或墙 -->
            <div class="wall wall-right" :class="{ window: windowSides.right }" aria-hidden="true"></div>
            <!-- 底部：窗或墙（有门时底部为墙） -->
            <div class="wall wall-bottom" :class="{ window: windowSides.bottom }">
              <template v-if="!windowSides.bottom">
                <span class="wall-seg" aria-hidden="true"></span>
                <span class="door" aria-hidden="true"></span>
                <span class="wall-seg" aria-hidden="true"></span>
              </template>
            </div>
          </div>
        </div>
        </div>
      </section>

      <!-- 已选座位操作条（常驻可见：未选座位时为提示态，提交按钮禁用） -->
      <section class="card selected-card" :class="{ 'selected-card--empty': !selected }">
        <div class="sel-info">
          <div class="sel-icon"><AppIcon name="seat" :size="25" /></div>
          <div class="sel-detail">
            <template v-if="selected">
              <div class="sel-title">
                已选择 <b>{{ selected.seat_no }}</b>
                <STag>{{ zoneName[selected.zone] || selected.zone }}</STag>
                <STag v-if="selected.has_power" type="success">电源</STag>
                <STag v-if="selected.near_window" type="warning">靠窗</STag>
              </div>
              <div class="muted">{{ date }} · {{ start }} – {{ end }} · {{ room?.name }}</div>
            </template>
            <template v-else>
              <div class="sel-placeholder">请选择座位</div>
              <div class="muted">点击座位平面图中可预约的座位后，再提交预约</div>
            </template>
          </div>
        </div>
        <SButton variant="primary" size="lg" :disabled="!selected" @click="confirmBooking">提交预约</SButton>
      </section>
    </div>
  </div>

  <!-- 座位悬浮提示 -->
  <Teleport to="body">
    <div
      v-if="hoverTip"
      class="seat-hover-tip"
      :style="{ left: `${hoverTip.x}px`, top: `${hoverTip.y}px` }"
      aria-hidden="true"
    >
      {{ hoverTip.text }}
    </div>
  </Teleport>

  <!-- 智能分配对话框 -->
  <SDialog v-model="allocVisible" title="智能分配座位" width="520px">
    <div class="alloc-form">
      <div class="filter-row">
        <label>偏好区域</label>
        <SSelect v-model="allocForm.zone" :options="zoneOptions" placeholder="不限（推荐）" clearable />
      </div>
      <div class="pref-checks">
        <SCheckbox v-model="allocForm.need_power">需要电源插座</SCheckbox>
        <SCheckbox v-model="allocForm.need_window">偏好靠窗</SCheckbox>
      </div>
    </div>

    <div v-if="recommendations.length" class="rec-title">推荐座位（评分从高到低）</div>
    <div v-if="recommendations.length" class="rec-list">
      <div v-for="(r, i) in recommendations" :key="r.seat.id" class="rec-item">
        <div class="rec-rank">{{ i + 1 }}</div>
        <div class="rec-main">
          <div class="rec-seat">
            <b>{{ r.seat.seat_no }}</b>
            <STag>{{ zoneName[r.seat.zone] || r.seat.zone }}</STag>
            <span v-if="r.seat.has_power" class="muted">· 电源</span>
            <span v-if="r.seat.near_window" class="muted">· 靠窗</span>
          </div>
          <div class="muted rec-score">匹配得分 {{ r.score.toFixed(1) }}</div>
        </div>
        <SButton size="sm" variant="primary" @click="pickRecommended(r)">选这个预约</SButton>
      </div>
    </div>

    <template #footer>
      <div class="alloc-footer">
        <SButton :loading="allocLoading" variant="secondary" @click="runAllocate(false)">
          <AppIcon name="eye" :size="16" />仅推荐
        </SButton>
        <SButton :loading="allocLoading" variant="primary" @click="runAllocate(true)">
          <AppIcon name="sparkles" :size="16" />立即分配并下单
        </SButton>
      </div>
    </template>
  </SDialog>
  </div>
</template>

<style scoped>
.booking-split {
  grid-template-columns: 320px 1fr;
  align-items: start;
}

/* 时段选择并排 */
.time-pair {
  display: grid;
  grid-template-columns: 1fr 20px 1fr;
  gap: 6px;
  align-items: center;
}
.time-pair .arrow {
  text-align: center;
  color: var(--text-4);
}

.actions {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
  margin-bottom: 14px;
}

.legend {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px 12px;
  padding-top: 12px;
  border-top: 1px solid var(--hairline);
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: var(--fs-caption);
  color: var(--text-3);
}
.dot {
  width: 15px;
  height: 15px;
  border-radius: 5px;
  display: inline-block;
  border: 1px solid transparent;
}
.dot-free     { background: var(--seat-free-bg); border-color: var(--seat-free-border); }
.dot-selected { background: var(--primary); }
.dot-occupied { background: var(--seat-occupied-bg); border-color: var(--seat-occupied-border); }
.dot-disabled { background: var(--surface-3); border-color: var(--border); }

/* 桌型 / 窗户 图例 */
.type-ico {
  width: 16px;
  height: 16px;
  display: inline-grid;
  place-items: center;
  color: var(--text-3);
  flex: 0 0 auto;
}
.type-ico--window {
  background:
    linear-gradient(180deg, var(--window-glass-hi), transparent 60%),
    repeating-linear-gradient(90deg, var(--window-frame) 0 2px, var(--window-glass) 2px 8px);
  border: 1px solid var(--window-frame);
  border-radius: 3px;
}
.legend--status {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--hairline);
}

/* 座位图 */
.seat-card {
  min-height: 0;
}
.room-meta {
  margin-top: 3px;
}
.seat-grid-wrap {
  margin-top: 8px;
  padding: 12px;
  background: var(--surface-2);
  border: 1px solid var(--hairline);
  border-radius: var(--r-lg);
  min-height: 280px;
  display: grid;
  place-items: start stretch;
}
.empty-tip {
  padding: 60px 0;
  font-size: var(--fs-body);
}
/* 房间框架：墙 + 窗 + 门 包裹座位区 */
.room-frame {
  width: 100%;
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr) 18px;
  grid-template-rows: 18px auto 18px;
  border-radius: var(--r-lg);
  overflow: hidden;
}
.wall { background: var(--wall-solid); }
.wall-top { grid-column: 1 / -1; grid-row: 1; }
.wall-left { grid-column: 1; grid-row: 2; }
.wall-right { grid-column: 3; grid-row: 2; }
.wall-bottom {
  grid-column: 1 / -1;
  grid-row: 3;
  display: flex;
  align-items: stretch;
}
.wall-seg { flex: 1 1 auto; }

/* 窗户：水平墙（顶/底）竖窗棂，竖直墙（左/右）横窗棂 */
.wall-top.window,
.wall-bottom.window {
  background:
    linear-gradient(180deg, var(--window-glass-hi), transparent 60%),
    repeating-linear-gradient(90deg, var(--window-frame) 0 3px, var(--window-glass) 3px 22px);
}
.wall-left.window,
.wall-right.window {
  background:
    linear-gradient(90deg, var(--window-glass-hi), transparent 60%),
    repeating-linear-gradient(180deg, var(--window-frame) 0 3px, var(--window-glass) 3px 22px);
}

/* 门：底墙中间开口 */
.door {
  flex: 0 0 46px;
  position: relative;
  background: var(--door);
  border: 2px solid var(--door-frame);
  border-bottom: none;
  border-radius: 6px 6px 0 0;
}
.door::after {
  content: '';
  position: absolute;
  right: 7px;
  top: 6px;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--door-frame);
}

/* 座位区 */
.seat-grid {
  grid-column: 2;
  grid-row: 2;
  min-width: 0;
  display: grid;
  padding: 8px;
  background: var(--room-floor);
}
.seat {
  min-height: 48px;
  border-radius: var(--r-md);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  font-variant-numeric: tabular-nums;
  letter-spacing: .02em;
  cursor: pointer;
  position: relative;
  transition: transform var(--dur-1) var(--ease), box-shadow var(--dur-1) var(--ease),
    background var(--dur-1) var(--ease), border-color var(--dur-1) var(--ease);
  user-select: none;
  border: 1px solid transparent;
}
.seat:hover {
  transform: translateY(-1px);
}
.seat-icon {
  display: grid;
  place-items: center;
  line-height: 0;
  opacity: .92;
}
.seat-no {
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
  opacity: .85;
}
/* 靠窗座位：桌面靠窗一侧淡蓝窗条（方向随数据） */
.seat-near-window-right::after,
.seat-near-window-left::after {
  content: '';
  position: absolute;
  top: 7px;
  bottom: 7px;
  width: 3px;
  border-radius: 2px;
  background: var(--window-frame);
}
.seat-near-window-right::after { right: 3px; }
.seat-near-window-left::after { left: 3px; }
.seat-near-window-top::after,
.seat-near-window-bottom::after {
  content: '';
  position: absolute;
  left: 7px;
  right: 7px;
  height: 3px;
  border-radius: 2px;
  background: var(--window-frame);
}
.seat-near-window-top::after { top: 3px; }
.seat-near-window-bottom::after { bottom: 3px; }
.seat-free {
  background: var(--seat-free-bg);
  color: var(--seat-free-color);
  border-color: var(--seat-free-border);
}
.seat-free:hover {
  background: var(--seat-free-hover-bg);
  border-color: var(--seat-free-hover-border);
}
.seat-selected {
  background: var(--primary);
  color: #fff;
  border-color: var(--primary-active);
  box-shadow: 0 0 0 3px var(--primary-ring), 0 4px 12px rgba(59, 102, 218, .32);
}
.seat-occupied {
  background: var(--seat-occupied-bg);
  color: var(--seat-occupied-color);
  border-color: var(--seat-occupied-border);
  cursor: not-allowed;
}
.seat-disabled {
  background:
    repeating-linear-gradient(135deg, var(--surface-3) 0 5px, var(--seat-disabled-stripe) 5px 9px);
  color: var(--text-4);
  border-color: var(--border);
  cursor: not-allowed;
  text-decoration: line-through;
  text-decoration-color: var(--text-4);
}

/* 座位悬浮提示 */
.seat-hover-tip {
  position: fixed;
  z-index: var(--z-popover);
  transform: translate(-50%, calc(-100% - 10px));
  background: var(--seat-tip-bg);
  color: var(--seat-tip-text);
  font-size: var(--fs-caption);
  padding: 5px 10px;
  border-radius: var(--r-sm);
  pointer-events: none;
  white-space: nowrap;
  box-shadow: var(--shadow-2);
}

.seat-hover-tip::after {
  content: '';
  position: absolute;
  left: 50%;
  bottom: -4px;
  width: 8px;
  height: 8px;
  transform: translateX(-50%) rotate(45deg);
  background: var(--seat-tip-bg);
}

/* 已选条 */
.selected-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  background: var(--primary-faint);
  border-color: var(--border-strong);
}
.sel-info {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}
.sel-detail {
  min-width: 0;
}
.sel-icon {
  width: 46px;
  height: 46px;
  border-radius: var(--r-lg);
  background: var(--surface);
  display: grid;
  place-items: center;
  color: var(--primary);
  box-shadow: var(--shadow-1);
  flex: 0 0 auto;
}
.sel-title {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--text-1);
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 2px;
}
/* 未选择座位：常驻提示态 */
.selected-card--empty {
  background: var(--surface);
  border-color: var(--border);
}
.selected-card--empty .sel-icon {
  background: var(--surface-3);
  color: var(--text-4);
  box-shadow: none;
}
.selected-card--empty .sel-detail .muted {
  color: var(--text-4);
}
.sel-placeholder {
  font-size: var(--fs-title);
  font-weight: 600;
  color: var(--text-3);
  margin-bottom: 2px;
}

/* 智能分配弹窗内 */
.alloc-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.pref-checks {
  display: flex;
  gap: 18px;
}
.rec-title {
  margin: 20px 0 10px;
  font-weight: 600;
  color: var(--text-2);
  font-size: var(--fs-body-sm);
}
.rec-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 264px;
  overflow: auto;
  scrollbar-width: thin;
}
.rec-item {
  display: grid;
  grid-template-columns: 34px 1fr auto;
  gap: 12px;
  align-items: center;
  padding: 10px 12px;
  border-radius: var(--r-lg);
  background: var(--surface-2);
  border: 1px solid var(--hairline);
}
.rec-rank {
  width: 28px;
  height: 28px;
  border-radius: var(--r-sm);
  background: var(--primary-weak);
  color: var(--primary-active);
  display: grid;
  place-items: center;
  font-weight: 700;
  font-size: var(--fs-caption);
  font-variant-numeric: tabular-nums;
}
.rec-main {
  min-width: 0;
}
.rec-seat {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: var(--fs-body);
}
.rec-score {
  margin-top: 2px;
  font-size: var(--fs-caption);
  font-variant-numeric: tabular-nums;
}
.alloc-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  width: 100%;
}

@media (max-width: 720px) {
  .seat-grid-wrap { padding: 10px; min-width: max-content; }
  .seat-grid { padding: 8px; }
  .seat { border-radius: var(--r-sm); font-size: 10px; }
  .selected-card { align-items: stretch; flex-direction: column; }
  .selected-card :deep(.s-btn) { width: 100%; }
  .pref-checks { align-items: flex-start; flex-direction: column; gap: 10px; }
  .rec-item { grid-template-columns: 32px minmax(0, 1fr); }
  .rec-item :deep(.s-btn) { grid-column: 1 / -1; width: 100%; }
  .seat-card { overflow: hidden; }
  .sel-info { min-width: 0; align-items: flex-start; }
  .sel-title { flex-wrap: wrap; }
  .alloc-footer { align-items: stretch; flex-direction: column; }
  .alloc-footer :deep(.s-btn) { width: 100%; }
}
</style>
