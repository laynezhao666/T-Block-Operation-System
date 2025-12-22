// Package cache config cache
package cache

import (
	"etrpc-go/config/util"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"reflect"
	"sync"
	"trpc.group/trpc-go/trpc-go/log"
)

// ConfigCachePath 本地配置缓存地址
var ConfigCachePath = defaultConfigCachePath

const (
	defaultConfigCachePath = "./tmp_etrpc_config.yaml"
)

// UpdateCallback 配置更新的回调函数
type UpdateCallback = func(oldVal, newVal any)

// InitCallback 配置初始化的回调函数
type InitCallback = func(val any)

// ConfigItem 注册配置项
type ConfigItem struct {
	Name       string         // 配置名称
	initObj    any            // 原始值,保存注册时的配置值
	CfgObj     any            // 配置对象,保存用户配置原始对象
	Prefix     string         // 配置前缀
	InitFunc   InitCallback   // 初始化CB
	HotUpdate  bool           // 是否热更新
	UpdateFunc UpdateCallback // 更新时CB

}

var (
	registerMux   sync.Mutex                     // 注册配置项锁，避免并发写map
	refreshCfgMux sync.Mutex                     // 刷新配置的锁,防止并发刷新配置出现异常
	cfgMap        map[string]any                 // 配置map格式，用于根据key获取配置
	cfgYaml       []byte                         // 配置yaml格式，用于根据对象进行反序列化读取配置
	registerMap   = make(map[string]*ConfigItem) // 注册列表
	isLoaded      = false                        // 配置是否已加载
)

// RefreshConfig 配置变化时,刷新缓存的配置信息
//
//	@param newCfgMap	新的配置map
//	@return error	    刷新过程中的错误信息
func RefreshConfig(newCfgMap map[string]any) error {
	refreshCfgMux.Lock()
	defer refreshCfgMux.Unlock()
	newCfgYaml, err := yaml.Marshal(newCfgMap)
	if err != nil {
		return errors.Wrapf(err, "更新配置缓存错误,Map配置转Yaml错误")
	}
	cfgYaml = newCfgYaml
	cfgMap = newCfgMap
	writeCfgFile(newCfgYaml)
	for name, cfgItem := range registerMap {
		if !isLoaded || cfgItem.HotUpdate {
			cfg := cfgYaml
			// 前缀不为空,读取出指定前缀的配置内容
			if len(cfgItem.Prefix) > 0 {
				cfg, err = GetSubCfgYaml(cfgItem.Prefix)
				if err != nil {
					return errors.Wrapf(err, "更新配置对象[%s]失败,注册配置前缀为[%s],前缀不允许指向具体值", name, cfgItem.Prefix)
				}
			}
			// 复制出原始值,用于UpdateFunc
			oldValue := cfgItem.CfgObj
			if cfgItem.UpdateFunc != nil {
				oldValue = clone(cfgItem.CfgObj)
			}
			// 从用户传递的原始对象中创建新对象,用于解析配置
			newValue := clone(cfgItem.initObj)
			if err = yaml.Unmarshal(cfg, newValue); err != nil {
				return errors.Wrapf(err, "更新配置对象[%s]失败,Map配置转Yaml错误", name)
			}
			// 将新的值设置到原来的值上,即*cfgItem.CfgObj = *newValue
			reflect.ValueOf(cfgItem.CfgObj).Elem().Set(reflect.ValueOf(newValue).Elem())
			// 首次加载,InitFunc不为空,调用InitCB
			if !isLoaded && cfgItem.InitFunc != nil {
				cfgItem.InitFunc(cfgItem.CfgObj)
			}
			// 调用更新CB
			if cfgItem.UpdateFunc != nil {
				cfgItem.UpdateFunc(oldValue, cfgItem.CfgObj)
			}
		}
	}
	isLoaded = true
	log.Infof("cache all config success, all config write to %s", ConfigCachePath)
	return nil
}

// Register 注册一个加载项
func Register(item *ConfigItem) {
	registerMux.Lock()
	defer registerMux.Unlock()
	objType := reflect.TypeOf(item.CfgObj)
	if objType.Kind() != reflect.Ptr || objType.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("注册配置对象错误,配置名称=%s,配置对象仅允许为Struct指针类型", item.Name))
	}
	if _, ok := registerMap[item.Name]; ok {
		panic(fmt.Sprintf("注册配置对象错误,配置名称=%s,重复注册", item.Name))
	}
	item.initObj = clone(item.CfgObj)
	registerMap[item.Name] = item
}

// GetCfgMap 获取当前缓存的配置map
func GetCfgMap() map[string]any {
	return cfgMap
}

// GetCfgYaml 获取当前缓存的配置yaml
func GetCfgYaml() []byte {
	return cfgYaml
}

func writeCfgFile(yamlCfg []byte) {
	if err := os.WriteFile(GetConfigCachePath(), yamlCfg, 0777); err != nil {
		log.Warn("写入配置到本地文件错误，错误信息为：", err.Error())
	}
}

// GetSubCfgYaml 根据key获取子yaml结构,请勿获取yaml的叶子节点
func GetSubCfgYaml(key string) ([]byte, error) {
	subMap, ok := util.GetByKey(cfgMap, key)
	if !ok {
		return []byte{}, nil
	}
	subYaml, err := yaml.Marshal(subMap)
	if err != nil {
		return nil, err
	}
	return subYaml, nil
}

// GetConfigCachePath 本地配置缓存地址
func GetConfigCachePath() string {
	if ConfigCachePath == defaultConfigCachePath && !flag.Parsed() {
		flag.StringVar(&ConfigCachePath, "conf_cache", defaultConfigCachePath, "cache config path")
		flag.Parse()
	}
	return ConfigCachePath
}

// Clone 使用反射克隆任意结构体指针
func clone(src interface{}) interface{} {
	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() != reflect.Ptr || srcValue.IsNil() {
		return nil
	}

	dstValue := reflect.New(srcValue.Elem().Type()).Elem()
	dstValue.Set(reflect.ValueOf(src).Elem())

	return dstValue.Addr().Interface()
}
