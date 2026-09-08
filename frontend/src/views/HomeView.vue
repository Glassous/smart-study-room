<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { getRooms } from '../api/room'
import { getOverview } from '../api/stats'
import { getCreditOverview } from '../api/credit'
import { unreadCount } from '../api/notification'
import AppIcon from '../components/AppIcon.vue'

const router = useRouter()
const auth = useAuthStore()

const who = computed(() => auth.user?.real_name || auth.user?.username || '同学')
const roleText = computed(() => (auth.isAdmin ? '管理员' : '学生'))
const helloText = computed(() => auth.isAdmin ? '欢迎回来，请查看今日运营情况' : '今天也要高效学习')

const rooms = ref([])
const statsOverview = ref(null)
const creditOverview = ref(null)
const unread = ref(0)
let unreadTimer = null

const currentCredit = computed(() => creditOverview.value?.score ?? auth.user?.credit_score ?? 100)
const totalSeats = computed(() => statsOverview.value?.total_seats ?? 158)
const inUseNow = computed(() => statsOverview.value?.in_use_now ?? 0)
const availableSeats = computed(() => Math.max(0, totalSeats.value - inUseNow.value))
const roomCount = computed(() => rooms.value.length || 3)
const openHours = computed(() => {
  if (rooms.value.length && rooms.value[0].open_time && rooms.value[0].close_time) {
    return {
      open: rooms.value[0].open_time.slice(0, 5),
      close: rooms.value[0].close_time.slice(0, 5)
    }
  }
  return { open: '08:00', close: '22:00' }
})

async function refreshUnread() {
  if (!auth.isLoggedIn) return
  try {
    const resp = await unreadCount()
    unread.value = resp.data?.count || 0
  } catch (e) {
    unread.value = 0
  }
}

onMounted(async () => {
  try {
    const roomResp = await getRooms()
    rooms.value = roomResp.data || []
  } catch (e) {
    // 降级使用默认值
  }

  try {
    const statsResp = await getOverview()
    statsOverview.value = statsResp.data
  } catch (e) {
    // 降级使用默认值
  }

  if (auth.isStudent) {
    try {
      const credResp = await getCreditOverview()
      creditOverview.value = credResp.data
    } catch (e) {
      // 降级使用默认值
    }
  }

  refreshUnread()
  unreadTimer = setInterval(refreshUnread, 30000)
})

onUnmounted(() => {
  if (unreadTimer) {
    clearInterval(unreadTimer)
    unreadTimer = null
  }
})

const shortcuts = computed(() => auth.isAdmin
  ? [
      { title: '热力图统计', desc: '座位利用率 · 高峰时段', icon: 'analytics', to: '/analytics', tone: 'warm' },
      { title: '管理端', desc: '自习室 · 座位 · 用户', icon: 'admin', to: '/admin', tone: 'primary' },
      {
        title: '消息中心',
        desc: unread.value > 0 ? `${unread.value} 条未读通知 · 点击查看` : '通知公告 · 异常提醒',
        icon: unread.value > 0 ? 'notification' : 'notification-empty',
        to: '/notifications',
        tone: 'success'
      },
      { title: '个人中心', desc: '查看管理员账户信息', icon: 'profile', to: '/profile', tone: 'violet' }
    ]
  : [
      { title: '座位预约', desc: '手动选座 · 智能分配', icon: 'booking', to: '/booking', tone: 'primary' },
      { title: '我的预约', desc: '签到 · 临时离开 · 签退', icon: 'reservations', to: '/mine', tone: 'success' },
      { title: '我的候补', desc: '查看排队与递补状态', icon: 'waitlist', to: '/waitlist', tone: 'warm' },
      {
        title: '消息中心',
        desc: unread.value > 0 ? `${unread.value} 条未读消息 · 点击查看` : '预约结果 · 违约警告 · 递补',
        icon: unread.value > 0 ? 'notification' : 'notification-empty',
        to: '/notifications',
        tone: 'violet'
      }
    ])
const cardToneClass = {
  primary: 'tone-primary',
  success: 'tone-success',
  warm: 'tone-warm',
  violet: 'tone-violet'
}
</script>

<template>
  <div class="page-view">
    <header class="view-heading" data-page-title><h1>首页</h1></header>
    <!-- 首页：页面级双栏 -->
    <div class="split home-split">
    <!-- 左栏：欢迎 + 账户 KPI + 快速贴士 -->
    <div class="split-left">
      <section class="card welcome-card">
        <div class="hello">
          <div class="hello-avatar">{{ who.slice(0, 1) }}</div>
          <div>
            <div class="hello-role">
              <el-tag size="small" effect="plain" :type="auth.isAdmin ? 'danger' : 'primary'">
                {{ roleText }}
              </el-tag>
            </div>
            <h2 class="hello-name">你好，{{ who }}</h2>
            <p class="hello-sub">{{ helloText }}</p>
          </div>
        </div>
      </section>

      <section v-if="auth.isStudent" class="card overview-card">
        <div class="card-title-row">
          <h3>账户概览</h3>
        </div>
        <div class="kpi-grid">
          <div class="kpi">
            <div class="kpi-num">{{ currentCredit }}</div>
            <div class="kpi-label">现有信用分</div>
          </div>
          <div class="kpi">
            <div class="kpi-num">{{ roomCount }}</div>
            <div class="kpi-label">自习室可用</div>
          </div>
          <div class="kpi">
            <div class="kpi-num">
              {{ availableSeats }}<span class="kpi-slash">/</span><span class="kpi-total">{{ totalSeats }}</span>
            </div>
            <div class="kpi-label">可用座位</div>
          </div>
          <div class="kpi">
            <div class="kpi-num">{{ openHours.open }}<br/><span style="font-size:11px;font-weight:500">~ {{ openHours.close }}</span></div>
            <div class="kpi-label">今日开放</div>
          </div>
        </div>
      </section>

      <section v-else class="card overview-card">
        <div class="card-title-row">
          <h3>管理概览</h3>
        </div>
        <div class="kpi-grid">
          <div class="kpi">
            <div class="kpi-num">{{ roomCount }}</div>
            <div class="kpi-label">自习室</div>
          </div>
          <div class="kpi">
            <div class="kpi-num">
              {{ availableSeats }}<span class="kpi-slash">/</span><span class="kpi-total">{{ totalSeats }}</span>
            </div>
            <div class="kpi-label">可用 / 总座位</div>
          </div>
          <div class="kpi">
            <div class="kpi-num">2</div>
            <div class="kpi-label">管理模块</div>
          </div>
          <div class="kpi">
            <div class="kpi-num">{{ openHours.open }}<br/><span style="font-size:11px;font-weight:500">~ {{ openHours.close }}</span></div>
            <div class="kpi-label">今日开放</div>
          </div>
        </div>
      </section>

      <section v-if="auth.isStudent" class="card tips-card">
        <h3>使用小贴士</h3>
        <ul class="tips">
          <li>距开始不足 30 分钟取消预约将扣 <b>2 分</b>信用分</li>
          <li>超时未签到会自动记为违约，扣 <b>8 分</b>，并释放座位</li>
          <li>满座时段可加入 <b>候补</b>，空位释放时按序自动递补</li>
          <li>信用分低于 60，<b>3 天内</b>无法发起新预约</li>
        </ul>
      </section>

      <section v-else class="card tips-card">
        <h3>管理提示</h3>
        <ul class="tips">
          <li>定期检查自习室开放时间与座位规模是否准确</li>
          <li>维护中的座位不会开放给普通用户预约</li>
          <li>可在用户管理中启用或禁用普通用户账号</li>
          <li>通过热力图了解座位利用率和高峰时段</li>
        </ul>
      </section>
    </div>

    <!-- 右栏：功能入口卡 -->
    <div class="split-right">
      <section class="card shortcuts-card">
        <div class="card-title-row">
          <h3>快速入口</h3>
          <span class="muted">点击卡片直接进入对应功能</span>
        </div>
        <div class="shortcut-grid">
          <div
            v-for="s in shortcuts"
            :key="s.to"
            class="shortcut"
            :class="cardToneClass[s.tone]"
            @click="router.push(s.to)"
          >
            <div class="shortcut-top">
              <div class="shortcut-icon"><AppIcon :name="s.icon" :size="32" /></div>
              <div class="chev"><AppIcon name="chevron-right" :size="20" /></div>
            </div>
            <div class="shortcut-body">
              <div class="shortcut-title">{{ s.title }}</div>
              <div class="shortcut-desc">{{ s.desc }}</div>
            </div>
          </div>
        </div>
      </section>

      <section class="card intro-card">
        <div class="card-title-row">
          <h3>系统介绍</h3>
        </div>
        <div v-if="auth.isStudent" class="intro-grid">
          <div class="intro-cell">
            <div class="intro-label">在线预约</div>
            <div class="intro-text">平面图可视化选座，时段冲突自动检测，支持签到 / 临时离开 / 签退全生命周期。</div>
          </div>
          <div class="intro-cell">
            <div class="intro-label">智能分配</div>
            <div class="intro-text">按区域、电源、靠窗偏好加权评分，一键推荐最优座位或直接自动下单。</div>
          </div>
          <div class="intro-cell">
            <div class="intro-label">候补递补</div>
            <div class="intro-text">满座时加入候补队列，空位释放后按排队顺序自动递补并发送通知。</div>
          </div>
          <div class="intro-cell">
            <div class="intro-label">信用治理</div>
            <div class="intro-text">违约扣分、履约加分、低分限约，配合满座候补递补机制，公平利用座位资源。</div>
          </div>
        </div>
        <div v-else class="intro-grid">
          <div class="intro-cell">
            <div class="intro-label">资源管理</div>
            <div class="intro-text">维护自习室名称、位置、开放时间和座位规模，统一管理可预约资源。</div>
          </div>
          <div class="intro-cell">
            <div class="intro-label">座位维护</div>
            <div class="intro-text">批量生成座位并维护座位属性与状态，确保不可用座位及时下线。</div>
          </div>
          <div class="intro-cell">
            <div class="intro-label">用户管理</div>
            <div class="intro-text">查看普通用户信息与信用状态，并按需启用或禁用用户账号。</div>
          </div>
          <div class="intro-cell">
            <div class="intro-label">运营分析</div>
            <div class="intro-text">查看座位利用率和时段热力分布，为资源配置与开放安排提供参考。</div>
          </div>
        </div>
      </section>
    </div>
    </div>
  </div>
</template>

<style scoped>
.home-split {
  grid-template-columns: 320px 1fr;
}

/* 欢迎卡 */
.welcome-card {
  background:
    linear-gradient(145deg, #eef3f9, #f7f9fc);
}
.hello {
  display: flex;
  gap: 14px;
  align-items: center;
}
.hello-avatar {
  width: 54px;
  height: 54px;
  border-radius: 14px;
  background: linear-gradient(145deg, #8ea6c4, #5f7ea3);
  color: #fff;
  display: grid;
  place-items: center;
  font-weight: 700;
  font-size: 22px;
  box-shadow: inset 0 1px 0 rgba(255,255,255,.25), 0 8px 18px rgba(95,126,163,.22);
}
.hello-role { margin-bottom: 2px; }
.hello-name {
  margin: 4px 0 2px;
  font-size: 20px;
}
.hello-sub {
  margin: 0;
  color: #828c9d;
  font-size: 13px;
}

/* KPI 中的可用/总数格式美化 */
.kpi-slash {
  font-size: 14px;
  margin: 0 1px;
  color: #8c9bb0;
  font-weight: 500;
}
.kpi-total {
  font-size: 14px;
  color: #7b889b;
  font-weight: 500;
}

/* 贴士 */
.tips {
  margin: 0;
  padding-left: 18px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #4b5567;
  font-size: 13px;
  line-height: 1.6;
}
.tips b {
  color: #456388;
  font-weight: 600;
}

/* 快速入口卡片 */
.shortcuts-card {
  display: flex;
  flex-direction: column;
}

/* PC 下快速入口：4 个占满高度的容器 */
.shortcut-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  min-height: 0;
}

.shortcut {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  height: 100%;
  padding: 24px 20px 22px;
  border-radius: 16px;
  cursor: pointer;
  border: 1px solid #eef1f6;
  background: #fbfcfe;
  transition: transform .22s ease, box-shadow .22s ease, border-color .22s ease, background-color .22s ease;
  box-sizing: border-box;
}

.shortcut:hover {
  transform: translateY(-4px);
  box-shadow: 0 14px 30px rgba(108, 128, 160, 0.16);
  border-color: rgba(95, 126, 163, 0.28);
  background: #fff;
}

.shortcut-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.shortcut-icon {
  width: 58px;
  height: 58px;
  border-radius: 16px;
  display: grid;
  place-items: center;
  background: #fff;
  box-shadow: 0 1px 2px rgba(108, 128, 160, 0.06), 0 4px 12px rgba(108, 128, 160, 0.08);
  flex-shrink: 0;
  transition: transform .2s ease;
}

.shortcut-icon .app-icon {
  width: 32px;
  height: 32px;
}

.shortcut:hover .shortcut-icon {
  transform: scale(1.06);
}

.shortcut-body {
  min-width: 0;
}

/* 大幅加大标题字号，使其更加突出显眼 */
.shortcut-title {
  font-size: 24px;
  font-weight: 800;
  color: #202836;
  line-height: 1.3;
  margin-bottom: 8px;
  letter-spacing: -0.3px;
  transition: color .2s ease;
}

.shortcut:hover .shortcut-title {
  color: #3b5f88;
}

.shortcut-desc {
  font-size: 13px;
  color: #828c9d;
  line-height: 1.5;
}

.shortcut .chev {
  color: #a2abb9;
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border-radius: 9px;
  background: rgba(0, 0, 0, 0.02);
  transition: transform .2s ease, color .2s ease, background .2s ease;
}

.shortcut:hover .chev {
  transform: translateX(4px);
  color: #456388;
  background: rgba(69, 99, 136, 0.08);
}

.tone-primary .shortcut-icon { background: linear-gradient(145deg, #dce8f7, #eff4fb); }
.tone-success .shortcut-icon { background: linear-gradient(145deg, #dff1e0, #eff9f0); }
.tone-warm    .shortcut-icon { background: linear-gradient(145deg, #f6e7d0, #fbf2e3); }
.tone-violet  .shortcut-icon { background: linear-gradient(145deg, #e4ddf5, #f1ecfa); }

/* 系统介绍 */
.intro-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}
.intro-cell {
  border-radius: 12px;
  padding: 14px 16px;
  background: linear-gradient(150deg, #f7f9fc, #eef2f8);
  border: 1px solid #eef1f6;
}
.intro-label {
  font-size: 13px;
  font-weight: 700;
  color: #456388;
  margin-bottom: 4px;
  letter-spacing: .3px;
}
.intro-text {
  font-size: 12.5px;
  color: #5b6576;
  line-height: 1.65;
}

/* ---------- 响应式断点：全部组件占满 100% 宽度，快速入口移至问候卡片下一行 ---------- */
@media (max-width: 1024px) {
  .home-split {
    display: flex !important;
    flex-direction: column !important;
    gap: 14px !important;
    width: 100% !important;
  }
  .home-split .split-left,
  .home-split .split-right {
    display: contents !important;
  }
  .welcome-card { order: 1; width: 100%; }
  .shortcuts-card { order: 2; width: 100%; }
  .overview-card { order: 3; width: 100%; }
  .tips-card { order: 4; width: 100%; }
  .intro-card { order: 5; width: 100%; }

  /* 快速入口单列全宽 */
  .shortcut-grid {
    grid-template-columns: 1fr;
    gap: 12px;
    width: 100%;
  }
  .shortcut {
    width: 100%;
    min-height: 92px;
    padding: 18px 22px;
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 18px;
  }
  .shortcut-top {
    display: contents;
  }
  .shortcut-icon {
    order: 1;
    width: 56px;
    height: 56px;
    border-radius: 16px;
    flex-shrink: 0;
  }
  .shortcut-icon .app-icon {
    width: 32px;
    height: 32px;
  }
  .shortcut-body {
    order: 2;
    flex: 1;
    min-width: 0;
  }
  .shortcut-title {
    font-size: 22px;
    font-weight: 800;
    margin-bottom: 4px;
    line-height: 1.3;
  }
  .shortcut-desc {
    font-size: 13px;
    color: #828c9d;
  }
  .shortcut .chev {
    order: 3;
    flex-shrink: 0;
    width: 34px;
    height: 34px;
    border-radius: 10px;
  }

  /* 账户概览 KPI 单列全宽行 */
  .kpi-grid {
    grid-template-columns: 1fr;
    gap: 10px;
    width: 100%;
  }
  .kpi {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 14px 18px;
    box-sizing: border-box;
  }
  .kpi-label {
    margin-top: 0;
    font-size: 13.5px;
    color: #788599;
  }
  .kpi-num {
    font-size: 22px;
    font-weight: 700;
    text-align: right;
  }

  /* 系统介绍单列全宽 */
  .intro-grid {
    grid-template-columns: 1fr;
    gap: 12px;
    width: 100%;
  }
  .intro-cell {
    width: 100%;
    box-sizing: border-box;
  }
}

@media (max-width: 560px) {
  .shortcut {
    min-height: 86px;
    padding: 16px 16px;
    gap: 14px;
  }
  .shortcut-icon {
    width: 48px;
    height: 48px;
    border-radius: 13px;
  }
  .shortcut-icon .app-icon {
    width: 26px;
    height: 26px;
  }
  .shortcut-title {
    font-size: 19px;
    font-weight: 800;
    margin-bottom: 3px;
  }
  .shortcut-desc {
    font-size: 12px;
  }
  .shortcut .chev {
    width: 30px;
    height: 30px;
  }
  .welcome-card .hello {
    align-items: flex-start;
  }
}
</style>
