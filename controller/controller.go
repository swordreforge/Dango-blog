package controller

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"myblog-gogogo/db"
	"myblog-gogogo/service"
	"myblog-gogogo/service/settings"
)

var templates *template.Template

// getAppearanceSettings 获取外观设置，如果失败则返回默认设置
func getAppearanceSettings() *settings.AppearanceSettings {
	return service.GetAppearanceSettingsSafe()
}

func init() {
	// 初始化模板
	// 优先使用当前工作目录，而不是可执行文件所在目录
	// 这样可以支持 go run 和编译后的可执行文件
	templatePath := "template/*.html"
	
	// 尝试解析模板
	tmpl, err := template.ParseGlob(templatePath)
	if err != nil {
		// 如果当前工作目录找不到，尝试使用可执行文件所在目录
		execPath, execErr := os.Executable()
		if execErr == nil {
			execDir := filepath.Dir(execPath)
			templatePath = filepath.Join(execDir, "template/*.html")
			tmpl, err = template.ParseGlob(templatePath)
		}
	}
	
	if err != nil {
		panic(fmt.Sprintf("Failed to parse templates: %v", err))
	}
	
	templates = template.Must(tmpl, nil)
}

// renderTemplate 渲染HTML模板
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		RenderStatusPage(w, http.StatusInternalServerError)
	}
}

// IndexHandler 首页处理器
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	// 获取外观设置
	appearanceSettings, err := settings.GetAppearance()
	if err != nil {
		// 如果获取失败，使用默认值
		appearanceSettings = &settings.AppearanceSettings{
			BackgroundImage:      "/img/test.png",
			GlobalOpacity:        "0.15",
			BackgroundSize:       "cover",
			BackgroundPosition:   "center",
			BackgroundRepeat:     "no-repeat",
			BackgroundAttachment: "fixed",
			BlurAmount:           "20px",
			SaturateAmount:       "180%",
		}
	}

	// 获取模板设置
	templateSettings, err := settings.GetTemplate()
	if err != nil {
		// 如果获取失败，使用默认值
		templateSettings = &settings.TemplateSettings{
			Name:    "欢迎来到我的博客",
			Greting: "这是一个使用 Go 语言构建的个人博客系统，支持文章管理、数据分析等功能。",
			Year:    "2026",
			Foodes:  "我的博客",
		}
	}

	data := map[string]interface{}{
		"title":                    "我的博客",
		"name":                     templateSettings.Name,
		"greting":                  templateSettings.Greting,
		"year":                     templateSettings.Year,
		"foodes":                   templateSettings.Foodes,
		"Settings":                 appearanceSettings,
		"SwitchNotice":             templateSettings.SwitchNotice,		"SwitchNoticeText":         templateSettings.SwitchNoticeText,
		"ExternalLinkWarning":      templateSettings.ExternalLinkWarning,
		"ExternalLinkWhitelist":    templateSettings.ExternalLinkWhitelist,
		"ExternalLinkWarningText":  templateSettings.ExternalLinkWarningText,
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
	renderTemplate(w, "index.html", data)
}

// PassageHandler 文章页面处理器
func PassageHandler(w http.ResponseWriter, r *http.Request) {
	// 获取外观设置
	appearanceSettings := getAppearanceSettings()

	// 获取模板设置
	templateSettings, err := settings.GetTemplate()
	if err != nil {
		// 如果获取失败，使用默认值
		templateSettings = &settings.TemplateSettings{
			Name:               "欢迎来到我的博客",
			Greting:            "这是一个使用 Go 语言构建的个人博客系统",
			Year:               "2026",
			Foodes:             "我的博客",
			ArticleTitle:       true,
			ArticleTitlePrefix: "文章",
			SwitchNotice:       true,
			SwitchNoticeText:   "回来继续阅读",
		}
	}

	// 检查是否通过 id 参数访问
	idStr := r.URL.Query().Get("id")
	if idStr != "" {
		// 通过 ID 从数据库获取文章
		id, err := strconv.Atoi(idStr)
		if err != nil {
			RenderStatusPage(w, http.StatusBadRequest)
			return
		}

		repo := db.GetPassageRepository()
		passage, err := repo.GetByID(id)
		if err != nil {
			RenderStatusPage(w, http.StatusNotFound)
			return
		}

		// 解析 markdown 文件
		// 使用当前工作目录
		markdownPath := filepath.Join("markdown", passage.FilePath+".md")
		doc, err := service.ParseMarkdownFile(markdownPath)
		if err != nil {
			RenderStatusPage(w, http.StatusInternalServerError)
			return
		}

		// 从路径中提取日期
		pathParts := strings.Split(passage.FilePath, "/")
		var articleDate string
		if len(pathParts) >= 4 {
			year := pathParts[0]
			month := pathParts[1]
			day := pathParts[2]
			articleDate = fmt.Sprintf("%s-%s-%s", year, month, day)
		} else {
			articleDate = doc.CreatedAt.Format("2006-01-02")
		}

		// 渲染模板
		data := map[string]interface{}{
			"title":                    doc.Title,
			"Content":                  template.HTML(doc.Content),
			"Date":                     articleDate,
			"Path":                     passage.FilePath,
			"PassageID":                passage.ID,
			"ReadTime":                 service.CalculateReadTime(doc.Content),
			"Settings":                 appearanceSettings,
			"SwitchNotice":             templateSettings.SwitchNotice,
			"SwitchNoticeText":         templateSettings.SwitchNoticeText,
			"ExternalLinkWarning":      templateSettings.ExternalLinkWarning,
			"ExternalLinkWhitelist":    templateSettings.ExternalLinkWhitelist,
			"ExternalLinkWarningText":  templateSettings.ExternalLinkWarningText,
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
			"SponsorEnabled":           templateSettings.SponsorEnabled,
			"SponsorTitle":             templateSettings.SponsorTitle,
			"SponsorImage":             templateSettings.SponsorImage,
			"SponsorDescription":       templateSettings.SponsorDescription,
			"SponsorButtonText":        templateSettings.SponsorButtonText,
		}
		renderTemplate(w, "passage.html", data)
		return
	}

	// 检查是否是具体的文章路径
	if strings.HasPrefix(r.URL.Path, "/passage/") && r.URL.Path != "/passage/" {
		// 获取 markdown 文件路径
		markdownPath, err := service.GetMarkdownPath(r.URL.Path)
		if err != nil {
			RenderStatusPage(w, http.StatusNotFound)
			return
		}

		// 解析 markdown 文件
		doc, err := service.ParseMarkdownFile(markdownPath)
		if err != nil {
			RenderStatusPage(w, http.StatusInternalServerError)
			return
		}

		// 从 URL 路径中提取日期（用于匹配附件目录）
		// URL 路径格式: /passage/:year/:month/:day/:name
		pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		var articleDate string
		if len(pathParts) >= 4 && pathParts[0] == "passage" {
			year := pathParts[1]
			month := pathParts[2]
			day := pathParts[3]
			articleDate = fmt.Sprintf("%s-%s-%s", year, month, day)
		} else {
			// 如果无法从 URL 路径提取，使用文件修改时间作为 fallback
			articleDate = doc.CreatedAt.Format("2006-01-02")
		}

		// 根据标题从数据库获取文章ID
		repo := db.GetPassageRepository()
		passages, err := repo.GetAll(1000, 0)
		if err != nil {
			RenderStatusPage(w, http.StatusInternalServerError)
			return
		}

		var passageID int
		for _, p := range passages {
			if p.Title == doc.Title {
				passageID = p.ID
				break
			}
		}

		// 渲染模板（服务器端渲染）
				data := map[string]interface{}{
					"title":                    doc.Title,
					"Content":                  template.HTML(doc.Content),
					"Date":                     articleDate,
					"Path":                     strings.TrimPrefix(r.URL.Path, "/passage/"),
									"PassageID":                passageID,
									"ReadTime":                 service.CalculateReadTime(doc.Content),
									"Settings":                 appearanceSettings,
									"SwitchNotice":             templateSettings.SwitchNotice,
									"SwitchNoticeText":         templateSettings.SwitchNoticeText,
									"ExternalLinkWarning":      templateSettings.ExternalLinkWarning,
									"ExternalLinkWhitelist":    templateSettings.ExternalLinkWhitelist,
									"ExternalLinkWarningText":  templateSettings.ExternalLinkWarningText,
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
									"SponsorEnabled":           templateSettings.SponsorEnabled,
									"SponsorTitle":             templateSettings.SponsorTitle,
									"SponsorImage":             templateSettings.SponsorImage,
									"SponsorDescription":       templateSettings.SponsorDescription,
									"SponsorButtonText":        templateSettings.SponsorButtonText,
								}
								renderTemplate(w, "passage.html", data)
		return
	}

	// 默认显示文章列表页面
	data := map[string]interface{}{
		"title":                    "文章列表",
		"Settings":                 appearanceSettings,
		"SwitchNotice":             templateSettings.SwitchNotice,
		"SwitchNoticeText":         templateSettings.SwitchNoticeText,
		"ExternalLinkWarning":      templateSettings.ExternalLinkWarning,
		"ExternalLinkWhitelist":    templateSettings.ExternalLinkWhitelist,
		"ExternalLinkWarningText":  templateSettings.ExternalLinkWarningText,
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
	renderTemplate(w, "passage.html", data)
}

// CollectHandler 归档页面处理器
func CollectHandler(w http.ResponseWriter, r *http.Request) {
	// 获取外观设置
	appearanceSettings := getAppearanceSettings()

	// 获取模板设置
	templateSettings, err := settings.GetTemplate()
	if err != nil {
		// 如果获取失败，使用默认值
		templateSettings = &settings.TemplateSettings{
			Name:               "欢迎来到我的博客",
			Greting:            "这是一个使用 Go 语言构建的个人博客系统",
			Year:               "2026",
			Foodes:             "我的博客",
			ArticleTitle:       true,
			ArticleTitlePrefix: "文章",
			SwitchNotice:       true,
			SwitchNoticeText:   "回来继续阅读",
		}
	}

	data := map[string]interface{}{
		"title":                    "我的归档",
		"Settings":                 appearanceSettings,
		"SwitchNotice":             templateSettings.SwitchNotice,
		"SwitchNoticeText":         templateSettings.SwitchNoticeText,
		"ExternalLinkWarning":      templateSettings.ExternalLinkWarning,
		"ExternalLinkWhitelist":    templateSettings.ExternalLinkWhitelist,
		"ExternalLinkWarningText":  templateSettings.ExternalLinkWarningText,
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
	renderTemplate(w, "collect.html", data)
}

// AnalyzeHandler 分析页面处理器
func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	// 获取模板设置
	templateSettings, err := settings.GetTemplate()
	if err != nil {
		// 如果获取失败，使用默认值
		templateSettings = &settings.TemplateSettings{
			Name:               "欢迎来到我的博客",
			Greting:            "这是一个使用 Go 语言构建的个人博客系统",
			Year:               "2026",
			Foodes:             "我的博客",
			ArticleTitle:       true,
			ArticleTitlePrefix: "文章",
			SwitchNotice:       true,
			SwitchNoticeText:   "回来继续阅读",
		}
	}

	data := map[string]interface{}{
		"title":                    "数据分析",
		"SwitchNotice":             templateSettings.SwitchNotice,
		"SwitchNoticeText":         templateSettings.SwitchNoticeText,
		"ExternalLinkWarning":      templateSettings.ExternalLinkWarning,
		"ExternalLinkWhitelist":    templateSettings.ExternalLinkWhitelist,
		"ExternalLinkWarningText":  templateSettings.ExternalLinkWarningText,
	}
	renderTemplate(w, "analyze.html", data)
}

// AboutHandler 关于页面处理器
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	// 获取外观设置
	appearanceSettings := getAppearanceSettings()

	// 获取模板设置
	templateSettings, err := settings.GetTemplate()
	if err != nil {
		// 如果获取失败，使用默认值
		templateSettings = &settings.TemplateSettings{
			Name:               "欢迎来到我的博客",
			Greting:            "这是一个使用 Go 语言构建的个人博客系统",
			Year:               "2026",
			Foodes:             "我的博客",
			ArticleTitle:       true,
			ArticleTitlePrefix: "文章",
			SwitchNotice:       true,
			SwitchNoticeText:   "回来继续阅读",
		}
	}

	data := map[string]interface{}{
		"title":                    "关于我",
		"Settings":                 appearanceSettings,
		"SwitchNotice":             templateSettings.SwitchNotice,
		"SwitchNoticeText":         templateSettings.SwitchNoticeText,
		"ExternalLinkWarning":      templateSettings.ExternalLinkWarning,
		"ExternalLinkWhitelist":    templateSettings.ExternalLinkWhitelist,
		"ExternalLinkWarningText":  templateSettings.ExternalLinkWarningText,
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
	renderTemplate(w, "about.html", data)
}

// RenderStatusPage 渲染状态码页面
func RenderStatusPage(w http.ResponseWriter, statusCode int) {
	var templateFile string
	var title string
	var codeStr string

	switch statusCode {
	case 302:
		templateFile = "302.html"
		title = "重定向"
		codeStr = "302"
	case 401:
		templateFile = "401.html"
		title = "未授权"
		codeStr = "401"
	case 404:
		templateFile = "404.html"
		title = "页面未找到"
		codeStr = "404"
	case 405:
		templateFile = "405.html"
		title = "方法不允许"
		codeStr = "405"
	case 409:
		templateFile = "409.html"
		title = "冲突"
		codeStr = "409"
	case 500:
		templateFile = "500.html"
		title = "服务器错误"
		codeStr = "500"
	case 999:
		templateFile = "999.html"
		title = "未知错误"
		codeStr = "999"
	case 429:
		templateFile = "429.html"
		title = "请求过多"
		codeStr = "429"
	default:
		templateFile = "404.html"
		title = "页面未找到"
		codeStr = "404"
	}

	// 使用当前工作目录
	tmplPath := filepath.Join("template", "status", templateFile)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "模板加载失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	data := map[string]interface{}{
		"title": title,
		"Code":  codeStr,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// StatusHandler 状态页面处理器
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	statusCodeStr := r.URL.Path[len("/status/"):]
	
	statusCode := 404
	switch statusCodeStr {
	case "302":
		statusCode = 302
	case "404":
		statusCode = 404
	case "405":
		statusCode = 405
	case "409":
		statusCode = 409
	case "500":
		statusCode = 500
	case "999":
		statusCode = 999
	case "429":
		statusCode = 429
	}
	
	RenderStatusPage(w, statusCode)
}

// sendErrorResponse 发送错误响应
func sendErrorResponse(w http.ResponseWriter, statusCode int, message string, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"success": false,
		"message": message,
	}
	
	if code != "" {
		response["code"] = code
	}
	
	json.NewEncoder(w).Encode(response)
}
