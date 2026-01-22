package validation

import (
	"errors"
)

// ValidateImage 通用图片验证（用于JPG、PNG、GIF、WEBP等）
func ValidateImage(content []byte) error {
	// 检查文件大小限制（50MB）
	if err := CheckFileSize(content, 50*1024*1024); err != nil {
		return err
	}

	// 检查是否为有效的图片文件
	if len(content) < 8 {
		return errors.New("图片文件太小，无法识别")
	}

	// 检查常见的图片文件头
	// JPEG: FF D8 FF
	if content[0] == 0xFF && content[1] == 0xD8 && content[2] == 0xFF {
		return nil
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if content[0] == 0x89 && content[1] == 0x50 && content[2] == 0x4E && content[3] == 0x47 &&
		content[4] == 0x0D && content[5] == 0x0A && content[6] == 0x1A && content[7] == 0x0A {
		return nil
	}

	// GIF: "GIF87a" 或 "GIF89a"
	if len(content) >= 6 && string(content[0:3]) == "GIF" {
		if string(content[3:6]) == "87a" || string(content[3:6]) == "89a" {
			return nil
		}
	}

	// WebP: "RIFF" + 4 bytes + "WEBP"
	if len(content) >= 12 && string(content[0:4]) == "RIFF" && string(content[8:12]) == "WEBP" {
		return nil
	}

	return errors.New("无效的图片文件格式")
}

// ValidateBMP 验证BMP文件
func ValidateBMP(content []byte) error {
	// 检查文件大小限制（50MB）
	if err := CheckFileSize(content, 50*1024*1024); err != nil {
		return err
	}

	// 检查BMP文件头
	if len(content) < 2 {
		return errors.New("BMP文件头无效")
	}

	// BMP文件应该以 'BM' 开头
	if content[0] != 'B' || content[1] != 'M' {
		return errors.New("无效的BMP文件")
	}

	return nil
}