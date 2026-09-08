<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as THREE from 'three'
import { OrbitControls } from 'three/addons/controls/OrbitControls.js'
import { gsap } from 'gsap'
import { buildRoom } from './models'
import { seatType, seatState, stateLabels, zoneName } from './seatPresentation'
import { useThemeStore } from '../../stores/theme'

const props = defineProps({ room: Object, seats: Array, selectedId: Number, view: String, blocked: Boolean })
const emit = defineEmits(['select', 'error', 'ready', 'manual-view'])
const host = ref(null)
const tip = ref(null)
const theme = useThemeStore()
let renderer, scene, camera, controls, model, observer, frame = 0, cameraTween, targetTween
let down = null, disposed = false, pointerCount = 0
const ray = new THREE.Raycaster()
const pointer = new THREE.Vector2()
const reduced = () => window.matchMedia('(prefers-reduced-motion: reduce)').matches
function invalidate() {
  if (disposed || frame) return
  frame = requestAnimationFrame(() => {
    frame = 0
    if (!renderer || !model) return
    controls.update(); model.updateWalls(camera); renderer.render(scene, camera)
  })
}
function resize() {
  if (!renderer || !host.value) return
  const { width, height } = host.value.getBoundingClientRect()
  renderer.setSize(Math.max(width, 1), Math.max(height, 1))
  camera.aspect = width / Math.max(height, 1); camera.updateProjectionMatrix(); invalidate()
}
function setView(view = 'overview', animate = true) {
  if (!model) return
  cameraTween?.kill(); targetTween?.kill()
  const halfFov = THREE.MathUtils.degToRad(camera.fov / 2)
  const span = Math.hypot(model.width, model.depth)
  const distance = span / (2 * Math.tan(halfFov) * Math.min(1, camera.aspect)) * 1.14
  const direction = view === 'top' ? new THREE.Vector3(0, 1, .001) : new THREE.Vector3(.8, 1.05, 1).normalize()
  const destination = direction.multiplyScalar(distance)
  controls.maxDistance = distance * 2.5
  camera.far = Math.max(500, distance * 5); camera.updateProjectionMatrix()
  const duration = animate && !reduced() ? .7 : 0
  controls.enabled = false
  cameraTween = gsap.to(camera.position, { x: destination.x, y: destination.y, z: destination.z, duration, ease: 'power2.inOut', onUpdate: invalidate, onComplete: () => { controls.enabled = !props.blocked; invalidate() } })
  targetTween = gsap.to(controls.target, { x: 0, y: 0, z: 0, duration, ease: 'power2.inOut', onUpdate: invalidate })
}
function rebuild() {
  if (!renderer) return
  if (model) { scene.remove(model.root); model.dispose() }
  model = buildRoom(props.room, props.seats, theme.isDark)
  scene.add(model.root)
  scene.background = new THREE.Color(theme.isDark ? '#1c2632' : '#f1f3f0')
  model.updateSelection(props.selectedId)
  tip.value = null; setView(props.view, false); invalidate()
}
function hit(event) {
  const rect = renderer.domElement.getBoundingClientRect()
  pointer.set((event.clientX - rect.left) / rect.width * 2 - 1, -(event.clientY - rect.top) / rect.height * 2 + 1)
  ray.setFromCamera(pointer, camera)
  const item = ray.intersectObjects(model.pickables, false)[0]
  const seat = item?.object.userData.seats?.[item.instanceId]
  tip.value = seat ? { seat, x: Math.max(8, Math.min(event.clientX - rect.left + 14, rect.width - 225)), y: Math.max(8, Math.min(event.clientY - rect.top + 14, rect.height - 100)) } : null
  renderer.domElement.style.cursor = seat ? 'pointer' : 'grab'
  return seat
}
function onDown(e) { pointerCount++; down = pointerCount === 1 ? { x: e.clientX, y: e.clientY, id: e.pointerId } : null }
function onMove(e) { if (!pointerCount && !props.blocked) hit(e); else tip.value = null }
function onUp(e) {
  pointerCount = Math.max(0, pointerCount - 1)
  if (down?.id === e.pointerId && !props.blocked && Math.hypot(e.clientX - down.x, e.clientY - down.y) < 6) {
    const seat = hit(e); if (seat) emit('select', seat)
  }
  down = null
}
function cancel() { down = null; pointerCount = 0; tip.value = null }
function contextLost(e) { e.preventDefault(); emit('error', '3D 显示已中断，请返回平面图后重试。') }
function manual() { tip.value = null; emit('manual-view') }
watch(() => props.selectedId, id => { model?.updateSelection(id); invalidate() })
watch(() => [props.room, props.seats, theme.isDark], rebuild)
watch(() => props.view, value => { if (value) setView(value) })
watch(() => props.blocked, value => { if (controls) controls.enabled = !value })
onMounted(() => {
  try {
    renderer = new THREE.WebGLRenderer({ antialias: true, powerPreference: 'high-performance' })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 1.75))
    renderer.outputColorSpace = THREE.SRGBColorSpace
    renderer.domElement.setAttribute('aria-label', '三维座位模型，拖动旋转，滚轮缩放；也可使用下方座位列表选择')
    host.value.prepend(renderer.domElement)
    scene = new THREE.Scene()
    scene.add(new THREE.HemisphereLight('#ffffff', '#738392', 2.5))
    const light = new THREE.DirectionalLight('#fff2df', 3); light.position.set(-8, 16, 8); scene.add(light)
    camera = new THREE.PerspectiveCamera(38, 1, .1, 500)
    controls = new OrbitControls(camera, renderer.domElement)
    controls.enableDamping = false; controls.enablePan = false
    controls.minDistance = 2; controls.maxPolarAngle = Math.PI / 2 - .08
    controls.addEventListener('change', invalidate); controls.addEventListener('start', manual)
    renderer.domElement.addEventListener('webglcontextlost', contextLost)
    resize(); rebuild()
    observer = new ResizeObserver(() => { resize(); if (props.view) setView(props.view, false) })
    observer.observe(host.value)
    emit('ready')
  } catch { emit('error', '当前设备无法启动 3D 场景，可返回平面图继续选座。') }
})
onBeforeUnmount(() => {
  disposed = true; cancelAnimationFrame(frame); observer?.disconnect(); cameraTween?.kill(); targetTween?.kill()
  controls?.dispose(); model?.dispose()
  renderer?.domElement.removeEventListener('webglcontextlost', contextLost)
  renderer?.dispose(); renderer?.forceContextLoss(); renderer?.domElement.remove()
})
</script>

<template>
  <div ref="host" class="scene-host" @pointerdown="onDown" @pointermove="onMove" @pointerup="onUp" @pointercancel="cancel" @pointerleave="cancel">
    <div v-if="tip" class="model-tip" :style="{ left: `${tip.x}px`, top: `${tip.y}px` }">
      <strong>{{ tip.seat.seat_no }} · {{ stateLabels[seatState(tip.seat, selectedId)] }}</strong>
      <span>{{ seatType(tip.seat) }} · {{ zoneName[tip.seat.zone] }}</span>
      <span v-if="tip.seat.near_window">靠窗座位</span>
    </div>
  </div>
</template>

<style scoped>
.scene-host { position: absolute; inset: 0; overflow: hidden; touch-action: none; }
.scene-host :deep(canvas) { display: block; width: 100%; height: 100%; }
.model-tip { position: absolute; z-index: 2; display: grid; gap: 5px; max-width: 220px; padding: 12px 15px; border: 1px solid var(--border); border-radius: 12px; background: var(--surface); color: var(--text-1); box-shadow: var(--shadow-3); pointer-events: none; font-size: 12px; }
.model-tip span { color: var(--text-3); }
</style>
