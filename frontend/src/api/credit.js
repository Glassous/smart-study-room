import request from './request'

// 信用分
export function getCreditOverview() {
  return request.get('/credit')
}
