<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import * as echarts from 'echarts'
import { getRooms } from '../api/room'
import { getHeatmap, getOverview } from '../api/stats'

const rooms = ref([])
const roomId = ref(null)
const date = ref(new Date().toISOString().slice(0, 10))

const heatEl = ref(null)
const trendEl = ref(null)
const peakEl = ref(null)
const topEl = ref(null)

const overview = ref(null)

let heatChart, trendChart, peakChart, topChart

async function loadHeatmap() {
  if (!roomId.value) return
  const resp = await getHeatmap(roomId.value, date.value)
  const d = resp.data
  if (!heatChart) heatChart = echarts.init(heatEl.value)

  // 组装 ECharts heatmap 数据: [x小时, y座位, value]
  const data = []
  d.seats.forEach((s, i) => {
    d.cells[i].forEach((v, j) => data.push([j, i, v]))
  })
  heatChart.setOption({
    tooltip: {
      position: 'top',
      formatter: (p) => `${d.seats[p.value[1]].seat_no} ${d.hours[p.value[0]]}<br/>占用：${p.value[2] ? '是' : '否'}`
    },
    grid: { top: 30, left: 70, right: 20, bottom: 60 },
    xAxis: {
      type: 'category', data: d.hours, splitArea: { show: true },
      axisLabel: { fontSize: 10 }
    },
    yAxis: {
      type: 'category',
      data: d.seats.map((s) => s.seat_no),
      splitArea: { show: true }, axisLabel: { fontSize: 9 }
    },
    visualMap: {
      min: 0, max: 1, calculable: false,
      orient: 'horizontal', left: 'center', bottom: 0,
      inRange: { color: ['#f5f7fa', '#f56c6c'] },
      text: ['占用', '空闲'], textStyle: { fontSize: 11 }
    },
    series: [{
      type: 'heatmap', data,
      label: { show: false },
      itemStyle: { borderColor: '#fff', borderWidth: 1 }
    }]
  })
}

async function loadOverview() {
  const resp = await getOverview()
  overview.value = resp.data
  const d = resp.data
  await nextTick()

  // 趋势折线
  if (!trendChart) trendChart = echarts.init(trendEl.value)
  trendChart.setOption({
    title: { text: '近14天使用率趋势', left: 'center', textStyle: { fontSize: 13 } },
    grid: { top: 40, left: 50, right: 20, bottom: 30 },
    tooltip: {
      trigger: 'axis',
      formatter: (ps) => {
        const p = ps[0]
        return `${p.name}<br/>使用率 ${(p.value * 100).toFixed(1)}%`
      }
    },
    xAxis: { type: 'category', data: d.trend_14.map((t) => t.date.slice(5)) },
    yAxis: { type: 'value', axisLabel: { formatter: '{value}%' } },
    series: [{
      type: 'line', smooth: true, data: d.trend_14.map((t) => (t.utilization * 100).toFixed(1) * 1),
      areaStyle: { opacity: 0.15 }, itemStyle: { color: '#409eff' }
    }]
  })

  // 高峰时段柱状
  if (!peakChart) peakChart = echarts.init(peakEl.value)
  peakChart.setOption({
    title: { text: '高峰时段分布(近14天)', left: 'center', textStyle: { fontSize: 13 } },
    grid: { top: 40, left: 50, right: 20, bottom: 30 },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: d.peak.map((p) => p.hour) },
    yAxis: { type: 'value', name: '预约数' },
    series: [{
      type: 'bar', data: d.peak.map((p) => p.count),
      itemStyle: { color: '#67c23a' }
    }]
  })

  // 热门座位条形
  if (!topChart) topChart = echarts.init(topEl.value)
  const tops = [...d.top_seats].reverse()
  topChart.setOption({
    title: { text: '热门座位 Top10', left: 'center', textStyle: { fontSize: 13 } },
    grid: { top: 40, left: 130, right: 30, bottom: 30 },
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    xAxis: { type: 'value', name: '占用小时' },
    yAxis: {
      type: 'category',
      data: tops.map((t) => `${t.room_name.slice(0, 3)} ${t.seat_no}`)
    },
    series: [{
      type: 'bar', data: tops.map((t) => t.hours),
      itemStyle: { color: '#e6a23c' }, label: { show: true, position: 'right' }
    }]
  })
}

async function init() {
  const resp = await getRooms()
  rooms.value = resp.data || []
  if (rooms.value.length) roomId.value = rooms.value[0].id
  await Promise.all([loadHeatmap(), loadOverview()])
}

watch([roomId, date], loadHeatmap)

onMounted(() => {
  init()
  window.addEventListener('resize', resizeAll)
})

function resizeAll() {
  heatChart?.resize(); trendChart?.resize(); peakChart?.resize(); topChart?.resize()
}
</script>

<template>
  <div class="page-card">
    <h2 style="margin-top: 0">热力图与数据分析</h2>

    <!-- 总览卡片 -->
    <div v-if="overview" class="stat-cards">
      <div class="stat-card">
        <div class="stat-num">{{ overview.total_seats }}</div>
        <div class="stat-label">总座位数</div>
      </div>
      <div class="stat-card">
        <div class="stat-num">{{ overview.active_today }}</div>
        <div class="stat-label">今日预约</div>
      </div>
      <div class="stat-card">
        <div class="stat-num">{{ overview.in_use_now }}</div>
        <div class="stat-label">当前使用中</div>
      </div>
      <div class="stat-card">
        <div class="stat-num">{{ (overview.today_utilization * 100).toFixed(1) }}%</div>
        <div class="stat-label">今日利用率</div>
      </div>
    </div>

    <!-- 热力图 -->
    <div class="heat-filter">
      <span>座位热力图</span>
      <el-select v-model="roomId" style="width: 180px">
        <el-option v-for="r in rooms" :key="r.id" :label="r.name" :value="r.id" />
      </el-select>
      <el-date-picker v-model="date" type="date" value-format="YYYY-MM-DD" style="width: 150px" />
      <span class="dim">颜色越红表示该座位在该小时越可能无空位</span>
    </div>
    <div ref="heatEl" class="chart chart-heat"></div>

    <!-- 统计图表 -->
    <div class="charts-row">
      <div ref="trendEl" class="chart"></div>
      <div ref="peakEl" class="chart"></div>
    </div>
    <div ref="topEl" class="chart"></div>
  </div>
</template>

<style scoped>
.stat-cards {
  display: flex;
  gap: 16px;
  margin-bottom: 18px;
  flex-wrap: wrap;
}
.stat-card {
  flex: 1;
  min-width: 140px;
  background: linear-gradient(135deg, #409eff22, #409eff0d);
  border: 1px solid #d9ecff;
  border-radius: 8px;
  padding: 14px;
  text-align: center;
}
.stat-num {
  font-size: 26px;
  font-weight: 700;
  color: #409eff;
}
.stat-label {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
.heat-filter {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  font-weight: 600;
}
.dim {
  font-size: 12px;
  color: #909399;
  font-weight: 400;
}
.chart {
  height: 320px;
  width: 100%;
}
.chart-heat {
  height: 420px;
  border: 1px dashed #ebeef5;
  border-radius: 6px;
  margin-bottom: 18px;
}
.charts-row {
  display: flex;
  gap: 16px;
}
.charts-row .chart {
  flex: 1;
  min-width: 0;
}
</style>
