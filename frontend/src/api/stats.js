import request from './request'

// 统计分析
export function getHeatmap(roomId, date) {
  return request.get('/stats/heatmap', { params: { room_id: roomId, date } })
}

export function getOverview() {
  return request.get('/stats/overview')
}
