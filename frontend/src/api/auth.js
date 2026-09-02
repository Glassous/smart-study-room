import request from './request'

// 认证相关 API
export function login(data) {
  return request.post('/auth/login', data)
}

export function register(data) {
  return request.post('/auth/register', data)
}

export function fetchProfile() {
  return request.get('/auth/profile')
}
