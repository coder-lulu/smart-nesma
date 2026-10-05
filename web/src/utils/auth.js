import { useUserStore } from '@/pinia/modules/user'

/**
 * 获取当前用户的token
 * @returns {string} 用户token
 */
export function getToken() {
  const userStore = useUserStore()
  return userStore.token
}

/**
 * 设置用户token
 * @param {string} token - 用户token
 */
export function setToken(token) {
  const userStore = useUserStore()
  userStore.setToken(token)
}

/**
 * 清除用户token
 */
export function removeToken() {
  const userStore = useUserStore()
  userStore.setToken('')
}

/**
 * 检查用户是否已登录
 * @returns {boolean} 是否已登录
 */
export function isLoggedIn() {
  const token = getToken()
  return !!(token && token.trim())
}

export default {
  getToken,
  setToken,
  removeToken,
  isLoggedIn
} 