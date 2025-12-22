// Package store 各种存储方案的实现逻辑
package store

import (
	"data-store/entity/model"
	"etrpc-go/config"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"trpc.group/trpc-go/trpc-go/log"
)

// PluginConfig 存储插件启用配置
type PluginConfig struct {
	Plugins []PlgConfig `yaml:"plugins"`
}

// PlgConfig 存储插件配置信息
type PlgConfig struct {
	Name  string    `yaml:"name"`  // 名称,用于标识一条存储方案
	Type  string    `yaml:"type"`  // 存储插件类型，需要先注册
	Async bool      `yaml:"async"` // 异步写入
	Extra yaml.Node `yaml:"extra"` // 额外配置信息,插件可以自定义
}

// IStorePlugin 数据存储接口
type IStorePlugin interface {
	// Setup 构建一个存储插件实例
	Setup(cfg PlgConfig) (IStorePlugin, error)
	// Write 向存储插件实例写入测点数据
	Write([]*model.OriginPointMsg)
	// Close 关闭存储通道
	Close() error
}

// PluginInstance 存储插件实例
type PluginInstance struct {
	cfg      PlgConfig
	instance IStorePlugin
}

var (
	storePlugins    = make(map[string]IStorePlugin) // 存储插件
	storePluginLock sync.RWMutex                    // 存储插件注册锁

	pluginConfig    = PluginConfig{}            // 存储插件启用配置
	pluginInstances = make([]PluginInstance, 0) // 启用的存储插件实例
	initOnce        sync.Once                   // 存储插件实例化Once锁
	wg              = &sync.WaitGroup{}
)

func init() {
	config.RegisterConfigWithPrefix("store.plugins", "store", &pluginConfig, false)
}

// Register 注册存储插件
func Register(storeType string, plugin IStorePlugin) {
	storePluginLock.Lock()
	defer storePluginLock.Unlock()
	storePlugins[storeType] = plugin
}

// Init 初始化所有启用的存储插件
func Init() {
	initOnce.Do(func() {
		// 按Order排序
		plugins := pluginConfig.Plugins
		for _, pluginCfg := range plugins {
			if plugin, ok := storePlugins[pluginCfg.Type]; ok {
				pluginInstance, err := plugin.Setup(pluginCfg)
				if err != nil {
					panic(errors.Wrapf(err, "store plugin [Name=%s,type=%s] setup failed", pluginCfg.Name, pluginCfg.Type))
				}
				pluginInstances = append(pluginInstances, PluginInstance{
					cfg:      pluginCfg,
					instance: pluginInstance,
				})
				log.Infof("store plugin [%s] setup success", pluginCfg.Name)
			} else {
				panic(fmt.Errorf("store plugin [Name=%s] setup failed, plugin [type=%s] not exist", pluginCfg.Name, pluginCfg.Type))
			}
		}
	})
}

// BatchWritePoint 批量存储测点
func BatchWritePoint(points []*model.OriginPointMsg) {
	innerWg := &sync.WaitGroup{}
	for _, plugin := range pluginInstances {
		// 异步写,不关心是否写入成功,无需等待写入完成
		if plugin.cfg.Async {
			wg.Add(1)
			go func() {
				defer wg.Done()
				plugin.instance.Write(points)
			}()
		} else {
			// 不同插件并发写入,等待所以插件写入完成
			innerWg.Add(1)
			go func() {
				defer innerWg.Done()
				plugin.instance.Write(points)
			}()
		}
	}
	innerWg.Wait()
}

// Close 关闭所有存储通道
func Close() {
	wg.Wait()
	for _, plugin := range pluginInstances {
		err := plugin.instance.Close()
		if err != nil {
			log.Warnf("plugin [%s] close fail, err: %v", plugin.cfg.Name, err)
		} else {
			log.Infof("plugin [%s] close success", plugin.cfg.Name)
		}
	}
}
