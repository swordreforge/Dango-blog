package validation

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ValidateZIP 验证ZIP文件
func ValidateZIP(content []byte) error {
	// 检查文件大小限制（500MB）
	if err := CheckFileSize(content, 500*1024*1024); err != nil {
		return err
	}

	// 检查ZIP文件头
	if len(content) < 4 || content[0] != 0x50 || content[1] != 0x4B ||
		content[2] != 0x03 || content[3] != 0x04 {
		return errors.New("无效的ZIP文件")
	}

	// 解压并检查内容
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return errors.New("解析ZIP文件失败")
	}

	// 检查压缩包中的文件
	for _, file := range reader.File {
		// 检查文件名是否包含路径遍历
		if strings.Contains(file.Name, "..") {
			return fmt.Errorf("ZIP文件包含路径遍历: %s", file.Name)
		}

		// 检查文件名是否以 / 开头（绝对路径）
		if strings.HasPrefix(file.Name, "/") {
			return fmt.Errorf("ZIP文件包含绝对路径: %s", file.Name)
		}
	}

	return nil
}

// ValidateTarGz 验证TAR.GZ文件，防止软链接
func ValidateTarGz(content []byte, ext string) error {
	// 检查文件大小限制（500MB）
	if err := CheckFileSize(content, 500*1024*1024); err != nil {
		return err
	}

	var reader io.Reader = bytes.NewReader(content)

	// 如果是.gz文件，先解压
	if ext == ".gz" || ext == ".tar.gz" {
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return errors.New("解压GZIP文件失败")
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// 解析TAR文件
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New("解析TAR文件失败")
		}

		// 检查文件名是否包含路径遍历
		if strings.Contains(header.Name, "..") {
			return fmt.Errorf("TAR文件包含路径遍历: %s", header.Name)
		}

		// 检查是否是软链接
		if header.Typeflag == tar.TypeSymlink {
			return fmt.Errorf("TAR文件包含软链接: %s", header.Name)
		}

		// 检查是否是硬链接
		if header.Typeflag == tar.TypeLink {
			return fmt.Errorf("TAR文件包含硬链接: %s", header.Name)
		}

		// 检查文件名是否以 / 开头（绝对路径）
		if strings.HasPrefix(header.Name, "/") {
			return fmt.Errorf("TAR文件包含绝对路径: %s", header.Name)
		}
	}

	return nil
}

// Validate7z 验证7z文件
func Validate7z(content []byte) error {
	// 检查文件大小限制（500MB）
	if err := CheckFileSize(content, 500*1024*1024); err != nil {
		return err
	}

	// 7z文件头检查
	if len(content) < 6 {
		return errors.New("无效的7z文件")
	}

	// 7z文件以 "7z\xBC\xAF\x27\x1C" 开头
	if string(content[0:2]) != "7z" {
		return errors.New("无效的7z文件")
	}

	// 注意：由于Go标准库不支持7z格式，这里只做基本验证
	// 实际使用时可以集成第三方库如 github.com/bodgit/sevenzip

	return nil
}

// ValidateRAR 验证RAR文件
func ValidateRAR(content []byte) error {
	// 检查文件大小限制（500MB）
	if err := CheckFileSize(content, 500*1024*1024); err != nil {
		return err
	}

	// RAR文件头检查
	if len(content) < 7 {
		return errors.New("无效的RAR文件")
	}

	// RAR文件以 "Rar!" 开头 (0x52 0x61 0x72 0x21 0x1A 0x07)
	if string(content[0:4]) != "Rar!" {
		return errors.New("无效的RAR文件")
	}

	// 注意：由于RAR是专有格式，这里只做基本验证
	// 实际使用时可以集成第三方库如 github.com/nwaples/rardecode

	return nil
}