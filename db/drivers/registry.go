package drivers

import (
	"errors"
	"sync"
)

var (
	drivers     = make(map[string]Driver)
	driversLock sync.RWMutex
)

// RegisterDriver 注册数据库驱动
func RegisterDriver(name string, driver Driver) {
	driversLock.Lock()
	defer driversLock.Unlock()
	
	if driver == nil {
		panic("driver cannot be nil")
	}
	
	if _, dup := drivers[name]; dup {
		panic("driver already registered: " + name)
	}
	
	drivers[name] = driver
}

// GetDriver 获取驱动
func GetDriver(name string) (Driver, error) {
	driversLock.RLock()
	defer driversLock.RUnlock()
	
	driver, ok := drivers[name]
	if !ok {
		return nil, errors.New("driver not found: " + name)
	}
	
	return driver, nil
}

// AvailableDrivers 获取可用驱动列表
func AvailableDrivers() []string {
	driversLock.RLock()
	defer driversLock.RUnlock()
	
	list := make([]string, 0, len(drivers))
	for name := range drivers {
		list = append(list, name)
	}
	
	return list
}
