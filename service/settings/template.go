package settings

import (
	"fmt"
	"log"
	"strconv"

	"myblog-gogogo/db"
)

// TemplateSettings 模板设置结构
type TemplateSettings struct {
	Name                    string `json:"name"`
	Greting                 string `json:"greting"`
	Year                    string `json:"year"`
	Foodes                  string `json:"foodes"`
	ArticleTitle            bool   `json:"article_title"`
	ArticleTitlePrefix      string `json:"article_title_prefix"`
	SwitchNotice            bool   `json:"switch_notice"`
	SwitchNoticeText        string `json:"switch_notice_text"`
	ExternalLinkWarning     bool   `json:"external_link_warning"`
	ExternalLinkWhitelist   string `json:"external_link_whitelist"`
	ExternalLinkWarningText string `json:"external_link_warning_text"`
	Live2DEnabled           bool   `json:"live2d_enabled"`
	Live2DShowOnIndex       bool   `json:"live2d_show_on_index"`
	Live2DShowOnPassage     bool   `json:"live2d_show_on_passage"`
	Live2DShowOnCollect     bool   `json:"live2d_show_on_collect"`
	Live2DShowOnAbout       bool   `json:"live2d_show_on_about"`
	Live2DShowOnAdmin       bool   `json:"live2d_show_on_admin"`
	Live2DModelId           string `json:"live2d_model_id"`
	Live2DModelPath         string `json:"live2d_model_path"`
	Live2DCDNPath           string `json:"live2d_cdn_path"`
	Live2DPosition          string `json:"live2d_position"`
	Live2DWidth             string `json:"live2d_width"`
	Live2DHeight            string `json:"live2d_height"`
	SponsorEnabled          bool   `json:"sponsor_enabled"`
	SponsorTitle            string `json:"sponsor_title"`
	SponsorImage            string `json:"sponsor_image"`
	SponsorDescription      string `json:"sponsor_description"`
	SponsorButtonText       string `json:"sponsor_button_text"`
	GlobalAvatar            string `json:"global_avatar"`
	AttachmentDefaultVisibility string `json:"attachment_default_visibility"`
	AttachmentMaxSize          int64  `json:"attachment_max_size"`
	AttachmentAllowedTypes     string `json:"attachment_allowed_types"`
}

// GetTemplate 获取模板设置
func GetTemplate() (*TemplateSettings, error) {
	repo := db.GetSettingRepository()

	settings := TemplateSettings{
		Name:                    "欢迎来到我的博客",
		Greting:                 "这是一个使用 Go 语言构建的个人博客系统，支持文章管理、数据分析等功能。",
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
		SponsorEnabled:          false,
		SponsorTitle:            "感谢您的支持",
		SponsorImage:            "/img/avatar.png",
		SponsorDescription:      "如果您觉得这个博客对您有帮助，欢迎赞助支持！",
		SponsorButtonText:       "❤️ 赞助支持",
		GlobalAvatar:            "/img/avatar.png",
		AttachmentDefaultVisibility: "public",
		AttachmentMaxSize:          500 * 1024 * 1024,
		AttachmentAllowedTypes:     "jpg,jpeg,png,gif,mp4,mp3,pdf,doc,docx,xls,xlsx,ppt,pptx,zip,rar,7z,tar,gz",
	}

	keys := []string{
		"template_name", "template_greting", "template_year", "template_foods",
		"template_article_title", "template_article_title_prefix",
		"template_switch_notice", "template_switch_notice_text",
		"external_link_warning", "external_link_whitelist", "external_link_warning_text",
		"live2d_enabled",
		"live2d_show_on_index", "live2d_show_on_passage", "live2d_show_on_collect",
		"live2d_show_on_about", "live2d_show_on_admin",
		"live2d_model_id", "live2d_model_path", "live2d_cdn_path",
		"live2d_position", "live2d_width", "live2d_height",
		"sponsor_enabled", "sponsor_title", "sponsor_image",
		"sponsor_description", "sponsor_button_text",
		"global_avatar",
		"attachment_default_visibility", "attachment_max_size", "attachment_allowed_types",
	}

	for _, key := range keys {
		setting, err := repo.GetByKey(key)
		if err != nil {
			log.Printf("Failed to get setting %s: %v", key, err)
			continue
		}
		if setting != nil {
			switch key {
			case "template_name":
				settings.Name = setting.Value
			case "template_greting":
				settings.Greting = setting.Value
			case "template_year":
				settings.Year = setting.Value
			case "template_foods":
				settings.Foodes = setting.Value
			case "template_article_title":
				settings.ArticleTitle = stringToBool(setting.Value)
			case "template_article_title_prefix":
				settings.ArticleTitlePrefix = setting.Value
			case "template_switch_notice":
				settings.SwitchNotice = stringToBool(setting.Value)
			case "template_switch_notice_text":
				settings.SwitchNoticeText = setting.Value
			case "external_link_warning":
				settings.ExternalLinkWarning = stringToBool(setting.Value)
			case "external_link_whitelist":
				settings.ExternalLinkWhitelist = setting.Value
			case "external_link_warning_text":
				settings.ExternalLinkWarningText = setting.Value
			case "live2d_enabled":
				settings.Live2DEnabled = stringToBool(setting.Value)
			case "live2d_show_on_index":
				settings.Live2DShowOnIndex = stringToBool(setting.Value)
			case "live2d_show_on_passage":
				settings.Live2DShowOnPassage = stringToBool(setting.Value)
			case "live2d_show_on_collect":
				settings.Live2DShowOnCollect = stringToBool(setting.Value)
			case "live2d_show_on_about":
				settings.Live2DShowOnAbout = stringToBool(setting.Value)
			case "live2d_show_on_admin":
				settings.Live2DShowOnAdmin = stringToBool(setting.Value)
			case "live2d_model_id":
				settings.Live2DModelId = setting.Value
			case "live2d_model_path":
				settings.Live2DModelPath = setting.Value
			case "live2d_cdn_path":
				settings.Live2DCDNPath = setting.Value
			case "live2d_position":
				settings.Live2DPosition = setting.Value
			case "live2d_width":
				settings.Live2DWidth = setting.Value
			case "live2d_height":
				settings.Live2DHeight = setting.Value
			case "sponsor_enabled":
				settings.SponsorEnabled = stringToBool(setting.Value)
			case "sponsor_title":
				settings.SponsorTitle = setting.Value
			case "sponsor_image":
				settings.SponsorImage = setting.Value
			case "sponsor_description":
				settings.SponsorDescription = setting.Value
			case "sponsor_button_text":
				settings.SponsorButtonText = setting.Value
			case "global_avatar":
				settings.GlobalAvatar = setting.Value
			case "attachment_default_visibility":
				settings.AttachmentDefaultVisibility = setting.Value
			case "attachment_max_size":
				settings.AttachmentMaxSize = stringToInt(setting.Value)
			case "attachment_allowed_types":
				settings.AttachmentAllowedTypes = setting.Value
			}
		}
	}

	return &settings, nil
}

// UpdateTemplate 更新模板设置
func UpdateTemplate(settings *TemplateSettings) error {
	repo := db.GetSettingRepository()

	updates := map[string]string{
		"template_name":               settings.Name,
		"template_greting":            settings.Greting,
		"template_year":                settings.Year,
		"template_foods":              settings.Foodes,
		"template_article_title":       boolToString(settings.ArticleTitle),
		"template_article_title_prefix": settings.ArticleTitlePrefix,
		"template_switch_notice":       boolToString(settings.SwitchNotice),
		"template_switch_notice_text":   settings.SwitchNoticeText,
		"external_link_warning":        boolToString(settings.ExternalLinkWarning),
		"external_link_whitelist":      settings.ExternalLinkWhitelist,
		"external_link_warning_text":    settings.ExternalLinkWarningText,
		"live2d_enabled":                boolToString(settings.Live2DEnabled),
		"live2d_show_on_index":          boolToString(settings.Live2DShowOnIndex),
		"live2d_show_on_passage":        boolToString(settings.Live2DShowOnPassage),
		"live2d_show_on_collect":        boolToString(settings.Live2DShowOnCollect),
		"live2d_show_on_about":          boolToString(settings.Live2DShowOnAbout),
		"live2d_show_on_admin":          boolToString(settings.Live2DShowOnAdmin),
		"live2d_model_id":               settings.Live2DModelId,
		"live2d_model_path":             settings.Live2DModelPath,
		"live2d_cdn_path":               settings.Live2DCDNPath,
		"live2d_position":               settings.Live2DPosition,
		"live2d_width":                  settings.Live2DWidth,
		"live2d_height":                 settings.Live2DHeight,
		"sponsor_enabled":               boolToString(settings.SponsorEnabled),
		"sponsor_title":                 settings.SponsorTitle,
		"sponsor_image":                 settings.SponsorImage,
		"sponsor_description":           settings.SponsorDescription,
		"sponsor_button_text":           settings.SponsorButtonText,
		"global_avatar":                 settings.GlobalAvatar,
		"attachment_default_visibility": settings.AttachmentDefaultVisibility,
		"attachment_max_size":           strconv.FormatInt(settings.AttachmentMaxSize, 10),
		"attachment_allowed_types":      settings.AttachmentAllowedTypes,
	}

	for key, value := range updates {
		if err := repo.UpdateByKey(key, value); err != nil {
			return fmt.Errorf("failed to update setting %s: %w", key, err)
		}
	}

	return nil
}

// UpdateTemplatePartial 增量更新模板设置
func UpdateTemplatePartial(updates map[string]interface{}) error {
	repo := db.GetSettingRepository()

	fieldMapping := map[string]string{
		"name":                      "template_name",
		"greting":                   "template_greting",
		"year":                      "template_year",
		"foodes":                    "template_foods",
		"article_title":             "template_article_title",
		"article_title_prefix":      "template_article_title_prefix",
		"switch_notice":             "template_switch_notice",
		"switch_notice_text":        "template_switch_notice_text",
		"external_link_warning":     "external_link_warning",
		"external_link_whitelist":   "external_link_whitelist",
		"external_link_warning_text": "external_link_warning_text",
		"live2d_enabled":            "live2d_enabled",
		"live2d_show_on_index":      "live2d_show_on_index",
		"live2d_show_on_passage":    "live2d_show_on_passage",
		"live2d_show_on_collect":    "live2d_show_on_collect",
		"live2d_show_on_about":      "live2d_show_on_about",
		"live2d_show_on_admin":      "live2d_show_on_admin",
		"live2d_model_id":           "live2d_model_id",
		"live2d_model_path":         "live2d_model_path",
		"live2d_cdn_path":           "live2d_cdn_path",
		"live2d_position":           "live2d_position",
		"live2d_width":              "live2d_width",
		"live2d_height":             "live2d_height",
		"sponsor_enabled":           "sponsor_enabled",
		"sponsor_title":             "sponsor_title",
		"sponsor_image":             "sponsor_image",
		"sponsor_description":       "sponsor_description",
		"sponsor_button_text":       "sponsor_button_text",
		"global_avatar":             "global_avatar",
	}

	for jsonField, value := range updates {
		dbKey, exists := fieldMapping[jsonField]
		if !exists {
			log.Printf("Unknown field: %s", jsonField)
			continue
		}

		var stringValue string
		switch v := value.(type) {
		case string:
			stringValue = v
		case bool:
			stringValue = boolToString(v)
		case float64:
			stringValue = fmt.Sprintf("%.0f", v)
		default:
			stringValue = fmt.Sprintf("%v", v)
		}

		if err := repo.UpdateByKey(dbKey, stringValue); err != nil {
			return fmt.Errorf("failed to update setting %s: %w", dbKey, err)
		}
	}

	return nil
}