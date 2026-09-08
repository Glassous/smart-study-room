<script setup>
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { createReservation } from '../api/reservation'
import { createAIConversation, deleteAIConversation, listAIConversations, listAIMessages, streamAIChat } from '../api/ai'

const emit = defineEmits(['reservation-created'])
const open = ref(false)
const showHistory = ref(false)
const conversations = ref([])
const conversationId = ref(0)
const messages = ref([])
const input = ref('')
const sending = ref(false)
const messageList = ref(null)
const lastPrompt = ref('')
const textareaRef = ref(null)
let controller = null

const quickPrompts = ['我的状态', '未读通知', '帮我选座']
const zoneNames = { quiet: '静音区', regular: '普通区', discussion: '研讨区', computer: '机房区' }

function adjustTextareaHeight() {
  const el = textareaRef.value
  if (!el) return
  const prevScrollTop = el.scrollTop
  el.style.height = 'auto'
  const minHeight = 34
  const maxHeight = 120
  const scrollHeight = el.scrollHeight

  if (scrollHeight <= minHeight) {
    el.style.height = `${minHeight}px`
    el.style.overflowY = 'hidden'
  } else if (scrollHeight >= maxHeight) {
    el.style.height = `${maxHeight}px`
    el.style.overflowY = 'auto'
    el.scrollTop = prevScrollTop
  } else {
    el.style.height = `${scrollHeight}px`
    el.style.overflowY = 'hidden'
  }
}

async function loadConversations() {
  const response = await listAIConversations()
  conversations.value = response.data || []
}
async function selectConversation(item) {
  conversationId.value = item.id
  const response = await listAIMessages(item.id)
  messages.value = (response.data || []).filter((m) => m.role !== 'system' && m.role !== 'tool')
  showHistory.value = false
  scrollBottom()
}
function newConversation() {
  conversationId.value = 0
  messages.value = []
  showHistory.value = false
}
async function removeConversation(item) {
  try { await ElMessageBox.confirm(`删除“${item.title}”及其消息？`, '删除会话', { type: 'warning' }) } catch { return }
  await deleteAIConversation(item.id)
  if (conversationId.value === item.id) newConversation()
  await loadConversations()
}
async function scrollBottom() {
  await nextTick()
  if (messageList.value) messageList.value.scrollTop = messageList.value.scrollHeight
}
async function send(text = input.value) {
  const prompt = String(text || '').trim()
  if (!prompt || sending.value) return
  lastPrompt.value = prompt
  input.value = ''
  messages.value.push({ role: 'user', content: prompt, status: 'complete' })
  const assistant = { role: 'assistant', content: '', recommendations: [], status: 'streaming' }
  messages.value.push(assistant)
  sending.value = true
  controller = new AbortController()
  scrollBottom()
  try {
    await streamAIChat({ conversation_id: conversationId.value || undefined, message: prompt }, (event, data) => {
      if (event === 'meta') conversationId.value = data.conversation_id
      if (event === 'delta') assistant.content += data.content || ''
      if (event === 'card') assistant.recommendations = data.recommendations || []
      if (event === 'done') assistant.status = 'complete'
      if (event === 'error') throw new Error(data.message || 'AI 暂时不可用')
      scrollBottom()
    }, controller.signal)
    await loadConversations()
  } catch (error) {
    assistant.status = 'failed'
    assistant.error = error.name === 'AbortError' ? '回复已停止' : error.message
  } finally {
    sending.value = false
    controller = null
    scrollBottom()
  }
}
const bookingRecKey = ref('')

function stop() { controller?.abort() }
async function book(rec) {
  const recKey = `${rec.seat_id}-${rec.date}-${rec.start_time}`
  try {
    await ElMessageBox.confirm(
      `确认预约 ${rec.room_name} ${rec.seat_no}？\n${rec.date} ${rec.start_time}–${rec.end_time}`,
      '最终预约确认', { confirmButtonText: '确认预约', cancelButtonText: '再想想' }
    )
  } catch { return }
  bookingRecKey.value = recKey
  try {
    const response = await createReservation({ seat_id: rec.seat_id, date: rec.date, start_time: rec.start_time, end_time: rec.end_time })
    ElMessage.success(`预约成功：${response.data.room_name} ${response.data.seat_no}`)
    messages.value.push({ role: 'assistant', content: `已预约 ${response.data.room_name} ${response.data.seat_no}，请按时签到。`, status: 'local' })
    emit('reservation-created', response.data)

    // 预约成功后，自动禁用所有推荐卡片的选择按钮，避免重复与多余预约
    for (const msg of messages.value) {
      if (msg.recommendations?.length) {
        for (const r of msg.recommendations) {
          r.disabled = true
        }
      }
    }
    rec.booked = true
  } catch {
    ElMessage.info('该推荐可能已失效，请让助手重新推荐。')
  } finally {
    bookingRecKey.value = ''
  }
}
function onKeydown(event) {
  if (event.key === 'Escape' && open.value) open.value = false
}
watch(input, async () => {
  await nextTick()
  adjustTextareaHeight()
})
watch(showHistory, async (val) => {
  if (!val) {
    await nextTick()
    adjustTextareaHeight()
  }
})
watch(open, (value) => {
  if (value) {
    loadConversations()
    scrollBottom()
    nextTick(adjustTextareaHeight)
  }
})
onMounted(() => {
  window.addEventListener('keydown', onKeydown)
  nextTick(adjustTextareaHeight)
})
onUnmounted(() => { window.removeEventListener('keydown', onKeydown); controller?.abort() })
</script>

<template>
  <button class="ai-fab" type="button" aria-label="打开 AI 助手" :aria-expanded="open" @click="open = !open">
    <span class="fab-icon" aria-hidden="true">✦</span><span class="fab-label">AI 助手</span>
  </button>
  <Transition name="sheet">
    <div v-if="open" class="ai-layer">
      <button class="ai-backdrop" aria-label="关闭 AI 助手" @click="open = false" />
      <section class="ai-sheet" role="dialog" aria-modal="true" aria-label="AI 学习助手">
        <div class="drag-handle" />
        <header class="ai-header">
          <div><b>AI 学习助手</b><small>信息查询与选座建议</small></div>
          <div class="header-actions">
            <button type="button" @click="showHistory = !showHistory">历史</button>
            <button type="button" @click="newConversation">新对话</button>
            <button type="button" aria-label="关闭" @click="open = false">×</button>
          </div>
        </header>
        <div v-if="showHistory" class="history-panel">
          <div v-if="!conversations.length" class="empty">还没有历史会话</div>
          <div v-for="item in conversations" :key="item.id" class="history-item" :class="{ active: item.id === conversationId }">
            <button type="button" @click="selectConversation(item)"><span>{{ item.title }}</span><small>{{ new Date(item.updated_at).toLocaleString() }}</small></button>
            <button class="delete" type="button" aria-label="删除会话" @click="removeConversation(item)">×</button>
          </div>
        </div>
        <div v-else ref="messageList" class="message-list">
          <div v-if="!messages.length" class="welcome">
            <span class="welcome-icon">✦</span><b>需要我帮什么？</b><p>我可以查询你的状态、通知，或根据偏好推荐座位。</p>
          </div>
          <div v-for="(message, index) in messages" :key="index" class="message" :class="message.role">
            <div v-if="message.content" class="bubble">{{ message.content }}</div>
            <div v-if="message.recommendations?.length" class="recommendations">
              <article
                v-for="rec in message.recommendations"
                :key="`${rec.seat_id}-${rec.date}-${rec.start_time}`"
                class="seat-rec"
                :class="{ 'is-booked': rec.booked, 'is-disabled': rec.disabled && !rec.booked }"
              >
                <div class="rec-top"><b>{{ rec.room_name }} · {{ rec.seat_no }}</b><span>{{ zoneNames[rec.zone] || rec.zone }}</span></div>
                <div class="rec-time">{{ rec.date }}　{{ rec.start_time }}–{{ rec.end_time }}</div>
                <div class="rec-tags"><span v-if="rec.has_power">有电源</span><span v-if="rec.near_window">靠窗</span></div>
                <p>{{ rec.reason }}</p>
                <button
                  type="button"
                  :disabled="Boolean(bookingRecKey || rec.disabled || rec.booked)"
                  @click="book(rec)"
                >
                  <template v-if="bookingRecKey === `${rec.seat_id}-${rec.date}-${rec.start_time}`">预约中…</template>
                  <template v-else-if="rec.booked">✓ 已选择此座位</template>
                  <template v-else-if="rec.disabled">已选择其他座位</template>
                  <template v-else>选择此座位</template>
                </button>
              </article>
            </div>
            <div v-if="message.status === 'streaming' && !message.content && !message.recommendations?.length" class="typing"><i/><i/><i/></div>
            <div v-if="message.status === 'failed'" class="message-error">{{ message.error }} <button type="button" @click="send(lastPrompt)">重试</button></div>
          </div>
        </div>
        <div v-if="!showHistory" class="composer">
          <div class="quick-prompts"><button v-for="q in quickPrompts" :key="q" type="button" :disabled="sending" @click="send(q)">{{ q }}</button></div>
          <div class="input-row">
            <textarea
              ref="textareaRef"
              v-model="input"
              maxlength="1000"
              rows="1"
              placeholder="输入问题…"
              @input="adjustTextareaHeight"
              @keydown.enter.exact.prevent="send()"
            />
            <button v-if="sending" class="send stop" type="button" aria-label="停止生成" @click="stop">■</button>
            <button v-else class="send" type="button" :disabled="!input.trim()" aria-label="发送消息" @click="send()">↑</button>
          </div>
        </div>
      </section>
    </div>
  </Transition>
</template>

<style scoped>
.ai-fab{position:fixed;right:24px;bottom:24px;z-index:35;border:1px solid rgba(255,255,255,.25);border-radius:999px;padding:14px 22px;font-size:15px;font-weight:650;background:linear-gradient(135deg,#456388,#6f8fb5);color:#fff;display:flex;gap:9px;align-items:center;box-shadow:none;cursor:pointer;transition:transform .2s,opacity .2s,background-color .2s}
.ai-fab:hover{transform:translateY(-2px);opacity:.96;box-shadow:none}
.ai-fab:active{transform:scale(.97)}
.ai-fab .fab-icon{font-size:18px;line-height:1}
.ai-layer{position:fixed;z-index:80;inset:0;pointer-events:none}
.ai-backdrop{position:absolute;inset:0;border:0;background:rgba(20,29,43,.25);pointer-events:auto}
.ai-sheet{position:absolute;right:20px;bottom:20px;width:min(460px,calc(100vw - 40px));height:min(700px,calc(100vh - 40px));background:#fff;border-radius:20px;border:1px solid #d4dce6;box-shadow:none;overflow:hidden;display:flex;flex-direction:column;pointer-events:auto}
.drag-handle{display:none}
.ai-header{padding:18px 18px 14px;border-bottom:1px solid #e8ecf2;display:flex;justify-content:space-between;align-items:center}
.ai-header div:first-child{display:flex;flex-direction:column;gap:2px}
.ai-header small{color:#8a96a8}
.header-actions{display:flex;gap:4px}
.header-actions button,.quick-prompts button{border:1px solid #dfe5ed;background:#f7f9fc;color:#52627a;border-radius:9px;padding:7px 9px;cursor:pointer}
.header-actions button:last-child{font-size:20px;padding:2px 9px}
.message-list{flex:1;overflow:auto;padding:18px;background:#f6f8fb;scrollbar-width:thin;scrollbar-color:rgba(146,160,179,.4) transparent}
.message-list::-webkit-scrollbar{width:6px}
.message-list::-webkit-scrollbar-track{background:transparent}
.message-list::-webkit-scrollbar-thumb{background-color:rgba(146,160,179,.4);border-radius:999px}
.message-list::-webkit-scrollbar-thumb:hover{background-color:rgba(82,111,149,.7)}
.welcome{text-align:center;color:#6d7889;padding:64px 24px}.welcome-icon{display:block;font-size:30px;color:#5d7da5}.welcome b{display:block;color:#334155;margin:10px}.welcome p{font-size:13px}.message{display:flex;flex-direction:column;margin:10px 0;align-items:flex-start}.message.user{align-items:flex-end}.bubble{max-width:86%;white-space:pre-wrap;line-height:1.55;padding:10px 13px;border-radius:14px;background:#fff;border:1px solid #e6eaf0;color:#334155}.user .bubble{background:#526f95;color:#fff;border:0}.typing{padding:12px;background:#fff;border-radius:14px}.typing i{display:inline-block;width:6px;height:6px;margin:0 2px;border-radius:50%;background:#92a0b3;animation:pulse 1s infinite}.typing i:nth-child(2){animation-delay:.15s}.typing i:nth-child(3){animation-delay:.3s}@keyframes pulse{50%{opacity:.25;transform:translateY(-2px)}}
.recommendations{width:100%;display:grid;gap:10px;margin-top:8px}
.seat-rec{background:#fff;border:1px solid #dce4ee;border-radius:14px;padding:13px;box-shadow:0 3px 10px rgba(71,92,120,.06);transition:all .2s}
.seat-rec.is-booked{border-color:#8ec9ab;background:#f6fbf8}
.seat-rec.is-disabled{opacity:.75}
.rec-top{display:flex;justify-content:space-between;gap:8px}
.rec-top span,.rec-tags span{font-size:12px;padding:2px 7px;background:#edf2f8;color:#526d8f;border-radius:999px}
.rec-time{font-size:13px;color:#68758a;margin:6px 0}
.rec-tags{display:flex;gap:5px}
.seat-rec p{font-size:13px;color:#667085;margin:7px 0}
.seat-rec button{width:100%;border:0;border-radius:9px;padding:8px;background:#526f95;color:#fff;font-weight:600;cursor:pointer;transition:all .2s}
.seat-rec button:disabled{background:#cbd5e1;color:#64748b;cursor:not-allowed;opacity:.9}
.seat-rec.is-booked button:disabled{background:#3f7c5f;color:#fff;opacity:1}
.message-error{font-size:12px;color:#c45656}.message-error button{border:0;background:none;color:#456388;cursor:pointer}
.composer{border-top:1px solid #e5e9ef;padding:11px 14px 14px;background:#fff}
.quick-prompts{display:flex;gap:7px;overflow-x:auto;margin-bottom:9px;scrollbar-width:none}
.quick-prompts::-webkit-scrollbar{display:none}
.quick-prompts button{white-space:nowrap;padding:6px 10px;font-size:12px}
.input-row{display:flex;gap:8px;align-items:flex-end;background:#f5f7fa;border:1px solid #dfe5ec;border-radius:13px;padding:6px 7px;transition:border-color .2s,background-color .2s,box-shadow .2s}
.input-row:focus-within{border-color:#6f8fb5;background:#fff;box-shadow:0 0 0 3px rgba(111,143,181,.15)}
.input-row textarea{flex:1;resize:none;border:0;outline:0;margin:0;background:transparent;font-family:inherit;font-size:14px;line-height:20px;height:34px;min-height:34px;max-height:120px;box-sizing:border-box;padding:7px 8px;overflow-y:hidden;color:#334155;scrollbar-width:thin;scrollbar-color:rgba(146,160,179,.5) transparent}
.input-row textarea::placeholder{color:#9aa4b2}
.input-row textarea::-webkit-scrollbar{width:5px}
.input-row textarea::-webkit-scrollbar-track{background:transparent}
.input-row textarea::-webkit-scrollbar-thumb{background-color:rgba(146,160,179,.45);border-radius:999px;transition:background-color .2s}
.input-row textarea::-webkit-scrollbar-thumb:hover{background-color:rgba(82,111,149,.8)}
.send{width:34px;height:34px;flex-shrink:0;display:inline-flex;align-items:center;justify-content:center;border:0;border-radius:10px;background:#526f95;color:#fff;font-size:15px;font-weight:600;cursor:pointer;transition:background-color .2s,opacity .2s,transform .1s}
.send:hover:not(:disabled){background:#415978}
.send:active:not(:disabled){transform:scale(.96)}
.send:disabled{opacity:.35;cursor:not-allowed}
.send.stop{background:#7b8798}
.send.stop:hover{background:#637082}
.history-panel{flex:1;overflow:auto;padding:12px;scrollbar-width:thin;scrollbar-color:rgba(146,160,179,.4) transparent}
.history-panel::-webkit-scrollbar{width:6px}
.history-panel::-webkit-scrollbar-track{background:transparent}
.history-panel::-webkit-scrollbar-thumb{background-color:rgba(146,160,179,.4);border-radius:999px}
.history-panel::-webkit-scrollbar-thumb:hover{background-color:rgba(82,111,149,.7)}
.history-item{display:grid;grid-template-columns:1fr auto;border-radius:11px}.history-item.active{background:#edf2f8}.history-item>button:first-child{border:0;background:none;text-align:left;padding:11px;display:flex;flex-direction:column;gap:3px;min-width:0;cursor:pointer}.history-item span{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.history-item small{color:#99a3b2}.history-item .delete{border:0;background:none;color:#a4acb9;padding:10px;cursor:pointer}.empty{text-align:center;color:#9aa4b2;padding:40px}.sheet-enter-active,.sheet-leave-active{transition:opacity .2s}.sheet-enter-from,.sheet-leave-to{opacity:0}.sheet-enter-active .ai-sheet,.sheet-leave-active .ai-sheet{transition:transform .22s,opacity .22s}.sheet-enter-from .ai-sheet,.sheet-leave-to .ai-sheet{transform:translateY(18px);opacity:0}
@media(max-width:720px){
  .ai-fab{right:18px;bottom:36px;width:48px;height:48px;min-width:48px;min-height:48px;padding:0;aspect-ratio:1/1;border-radius:50%;display:flex;align-items:center;justify-content:center;box-shadow:none;border:1.5px solid rgba(255,255,255,.35)}
  .fab-label{display:none}
  .ai-fab .fab-icon{font-size:20px;line-height:1}
  .ai-sheet{inset:auto 0 0;width:100%;height:min(82vh,720px);border-radius:20px 20px 0 0;border:0;border-top:1px solid #d4dce6;box-shadow:none}
  .drag-handle{display:block;width:42px;height:4px;border-radius:999px;background:#d5dbe4;margin:8px auto 0}
  .ai-header{padding-top:10px}
  .ai-backdrop{background:rgba(20,29,43,.35)}
}
</style>
