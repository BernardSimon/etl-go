// src/i18n.ts
import { createI18n } from 'vue-i18n'

// 1. 定义支持的语言列表
const supportedLocales = ['en', 'zh']

// 2. 创建 i18n 实例 (messages 初始为空)
const i18n = createI18n({
    locale: 'en', // 默认语言
    fallbackLocale: 'en',
    messages: {}
})

export default i18n

/**
 * 按需加载语言包
 * @param lang 目标语言代码
 */
// 预先导入所有支持的语言包
const languageModules = {
    en: () => import('./locales/en.json'),
    zh: () => import('./locales/zh.json')
}

export async function loadLanguageAsync(lang: string): Promise<void> {
    if (!supportedLocales.includes(lang)) {
        lang = 'en'
    }

    try {
        const loader = languageModules[lang as keyof typeof languageModules]
        if (loader) {
            const messages = await loader()
            i18n.global.setLocaleMessage(lang, messages.default)
            i18n.global.locale = lang
        }
    } catch (error) {
        console.error(`Failed to load language ${lang}:`, error)
    }
}


export function getCurrentLocale(): string {
    return i18n.global.locale
}
