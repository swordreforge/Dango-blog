package admin

import (
	"html/template"
	"net/http"

	"myblog-gogogo/service"
	"myblog-gogogo/service/settings"
)

// AdminHandler 管理后台页面处理器
func AdminHandler(w http.ResponseWriter, r *http.Request) {
	// 使用当前工作目录（支持 go run 和编译后的可执行文件）
	tmplPath := "template/admin/admin.html"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "模板加载失败", http.StatusInternalServerError)
		return
	}

	// 获取外观设置
	appearanceSettings := service.GetAppearanceSettingsMap()

	// 获取模板设置
	templateSettings, err := settings.GetTemplate()
	if err != nil {
		templateSettings = &settings.TemplateSettings{
			Name:                    "欢迎来到我的博客",
			Greting:                 "这是一个使用 Go 语言构建的个人博客系统",
			Year:                    "2026",
			Foodes:                  "我的博客",
			ArticleTitle:            true,
			ArticleTitlePrefix:      "文章",
			SwitchNotice:            true,
			SwitchNoticeText:        "回来继续阅读",
			ExternalLinkWarning:     true,
			ExternalLinkWhitelist:   "github.com,gitee.com,stackoverflow.com",
			ExternalLinkWarningText: "您即将离开本站，前往外部链接",
			Live2DEnabled:           false,
			Live2DShowOnIndex:       true,
			Live2DShowOnPassage:     true,
			Live2DShowOnCollect:     true,
			Live2DShowOnAbout:       true,
			Live2DShowOnAdmin:       false,
			Live2DModelId:           "1",
			Live2DModelPath:         "",
			Live2DCDNPath:           "https://unpkg.com/live2d-widget-model@1.0.5/",
			Live2DPosition:          "right",
			Live2DWidth:             "280px",
			Live2DHeight:            "250px",
		}
	}

	data := map[string]interface{}{
		"title":    "管理后台",
		"year":     "2026",
		"foodes":   "我的博客",
		"Settings": appearanceSettings,
		"Live2DEnabled":            templateSettings.Live2DEnabled,
		"Live2DShowOnIndex":        templateSettings.Live2DShowOnIndex,
		"Live2DShowOnPassage":      templateSettings.Live2DShowOnPassage,
		"Live2DShowOnCollect":      templateSettings.Live2DShowOnCollect,
		"Live2DShowOnAbout":        templateSettings.Live2DShowOnAbout,
		"Live2DShowOnAdmin":        templateSettings.Live2DShowOnAdmin,
		"Live2DModelId":            templateSettings.Live2DModelId,
		"Live2DModelPath":          templateSettings.Live2DModelPath,
		"Live2DCDNPath":            templateSettings.Live2DCDNPath,
		"Live2DPosition":           templateSettings.Live2DPosition,
		"Live2DWidth":              templateSettings.Live2DWidth,
		"Live2DHeight":             templateSettings.Live2DHeight,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}