package mesql

import (
	"time"
)

var (
	// 默认最大连接 50
	DefaultMaxOpenConns = 50
	// 默认最大空闲连接10
	DefaultMaxIdleConns = 10
	// 默认最大连接失效 30分钟
	DefaultConnMaxLifetime = 30 * time.Minute
	// 默认解析的tag
	DefaultTagName = "db"
)