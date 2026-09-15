import request from './request'

// 消息通知
export function listNotifications() {
  return request.get('/notifications')
}

export function unreadCount() {
  return request.get('/notifications/unread_count')
}

export function markRead(id) {
  return request.post(`/notifications/${id}/read`)
}

export function markAllRead() {
  return request.post('/notifications/read_all')
}
