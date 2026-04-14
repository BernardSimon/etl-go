// 本地存储工具类

const TOKEN_KEY = 'ltts_token'
const REFRESH_TOKEN_KEY = 'ltts_refresh_token'
const USER_INFO_KEY = 'ltts_user_info'
const LANGUAGE_KEY = 'ltts_language'
const THEME_KEY = 'ltts_theme'

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
 * 获取 refresh token
 */
export const getRefreshToken = (): string | null => {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

/**
 * 设置 refresh token
 */
export const setRefreshToken = (token: string): void => {
  localStorage.setItem(REFRESH_TOKEN_KEY, token)
}

/**
 * 移除 refresh token
 */
export const removeRefreshToken = (): void => {
  localStorage.removeItem(REFRESH_TOKEN_KEY)
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
 * 获取语言
 */
export const getLanguage = (): string | null => {
  return localStorage.getItem(LANGUAGE_KEY)
}

/**
 * 设置语言
 */
export const setLanguage = (language: string): void => {
  localStorage.setItem(LANGUAGE_KEY, language)
}

/**
 * 移除语言
 */
export const removeLanguage = (): void => {
  localStorage.removeItem(LANGUAGE_KEY)
}

/**
 * 获取主题
 */
export const getTheme = (): string | null => {
  return localStorage.getItem(THEME_KEY)
}

/**
 * 设置主题
 */
export const setTheme = (theme: string): void => {
  localStorage.setItem(THEME_KEY, theme)
}

/**
 * 移除主题
 */
export const removeTheme = (): void => {
  localStorage.removeItem(THEME_KEY)
}

/**
 * 清除所有存储
 */
export const clearStorage = (): void => {
  removeToken()
  removeRefreshToken()
  removeUserInfo()
  removeLanguage()
  removeTheme()
}
