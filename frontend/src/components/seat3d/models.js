import * as THREE from 'three'
import { getWindowSides, seatState, stateColors } from './seatPresentation'

// Model parts are batched by material. Every instance retains its seat identity.
export function buildRoom(room, seats, dark) {
  const root = new THREE.Group()
  const geometry = new THREE.BoxGeometry(1, 1, 1)
  const batches = new Map()
  const materials = new Set()
  const pickables = []
  const markers = []
  const walls = []
  const width = room.seat_cols * 1.8 + 1.2
  const depth = room.seat_rows * 1.9 + 1.2
  const colors = { wood: '#d8b58e', leg: '#727d85', chair: '#7b9d95', case: '#394651', screen: '#8bcdd5', key: '#becbd0', socket: '#eee9dc', hole: '#555e66' }
  function part(material, x, y, z, sx, sy, sz, seat) {
    if (!batches.has(material)) batches.set(material, [])
    batches.get(material).push({ x, y, z, sx, sy, sz, seat })
  }
  function ordinaryDesk(x, z, seat) {
    part('wood', x, .82, z, 1.28, .12, .78, seat)
    for (const dx of [-.52, .52]) for (const dz of [-.27, .27]) part('leg', x + dx, .4, z + dz, .065, .8, .065, seat)
    part('chair', x, .46, z + .69, .54, .09, .5, seat)
    part('chair', x, .78, z + .91, .54, .55, .065, seat)
    for (const dx of [-.21, .21]) for (const dz of [.5, .88]) part('leg', x + dx, .23, z + dz, .045, .46, .045, seat)
  }
  function socket(x, z, seat) {
    part('socket', x + .44, .91, z - .22, .26, .065, .13, seat)
    for (const dx of [.37, .42, .48, .53]) part('hole', x + dx, .946, z - .22, .016, .006, .045, seat)
  }
  function powerDesk(x, z, seat) { ordinaryDesk(x, z, seat); socket(x, z, seat) }
  function computerDesk(x, z, seat) {
    ordinaryDesk(x, z, seat)
    part('case', x, 1.23, z - .19, .65, .42, .06, seat)
    part('screen', x, 1.24, z - .154, .58, .34, .013, seat)
    part('case', x, .97, z - .19, .05, .2, .05, seat)
    part('case', x, .89, z - .19, .28, .025, .19, seat)
    part('key', x, .903, z + .15, .46, .035, .17, seat)
    part('key', x + .39, .908, z + .15, .085, .045, .12, seat)
    part('case', x + .43, .27, z, .22, .48, .4, seat)
    if (seat.has_power) socket(x, z, seat)
  }
  for (const seat of seats) {
    const x = (seat.col_no - (room.seat_cols + 1) / 2) * 1.8
    const z = (seat.row_no - (room.seat_rows + 1) / 2) * 1.9
    if (seat.zone === 'computer') computerDesk(x, z, seat)
    else if (seat.has_power) powerDesk(x, z, seat)
    else ordinaryDesk(x, z, seat)
    markers.push({ x, z, seat })
  }
  const dummy = new THREE.Object3D()
  for (const [key, parts] of batches) {
    const material = new THREE.MeshStandardMaterial({ color: colors[key], roughness: .72 })
    materials.add(material)
    const mesh = new THREE.InstancedMesh(geometry, material, parts.length)
    mesh.userData.seats = parts.map(p => p.seat)
    parts.forEach((p, i) => {
      dummy.position.set(p.x, p.y, p.z); dummy.scale.set(p.sx, p.sy, p.sz); dummy.updateMatrix()
      mesh.setMatrixAt(i, dummy.matrix)
    })
    root.add(mesh); pickables.push(mesh)
  }
  const markerMaterial = new THREE.MeshStandardMaterial({ roughness: .85 })
  materials.add(markerMaterial)
  const statusMesh = new THREE.InstancedMesh(geometry, markerMaterial, markers.length)
  statusMesh.userData.seats = seats
  markers.forEach((p, i) => {
    dummy.position.set(p.x, .025, p.z + .22); dummy.scale.set(1.5, .04, 1.64); dummy.updateMatrix()
    statusMesh.setMatrixAt(i, dummy.matrix)
  })
  root.add(statusMesh); pickables.push(statusMesh)
  function box(parent, color, x, y, z, sx, sy, sz, glass = false) {
    const material = new THREE.MeshStandardMaterial({ color, roughness: glass ? .15 : .8, metalness: glass ? .12 : 0, transparent: glass, opacity: glass ? .32 : 1, depthWrite: !glass })
    materials.add(material)
    const mesh = new THREE.Mesh(geometry, material)
    mesh.position.set(x, y, z); mesh.scale.set(sx, sy, sz); parent.add(mesh)
  }
  box(root, dark ? '#293542' : '#e9e5dc', 0, -.1, 0, width, .2, depth)
  const windowSides = getWindowSides(room, seats)
  function windowModel(parent, length, wallColor, outward) {
    // A full-height facade with actual openings, masonry piers and hinged sashes.
    const height = 2.9, sill = .98, lintel = 2.48
    const count = Math.max(1, Math.floor(length / 1.85))
    const bay = length / count, pier = .18
    const frame = dark ? '#9aadb7' : '#d5e0e2'
    box(parent, wallColor, 0, sill / 2, 0, length, sill, .22)
    box(parent, wallColor, 0, (height + lintel) / 2, 0, length, height - lintel, .22)
    box(parent, dark ? '#75838a' : '#e0ded5', 0, sill, 0, length + .04, .09, .38)
    box(parent, dark ? '#2c3943' : '#d6d2c8', 0, .07, -outward * .125, length, .14, .04)
    for (let i = 0; i <= count; i++) {
      box(parent, wallColor, -length / 2 + bay * i, (sill + lintel) / 2, 0, pier, lintel - sill, .22)
    }
    for (let i = 0; i < count; i++) {
      const center = -length / 2 + bay * (i + .5)
      const openingWidth = bay - pier, openingHeight = lintel - sill
      const middle = (sill + lintel) / 2
      for (const y of [sill + .04, lintel - .04]) box(parent, frame, center, y, 0, openingWidth, .065, .12)
      for (const x of [center - openingWidth / 2 + .03, center, center + openingWidth / 2 - .03]) box(parent, frame, x, middle, 0, .055, openingHeight, .12)
      // Two separate leaves per opening. Alternate open windows create a clear silhouette.
      for (let leaf = 0; leaf < 2; leaf++) {
        const sash = new THREE.Group()
        const leafWidth = openingWidth / 2 - .055
        const leafHeight = openingHeight - .13
        sash.position.set(center - openingWidth / 2 + .04 + leaf * openingWidth / 2, middle, 0)
        const opened = leaf === 1 && i % 3 === 0
        if (opened) sash.rotation.y = -outward * Math.PI / 5
        for (const x of [.025, leafWidth - .025]) box(sash, frame, x, 0, 0, .05, leafHeight, .065)
        for (const y of [-leafHeight / 2 + .025, leafHeight / 2 - .025]) box(sash, frame, leafWidth / 2, y, 0, leafWidth, .05, .065)
        box(sash, '#a3d6e6', leafWidth / 2, 0, 0, leafWidth - .1, leafHeight - .1, .018, true)
        box(sash, '#647985', leafWidth - .085, -.08, -outward * .055, .035, .16, .035)
        for (const y of [-leafHeight * .32, leafHeight * .32]) box(sash, '#788992', 0, y, 0, .065, .1, .09)
        parent.add(sash)
      }
    }
  }
  for (const side of ['top', 'bottom', 'left', 'right']) {
    const wall = new THREE.Group()
    const horizontal = ['top', 'bottom'].includes(side)
    const length = horizontal ? width : depth
    if (horizontal) wall.position.z = (side === 'top' ? -1 : 1) * depth / 2
    else { wall.position.x = (side === 'left' ? -1 : 1) * width / 2; wall.rotation.y = Math.PI / 2 }
    const wallColor = dark ? '#405160' : '#f5f1e9'
    if (windowSides[side]) {
      windowModel(wall, length, wallColor, ['top', 'left'].includes(side) ? -1 : 1)
    } else if (side === 'bottom') {
      for (const sign of [-1, 1]) box(wall, wallColor, sign * (length / 4 + .3), .27, 0, length / 2 - .6, .54, .12)
      box(wall, '#bba78e', 0, .012, 0, 1.15, .025, .22)
    } else box(wall, wallColor, 0, .27, 0, length, .54, .12)
    const facadeMaterials = []
    wall.traverse(o => {
      if (o.isMesh) facadeMaterials.push({ material: o.material, opacity: o.material.opacity, depthWrite: o.material.depthWrite })
    })
    root.add(wall); walls.push({ side, mesh: wall, window: windowSides[side], facadeMaterials, faded: false })
  }
  function updateSelection(selectedId) {
    markers.forEach((p, i) => statusMesh.setColorAt(i, new THREE.Color(stateColors[seatState(p.seat, selectedId)])))
    if (statusMesh.instanceColor) statusMesh.instanceColor.needsUpdate = true
  }
  updateSelection(null)
  return { root, pickables, width, depth, updateSelection,
    updateWalls(camera) {
      for (const wall of walls) {
        const { side, mesh } = wall
        const facing = side === 'top' ? camera.position.z < -depth / 2 : side === 'bottom' ? camera.position.z > depth / 2 : side === 'left' ? camera.position.x < -width / 2 : camera.position.x > width / 2
        mesh.visible = wall.window || !facing
        // Keep the window silhouette when it is in front, without hiding nearby seats.
        if (wall.window && wall.faded !== facing) {
          wall.faded = facing
          for (const { material, opacity, depthWrite } of wall.facadeMaterials) {
            material.opacity = facing ? opacity * .22 : opacity
            material.transparent = facing || opacity < 1
            material.depthWrite = facing ? false : depthWrite
            material.needsUpdate = true
          }
        }
      }
    },
    dispose() { root.traverse(o => { if (o.isInstancedMesh) o.dispose() }); geometry.dispose(); materials.forEach(m => m.dispose()) }
  }
}
