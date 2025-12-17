// 本地存储工具类

const TOKEN_KEY = 'ltts_token'
const USER_INFO_KEY = 'ltts_user_info'

/**
 * 获取 token
 */
export const getToken = (): string | null => {
  return localStorage.getItem(TOKEN_KEY)
}

/**
 * 设置 token
 */
export const setToken = (token: string): void => {
  localStorage.setItem(TOKEN_KEY, token)
}

/**
 * 移除 token
 */
export const removeToken = (): void => {
  localStorage.removeItem(TOKEN_KEY)
}

/**
 * 获取用户信息
 */
export const getUserInfo = (): any => {
  const userInfo = localStorage.getItem(USER_INFO_KEY)
  return userInfo ? JSON.parse(userInfo) : null
}

/**
 * 设置用户信息
 */
export const setUserInfo = (userInfo: any): void => {
  localStorage.setItem(USER_INFO_KEY, JSON.stringify(userInfo))
}

/**
 * 移除用户信息
 */
export const removeUserInfo = (): void => {
  localStorage.removeItem(USER_INFO_KEY)
}

/**
 * 清除所有存储
 */
export const clearStorage = (): void => {
  removeToken()
  removeUserInfo()
}

