<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getRooms, getSeatMap } from '../../api/room'
import { createRoom, updateRoom, deleteRoom, batchGenSeats, updateSeat, listUsers, setUserStatus } from '../../api/admin'

const activeTab = ref('rooms')

// ---------- 房间管理 ----------
const rooms = ref([])
const roomDialog = ref(false)
const editingId = ref(null)
const roomForm = ref({
  name: '', location: '', open_time: '08:00', close_time: '22:00',
  seat_rows: 6, seat_cols: 8, description: ''
})

async function loadRooms() {
  const resp = await getRooms()
  rooms.value = resp.data || []
}

function openCreate() {
  editingId.value = null
  roomForm.value = { name: '', location: '', open_time: '08:00', close_time: '22:00', seat_rows: 6, seat_cols: 8, description: '' }
  roomDialog.value = true
}

function openEdit(r) {
  editingId.value = r.id
  roomForm.value = {
    name: r.name, location: r.location,
    open_time: r.open_time.slice(0, 5), close_time: r.close_time.slice(0, 5),
    seat_rows: r.seat_rows, seat_cols: r.seat_cols, description: r.description
  }
  roomDialog.value = true
}

async function saveRoom() {
  const f = roomForm.value
  if (editingId.value) {
    await updateRoom(editingId.value, f)
    ElMessage.success('房间已更新')
  } else {
    const resp = await createRoom(f)
    ElMessage.success(`房间已创建（ID=${resp.data.id}），可批量生成座位`)
  }
  roomDialog.value = false
  loadRooms()
}

async function onBatchGen(r) {
  try {
    await ElMessageBox.prompt(
      `按行列批量生成 ${r.name} 的座位（将覆盖房间默认规模）`,
      '批量生成座位',
      {
        inputValue: `${r.seat_rows} ${r.seat_cols}`,
        inputPattern: /^\s*\d{1,2}\s+\d{1,2}\s*$/,
        inputErrorMessage: '格式: 行 列（如 6 8）'
      }
    ).then(async ({ value }) => {
      const [rows, cols] = value.trim().split(/\s+/).map(Number)
      const resp = await batchGenSeats(r.id, { seat_rows: rows, seat_cols: cols })
      ElMessage.success(`已生成 ${resp.data.created} 个座位`)
      loadRooms()
    })
  } catch { /* 取消 */ }
}

async function onDeleteRoom(r) {
  try {
    await ElMessageBox.confirm(
      `删除房间「${r.name}」将级联删除其全部座位与预约记录，确认？`,
      '危险操作', { type: 'error', confirmButtonText: '确认删除' }
    )
  } catch { return }
  await deleteRoom(r.id)
  ElMessage.success('已删除')
  loadRooms()
}

// ---------- 座位维护 ----------
const seatDialog = ref(false)
const seatRoom = ref(null)
const seatList = ref([])
const seatDate = ref(new Date().toISOString().slice(0, 10))

async function openSeats(r) {
  seatRoom.value = r
  seatDialog.value = true
  const resp = await getSeatMap(r.id, seatDate.value, '08:00', '22:00')
  seatList.value = resp.data.seats || []
}

async function toggleSeatStatus(s) {
  const next = s.status === 'available' ? 'maintenance' : 'available'
  await updateSeat(s.id, { zone: s.zone, has_power: s.has_power, near_window: s.near_window, status: next })
  s.status = next
  ElMessage.success(`${s.seat_no} 已${next === 'available' ? '恢复可用' : '进入维护'}`)
}

// ---------- 用户管理 ----------
const users = ref([])

async function loadUsers() {
  const resp = await listUsers()
  users.value = resp.data || []
}

async function toggleUser(u) {
  const next = u.status === 'active' ? 'disabled' : 'active'
  await setUserStatus(u.id, next)
  u.status = next
  ElMessage.success(`${u.username} 已${next === 'active' ? '启用' : '禁用'}`)
}

onMounted(() => {
  loadRooms()
  loadUsers()
})
</script>

<template>
  <div class="page-card">
    <h2 style="margin-top: 0">管理端</h2>

    <el-tabs v-model="activeTab">
      <!-- 房间管理 -->
      <el-tab-pane label="自习室管理" name="rooms">
        <div style="margin-bottom: 12px">
          <el-button type="primary" @click="openCreate">新建自习室</el-button>
        </div>
        <el-table :data="rooms" stripe>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="name" label="名称" min-width="130" />
          <el-table-column prop="location" label="位置" min-width="120" />
          <el-table-column label="开放时间" width="130">
            <template #default="{ row }">
              {{ row.open_time.slice(0, 5) }} - {{ row.close_time.slice(0, 5) }}
            </template>
          </el-table-column>
          <el-table-column label="规模" width="90">
            <template #default="{ row }">{{ row.seat_rows }} × {{ row.seat_cols }}</template>
          </el-table-column>
          <el-table-column label="操作" min-width="300">
            <template #default="{ row }">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="success" plain @click="onBatchGen(row)">批量生成座位</el-button>
              <el-button size="small" @click="openSeats(row)">座位维护</el-button>
              <el-button size="small" type="danger" plain @click="onDeleteRoom(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 用户管理 -->
      <el-tab-pane label="用户管理" name="users">
        <el-table :data="users" stripe>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="username" label="用户名" width="110" />
          <el-table-column prop="real_name" label="姓名" width="100" />
          <el-table-column prop="student_no" label="学号" width="110" />
          <el-table-column label="角色" width="90">
            <template #default="{ row }">
              <el-tag size="small" :type="row.role === 'admin' ? 'danger' : 'primary'">
                {{ row.role === 'admin' ? '管理员' : '学生' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="信用分" width="90">
            <template #default="{ row }">
              <span :style="{ color: row.credit_score < 60 ? '#f56c6c' : row.credit_score < 80 ? '#e6a23c' : '#67c23a', fontWeight: 600 }">
                {{ row.credit_score }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag size="small" :type="row.status === 'active' ? 'success' : 'info'">
                {{ row.status === 'active' ? '正常' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button
                size="small"
                :type="row.status === 'active' ? 'danger' : 'success'"
                plain
                :disabled="row.role === 'admin'"
                @click="toggleUser(row)"
              >
                {{ row.status === 'active' ? '禁用' : '启用' }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>

  <!-- 房间编辑对话框 -->
  <el-dialog v-model="roomDialog" :title="editingId ? '编辑自习室' : '新建自习室'" width="480px">
    <el-form :model="roomForm" label-width="90px">
      <el-form-item label="名称" required>
        <el-input v-model="roomForm.name" placeholder="如 1F-静音自习室" />
      </el-form-item>
      <el-form-item label="位置">
        <el-input v-model="roomForm.location" />
      </el-form-item>
      <el-form-item label="开放时间">
        <el-time-select v-model="roomForm.open_time" start="05:00" step="00:30" end="12:00" style="width: 120px" />
        <span style="margin: 0 6px">至</span>
        <el-time-select v-model="roomForm.close_time" start="12:00" step="00:30" end="23:30" style="width: 120px" />
      </el-form-item>
      <el-form-item label="默认规模">
        <el-input-number v-model="roomForm.seat_rows" :min="1" :max="30" /> 行
        <el-input-number v-model="roomForm.seat_cols" :min="1" :max="30" style="margin-left: 8px" /> 列
      </el-form-item>
      <el-form-item label="说明">
        <el-input v-model="roomForm.description" type="textarea" :rows="2" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="roomDialog = false">取消</el-button>
      <el-button type="primary" @click="saveRoom">保存</el-button>
    </template>
  </el-dialog>

  <!-- 座位维护对话框 -->
  <el-dialog v-model="seatDialog" :title="`座位维护 - ${seatRoom?.name || ''}`" width="720px">
    <el-alert type="info" :closable="false" style="margin-bottom: 10px"
      title="点击「维护/恢复」切换座位状态；维护中的座位不可被预约。" />
    <el-table :data="seatList" stripe max-height="480">
      <el-table-column prop="seat_no" label="座位号" width="80" />
      <el-table-column label="坐标" width="80">
        <template #default="{ row }">{{ row.row_no }},{{ row.col_no }}</template>
      </el-table-column>
      <el-table-column prop="zone" label="区域" width="90" />
      <el-table-column label="电源" width="60">
        <template #default="{ row }">{{ row.has_power ? '✔' : '—' }}</template>
      </el-table-column>
      <el-table-column label="靠窗" width="60">
        <template #default="{ row }">{{ row.near_window ? '✔' : '—' }}</template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="row.status === 'available' ? 'success' : 'warning'">
            {{ row.status === 'available' ? '可用' : '维护' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" @click="toggleSeatStatus(row)">
            {{ row.status === 'available' ? '设为维护' : '恢复' }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-dialog>
</template>
