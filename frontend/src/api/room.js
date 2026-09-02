import request from './request'

// 房间与座位
export function getRooms() {
  return request.get('/rooms')
}

export function getSeatMap(roomId, date, start, end) {
  return request.get(`/rooms/${roomId}/seats`, { params: { date, start, end } })
}
