package service

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/oschwald/geoip2-golang"
)

var (
	geoipDB *geoip2.Reader
)

// InitGeoIP 初始化 GeoIP 数据库
func InitGeoIP() error {
	// 优先使用当前工作目录（支持 go run 和编译后的可执行文件）
	baseDir := "."

	// 尝试从多个位置查找 GeoIP 数据库文件
	dbPaths := []string{
		filepath.Join(baseDir, "data", "GeoLite2-City.mmdb"),
		filepath.Join(baseDir, "GeoLite2-City.mmdb"),
		"./data/GeoLite2-City.mmdb",
		"./GeoLite2-City.mmdb",
		"/usr/share/GeoIP/GeoLite2-City.mmdb",
		"/var/lib/GeoIP/GeoLite2-City.mmdb",
	}

	// 如果上述路径都找不到，尝试使用可执行文件所在目录
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		dbPaths = append(dbPaths,
			filepath.Join(execDir, "data", "GeoLite2-City.mmdb"),
			filepath.Join(execDir, "GeoLite2-City.mmdb"),
		)
	}

	var dbPath string
	for _, path := range dbPaths {
		if _, err := os.Stat(path); err == nil {
			dbPath = path
			break
		}
	}

	if dbPath == "" {
		// 如果找不到数据库文件，创建一个空的 reader
		// 这样可以避免程序崩溃，但无法提供地理位置信息
		return fmt.Errorf("GeoIP database file not found. Please download GeoLite2-City.mmdb from MaxMind and place it in the data/ directory")
	}

	db, err := geoip2.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open GeoIP database: %w", err)
	}

	geoipDB = db
	return nil
}

// GetLocationFromIP 根据 IP 地址获取地理位置信息
func GetLocationFromIP(ipStr string) (country, city, region string, err error) {
	if geoipDB == nil {
		// 如果 GeoIP 数据库未初始化，返回默认值
		return "", "", "", nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", "", "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	record, err := geoipDB.City(ip)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to lookup IP: %w", err)
	}

	// 获取国家名称
	country = ""
	if record.Country.Names != nil {
		if name, ok := record.Country.Names["zh-CN"]; ok {
			country = name
		} else if name, ok := record.Country.Names["en"]; ok {
			country = name
		}
	}

	// 获取城市名称
	city = ""
	if record.City.Names != nil {
		if name, ok := record.City.Names["zh-CN"]; ok {
			city = name
		} else if name, ok := record.City.Names["en"]; ok {
			city = name
		}
	}

	// 获取地区/省份名称
	region = ""
	if len(record.Subdivisions) > 0 {
		if record.Subdivisions[0].Names != nil {
			if name, ok := record.Subdivisions[0].Names["zh-CN"]; ok {
				region = name
			} else if name, ok := record.Subdivisions[0].Names["en"]; ok {
				region = name
			}
		}
	}

	return country, city, region, nil
}

// CloseGeoIP 关闭 GeoIP 数据库
func CloseGeoIP() error {
	if geoipDB != nil {
		return geoipDB.Close()
	}
	return nil
}

// GetGeoIPDBPath 获取 GeoIP 数据库文件路径
func GetGeoIPDBPath() string {
	// 创建 data 目录（如果不存在）
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return ""
	}

	return filepath.Join(dataDir, "GeoLite2-City.mmdb")
}

// GetClientIP 从请求中获取客户端真实 IP 地址
// 优先级：X-Forwarded-For > X-Real-IP > RemoteAddr
func GetClientIP(remoteAddr, forwardedFor, realIP string) string {
	ip := remoteAddr

	// 优先使用 X-Forwarded-For
	if forwardedFor != "" {
		// X-Forwarded-For 可能包含多个 IP，取第一个
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
		}
	} else if realIP != "" {
		// 其次使用 X-Real-IP
		ip = realIP
	}

	// 提取纯IP地址，去掉端口号
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		// 检查是否为 IPv6 地址
		if strings.Contains(ip, "[") {
			// IPv6 格式：[::1]:port 或 [2001:db8::1]:port
			if strings.HasPrefix(ip, "[") {
				endIdx := strings.Index(ip, "]")
				if endIdx != -1 {
					ip = ip[1:endIdx]
				}
			}
		} else {
			// IPv4 格式：192.168.1.1:port
			ip = ip[:idx]
		}
	}

	return ip
}

// IsLocalIP 判断是否为本地 IP 地址
func IsLocalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// 检查是否为 IPv4 本地地址
	if ip4 := ip.To4(); ip4 != nil {
		// 127.0.0.0/8 (loopback)
		if ip4[0] == 127 {
			return true
		}
		// 10.0.0.0/8 (private)
		if ip4[0] == 10 {
			return true
		}
		// 172.16.0.0/12 (private)
		if ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31 {
			return true
		}
		// 192.168.0.0/16 (private)
		if ip4[0] == 192 && ip4[1] == 168 {
			return true
		}
		// 169.254.0.0/16 (link-local)
		if ip4[0] == 169 && ip4[1] == 254 {
			return true
		}
	} else {
		// 检查 IPv6 本地地址
		// ::1 (loopback)
		if ip.IsLoopback() {
			return true
		}
		// fc00::/7 (unique local)
		if ip[0] == 0xfc || ip[0] == 0xfd {
			return true
		}
		// fe80::/10 (link-local)
		if ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
			return true
		}
	}

	return false
}