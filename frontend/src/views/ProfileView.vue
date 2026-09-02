<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { getCreditOverview } from '../api/credit'

const auth = useAuthStore()
const overview = ref(null)

const scoreColor = computed(() => {
  const s = overview.value?.score ?? 100
  if (s >= 80) return '#67c23a'
  if (s >= 60) return '#e6a23c'
  return '#f56c6c'
})

const banned = computed(() => {
  const until = overview.value?.banned_until
  return until && new Date(until) > new Date() ? new Date(until).toLocaleString('zh-CN', { hour12: false }) : null
})

onMounted(async () => {
  const resp = await getCreditOverview()
  overview.value = resp.data
})
</script>

<template>
  <div class="page-card">
    <h2 style="margin-top: 0">个人中心</h2>

    <div class="grid2">
      <!-- 基本信息卡片 -->
      <div class="card">
        <h3>基本信息</h3>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="用户名">{{ auth.user?.username }}</el-descriptions-item>
          <el-descriptions-item label="姓名">{{ auth.user?.real_name }}</el-descriptions-item>
          <el-descriptions-item label="学号">{{ auth.user?.student_no || '—' }}</el-descriptions-item>
          <el-descriptions-item label="角色">
            <el-tag size="small" :type="auth.isAdmin ? 'danger' : 'primary'">
              {{ auth.isAdmin ? '管理员' : '学生' }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- 信用卡片 -->
      <div class="card">
        <h3>我的信用</h3>
        <div v-if="overview" class="credit-box">
          <el-progress type="dashboard" :percentage="overview.score" :color="scoreColor" :width="140">
            <template #default>
              <div class="score-num" :style="{ color: scoreColor }">{{ overview.score }}</div>
              <div class="score-label">信用分</div>
            </template>
          </el-progress>
          <el-alert
            v-if="banned"
            type="error" :closable="false" style="margin-top: 12px"
            :title="`信用分低于 60，禁止预约至 ${banned}`"
          />
          <el-alert
            v-else type="success" :closable="false" style="margin-top: 12px"
            title="信用状态正常，可正常预约"
          />
          <div class="rule-tip">
            规则：违约 −8 · 迟到取消 −2 · 按期履约 +1 · 低于 60 分禁约 3 天
          </div>
        </div>
      </div>
    </div>

    <!-- 信用流水 -->
    <h3>信用流水</h3>
    <el-table :data="overview?.logs || []" stripe max-height="400">
      <el-table-column label="时间" width="170">
        <template #default="{ row }">{{ new Date(row.created_at).toLocaleString('zh-CN', { hour12: false }) }}</template>
      </el-table-column>
      <el-table-column label="变动" width="90">
        <template #default="{ row }">
          <span :class="row.delta > 0 ? 'up' : 'down'">
            {{ row.delta > 0 ? '+' + row.delta : row.delta }}
          </span>
        </template>
      </el-table-column>
      <el-table-column label="事由" prop="reason" min-width="260" />
      <el-table-column label="关联预约" width="100">
        <template #default="{ row }">#{{ row.reservation_id ?? '—' }}</template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.grid2 {
  display: flex;
  gap: 16px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}
.card {
  flex: 1;
  min-width: 320px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 14px;
}
.credit-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 0;
}
.score-num {
  font-size: 30px;
  font-weight: 700;
}
.score-label {
  font-size: 12px;
  color: #909399;
}
.rule-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 10px;
}
.up {
  color: #67c23a;
  font-weight: 600;
}
.down {
  color: #f56c6c;
  font-weight: 600;
}
</style>
