package i18n

import (
	"embed"
	"encoding/json"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locale/*
var localeFs embed.FS

var AcceptLanguages = make([]string, 0)

// 本地化消息缓存，提高性能
var (
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	once      sync.Once
)

// 初始化i18n bundle
func init() {
	once.Do(func() {
		bundle = i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

		// 加载所有语言文件
		files, _ := localeFs.ReadDir("locale")
		for _, file := range files {
			if !file.IsDir() {
				if !strings.HasSuffix(file.Name(), ".json") {
					continue
				}
				data, _ := localeFs.ReadFile("locale/" + file.Name())
				AcceptLanguages = append(AcceptLanguages, strings.TrimSuffix(file.Name(), ".json"))
				bundle.ParseMessageFileBytes(data, file.Name())
			}
		}
	})
}

// Translate 根据语言和键获取翻译值
// 优先级：指定语言 -> 英语 -> 原始键
func Translate(lang, key string) string {
	localizer := i18n.NewLocalizer(bundle, lang, "en")

	result, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})

	// 如果找不到翻译，则返回原始键
	if err != nil {
		return key
	}

	return result
}

func TranslateWithContext(c *gin.Context, key string) string {
	lang := c.GetString("language")
	return Translate(lang, key)
}
