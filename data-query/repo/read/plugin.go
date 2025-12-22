package read

import (
	"context"
	"data-query/entity"
	"etrpc-go/config"
	"fmt"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config 插件配置
type Config struct {
	Plugins []PlgConfig `yaml:"plugins"`
}

// PlgConfig 存储插件配置信息
type PlgConfig struct {
	Name  string    `yaml:"name"`  // 名称,用于标识一条存储方案
	Type  string    `yaml:"type"`  // 存储插件类型，需要先注册
	Order int32     `yaml:"order"` // 优先级,越小优先级越高
	Extra yaml.Node `yaml:"extra"` // 额外配置信息,插件可以自定义
}

// IReadPlugin 数据查询接口
type IReadPlugin interface {
	// Setup 构建一个查询插件
	Setup(cfg PlgConfig) (IReadPlugin, error)
	// GetType 判断插件类型
	GetType() string
	// CanRead 判断当前时间范围能否查询
	CanRead(begin, end int64) bool
	// ReadRange 查询时间范围数据
	ReadRange(ctx context.Context, pointName []string, begin, end int64) (map[string][]*entity.Point, error)
	// ReadLatest 查询某个时间点前最新数据
	ReadLatest(ctx context.Context, pointName []string, max int64) (map[string]*entity.Point, error)
	// ReadChanged 查询测点最近变化的时间
	ReadChanged(ctx context.Context, pointName []string, begin int64, end int64) (map[string]int64, error)
}

var (
	readPlugins    = make(map[string]IReadPlugin)
	readPluginLock sync.RWMutex
	pluginConfigs  Config
	enablePlugins  = make([]IReadPlugin, 0)
	initOnce       sync.Once
)

func init() {
	config.RegisterConfigWithPrefix("read_plugin_cfg", "read", &pluginConfigs, false)
}

// Register 注册读取插件
func Register(readType string, plugin IReadPlugin) {
	readPluginLock.Lock()
	defer readPluginLock.Unlock()
	readPlugins[readType] = plugin
}

// Init 初始化所有启用的存储插件
func Init() {
	initOnce.Do(func() {
		// 按Order排序
		plugins := pluginConfigs.Plugins
		sort.Slice(plugins, func(i, j int) bool {
			return plugins[i].Order < plugins[j].Order
		})
		for _, pluginCfg := range plugins {
			if plugin, ok := readPlugins[pluginCfg.Type]; ok {
				pluginInstance, err := plugin.Setup(pluginCfg)
				if err != nil {
					panic(errors.Wrapf(err, "read plugin [name=%s,type=%s] setup failed", pluginCfg.Name, pluginCfg.Type))
				}
				enablePlugins = append(enablePlugins, pluginInstance)
			} else {
				panic(fmt.Errorf("read plugin [name=%s] setup failed, plugin [type=%s] not exist", pluginCfg.Name, pluginCfg.Type))
			}
		}
	})
}

// Get 获取所有启用的存储插件
func Get() []IReadPlugin {
	return enablePlugins
}

// BatchReadRangePoints 按时间范围批量读取数据
func BatchReadRangePoints(ctx context.Context, pointNames []string, begin, end int64) (
	map[string][]*entity.Point, string, error) {
	for _, plugin := range enablePlugins {
		if plugin.CanRead(begin, end) {
			if res, err := plugin.ReadRange(ctx, pointNames, begin, end); err == nil {
				return res, plugin.GetType(), nil
			} else {
				return nil, plugin.GetType(), err
			}
		}
		//return nil, errors.New("given time can not read")
	}
	return nil, "", errors.New("no read plugin found or available")
}

// BatchReadLatestPoint 按时间点批量读取最近数据
func BatchReadLatestPoint(ctx context.Context, pointNames []string, max int64) (map[string]*entity.Point, error) {
	for _, plugin := range enablePlugins {
		if plugin.CanRead(max, max) {
			if res, err := plugin.ReadLatest(ctx, pointNames, max); err == nil {
				return res, nil
			}
		}
	}
	return nil, errors.New("no read plugin found or available")
}

// BatchReadChangedPoint 查询测点最近变化的时间
func BatchReadChangedPoint(ctx context.Context, pointNames []string, begin int64, end int64) (map[string]int64, error) {
	for _, plugin := range enablePlugins {
		if plugin.CanRead(end, end) {
			if res, err := plugin.ReadChanged(ctx, pointNames, begin, end); err == nil {
				return res, nil
			}
		}
	}
	return nil, errors.New("no read plugin found or available")
}
