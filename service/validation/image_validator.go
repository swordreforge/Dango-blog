package validation

import (
	"errors"
)

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