// Package config is the config package of etrpc-go
package config

import (
	"etrpc-go/config/cache"
	"etrpc-go/config/util"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"strconv"
	"strings"
	"time"
)

// Option sets config options.
type Option func(item *cache.ConfigItem)

// WithHotUpdate set hot update
func WithHotUpdate(hotUpdate bool) Option {
	return func(item *cache.ConfigItem) {
		item.HotUpdate = hotUpdate
	}
}

// WithPrefix  set config prefix
func WithPrefix(prefix string) Option {
	return func(item *cache.ConfigItem) {
		item.Prefix = prefix
	}
}

// WithInitFunc set init callback func
func WithInitFunc(callback cache.InitCallback) Option {
	return func(item *cache.ConfigItem) {
		item.InitFunc = callback
	}
}

// WithUpdateFunc set hot update callback func
func WithUpdateFunc(callback cache.UpdateCallback) Option {
	return func(item *cache.ConfigItem) {
		item.UpdateFunc = callback
	}
}

// Register 注册配置
func Register(cfgName string, cfgObj any, opts ...Option) {
	cfgItem := &cache.ConfigItem{
		Name:   cfgName,
		CfgObj: cfgObj,
	}
	for _, opt := range opts {
		opt(cfgItem)
	}
	cache.Register(cfgItem)
}

// RegisterConfig 在配置加载前注册配置对象(通过init函数),这些对象会在配置首次加载后初始化,配置刷新时注册的对象也会刷新
//
//	@param cfgName		需要加载配置的对象的唯一标识
//	@param cfgObj		需要加载配置的对象,需要传递Struct对象指针
//	@param hotUpdate    是否需要热更新
func RegisterConfig(cfgName string, cfgObj any, hotUpdate bool) {
	Register(cfgName, cfgObj, WithHotUpdate(hotUpdate))
}

// RegisterConfigWithPrefix 在配置加载前注册配置对象(通过init函数),这些对象会在配置首次加载后初始化,配置刷新时注册的对象也会刷新
//
//	@param cfgName		需要加载配置的对象的唯一标识
//	@param prefix		需要加载配置的对象的配置前缀,如前缀为a.b,则只需定义a.b下配置的结构体即可
//	@param cfgObj		需要加载配置的对象,需要传递Struct对象指针
//	@param hotUpdate    是否需要热更新
func RegisterConfigWithPrefix(cfgName, prefix string, cfgObj any, hotUpdate bool) {
	Register(cfgName, cfgObj, WithHotUpdate(hotUpdate), WithPrefix(prefix))
}

// Load 使用yaml配置对象加载配置
//
//	@param cfg 		需要加载配置的对象,需要传递Struct对象指针
//	@return error	加载错误信息
func Load(cfg any) error {
	return yaml.Unmarshal(cache.GetCfgYaml(), cfg)
}

// LoadWithPrefix 使用yaml配置对象加载配置
//
//	@param cfg 		需要加载配置的对象,需要传递Struct对象指针
//	@param prefix 	配置指定前缀
//	@return error	加载错误信息
func LoadWithPrefix(cfg any, prefix string) error {
	cfgYaml, err := cache.GetSubCfgYaml(prefix)
	if err != nil {
		return errors.Wrapf(err, "注册配置前缀为[%s],前缀不允许指向具体值", prefix)
	}
	return yaml.Unmarshal(cfgYaml, cfg)
}

// Get 获取配置项，key格式为a.b.c
//
//	@param key		格式为a.b.c
//	@return any		key对应的配置项，如果传递空字符串，返回整个配置项
//	@return bool	key是否存在
func Get(key string) (any, bool) {
	return util.GetByKey(cache.GetCfgMap(), key)
}

// GetOrDefault 获取配置项，如果key不存在，返回默认值
//
//	@param key			key格式为a.b.c
//	@param defaultVal	key不存在时返回的默认值
//	@return any			key对应的配置项，如果传递空字符串，返回整个配置项
func GetOrDefault(key string, defaultVal any) any {
	if val, ok := util.GetByKey(cache.GetCfgMap(), key); ok {
		return val
	}
	return defaultVal
}

// GetInt32 获取Int32配置项
//
//	@param key		key格式为a.b.c
//	@return int32	key对应的值
//	@return bool	key是否存在,类型错误时也返回false
func GetInt32(key string) (int32, bool) {
	if val, ok := Get(key); ok {
		if intVal, ok := val.(int32); ok {
			return intVal, true
		}
		if intVal, ok := val.(int); ok {
			return int32(intVal), true
		}
		if intStr, ok := val.(string); ok {
			if intVal, err := strconv.Atoi(intStr); err == nil {
				return int32(intVal), true
			}
		}
	}
	return 0, false
}

// GetInt32OrDefault 获取Int32配置项，如果key不存在，返回默认值
//
//	@param key			key格式为a.b.c
//	@param defaultVal	key不存在时返回的默认值
//	@return int32		key对应的值,类型错误时返回默认值
func GetInt32OrDefault(key string, defaultVal int32) int32 {
	if intVal, ok := GetInt32(key); ok {
		return intVal
	}
	return defaultVal
}

// GetInt64 获取Int64配置项
//
//	@param key			key格式为a.b.c
//	@return int64	    key对应的值
//	@return bool	    key是否存在,类型错误时也返回false
func GetInt64(key string) (int64, bool) {
	if val, ok := Get(key); ok {
		if intVal, ok := val.(int64); ok {
			return intVal, true
		}
		if intVal, ok := val.(int); ok {
			return int64(intVal), true
		}
		if intVal, ok := val.(int32); ok {
			return int64(intVal), true
		}
		if intStr, ok := val.(string); ok {
			if intVal, err := strconv.ParseInt(intStr, 10, 0); err == nil {
				return intVal, true
			}
		}
	}
	return 0, false
}

// GetInt64OrDefault 获取Int64配置项，如果key不存在，返回默认值
//
//	@param key			key格式为a.b.c
//	@param defaultVal	key不存在时返回的默认值
//	@return int64		key对应的值
func GetInt64OrDefault(key string, defaultVal int64) int64 {
	if intVal, ok := GetInt64(key); ok {
		return intVal
	}
	return defaultVal
}

// GetFloat32 获取Float32配置项
//
//	@param key			key格式为a.b.c
//	@return float32		key对应的值
//	@return bool		key是否存在,类型错误时也返回false
func GetFloat32(key string) (float32, bool) {
	if val, ok := Get(key); ok {
		if floatVal, ok := val.(float32); ok {
			return floatVal, true
		}
		if floatVal, ok := val.(float64); ok {
			return float32(floatVal), true
		}
		if floatStr, ok := val.(string); ok {
			if floatVal, err := strconv.ParseFloat(floatStr, 32); err == nil {
				return float32(floatVal), true
			}
		}
	}
	return 0, false
}

// GetFloat32OrDefault 获取Float32配置项，如果key不存在或类型错误，返回默认值
//
//	@param key			key格式为a.b.c
//	@param defaultVal	key不存在时返回的默认值
//	@return float32		key对应的值，如果key不存在或类型错误，返回默认值
func GetFloat32OrDefault(key string, defaultVal float32) float32 {
	if floatVal, ok := GetFloat32(key); ok {
		return floatVal
	}
	return defaultVal
}

// GetFloat64 获取Float64配置项
//
//	@param key			key格式为a.b.c
//	@return float64		key对应的值
//	@return bool		key是否存在,类型错误时也返回false
func GetFloat64(key string) (float64, bool) {
	if val, ok := Get(key); ok {
		if floatVal, ok := val.(float64); ok {
			return floatVal, true
		}
		if floatVal, ok := val.(float32); ok {
			return float64(floatVal), true
		}
		if floatStr, ok := val.(string); ok {
			if floatVal, err := strconv.ParseFloat(floatStr, 64); err == nil {
				return floatVal, true
			}
		}
	}
	return 0, false
}

// GetFloat64OrDefault 获取Float64配置项，如果key不存在或类型错误，返回默认值
//
//	@param key				key格式为a.b.c
//	@param defaultVal		key不存在时返回的默认值
//	@return float64			key对应的值，如果key不存在或类型错误，返回默认值
func GetFloat64OrDefault(key string, defaultVal float64) float64 {
	if floatVal, ok := GetFloat64(key); ok {
		return floatVal
	}
	return defaultVal
}

// GetString 获取String配置项,如果key存在但类型不为string,则强制转化为string
//
//	@param key				key格式为a.b.c
//	@return string			key对应的值
//	@return bool			key是否存在
func GetString(key string) (string, bool) {
	if val, ok := Get(key); ok {
		if str, ok := val.(string); ok {
			return str, true
		}
		if val == nil {
			return "", true
		}
		return fmt.Sprint(val), true
	}
	return "", false
}

// GetStringOrDefault 获取String配置项，key存在但类型不为string,则强制转化为string,key不存在则返回默认值
//
//	@param key				key格式为a.b.c
//	@param defaultVal		key不存在时返回的默认值
//	@return string			key对应的值
func GetStringOrDefault(key string, defaultVal string) string {
	if val, ok := GetString(key); ok {
		return val
	}
	return defaultVal
}

// GetBool 获取Bool配置项
//
//	@param key				key格式为a.b.c
//	@return bool			key对应的值
//	@return bool			key是否存在,如果key不存在或者类型错误，则返回false
func GetBool(key string) (bool, bool) {
	if val, ok := Get(key); ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal, true
		}
		if boolStr, ok := val.(string); ok {
			if boolVal, err := strconv.ParseBool(boolStr); err == nil {
				return boolVal, true
			}
		}
	}
	return false, false
}

// GetBoolOrDefault 获取Bool配置项,如果key不存在或者类型错误，则返回默认值
//
//	@param key				key格式为a.b.c
//	@param defaultVal		key不存在时返回的默认值
//	@return bool			key对应的值
func GetBoolOrDefault(key string, defaultVal bool) bool {
	if val, ok := GetBool(key); ok {
		return val
	}
	return defaultVal
}

// GetTime 获取Time配置项
//
//	@param key				key格式为a.b.c
//	@return time.Time		key对应的值
//	@return bool			key是否存在,如果key不存在或者类型错误，则返回false
func GetTime(key string) (time.Time, bool) {
	if val, ok := Get(key); ok {
		if dateVal, ok := val.(time.Time); ok {
			return dateVal, true
		}
		if dateStr, ok := val.(string); ok {
			dateStr = strings.Trim(dateStr, " \r\n\t")
			if dateVal, err := time.Parse(time.DateOnly, dateStr); err == nil {
				return dateVal, true
			}
			if dateVal, err := time.Parse(time.DateTime, dateStr); err == nil {
				return dateVal, true
			}
		}
	}
	return time.Time{}, false
}

// GetTimeOrDefault 获取Time配置项,如果key不存在或者类型错误，则返回默认值
//
//	@param key				key格式为a.b.c
//	@param defaultVal		key不存在时返回的默认值
//	@return time.Time		key对应的值
func GetTimeOrDefault(key string, defaultVal time.Time) time.Time {
	if val, ok := GetTime(key); ok {
		return val
	}
	return defaultVal
}
