// Package config collector配置
package config

import (
	"etrpc-go/config"
)

const (
	clientConfigPrefix       string = "client"
	featuresConfigPrefix     string = "features"
	commonReportConfigPrefix string = "common_report"
	CollectorSendType        string = "collector"
	KafkaSendType            string = "kafka"
)

var (
	featuresConf     = &FeaturesConf{}
	clientConf       = &ClientConf{}
	commonReportConf = &CommonReportMetrics{}
)

// FeaturesConf collector额外配置
type FeaturesConf struct {
	SendType string `yaml:"send_type"`
	Trace    bool   `yaml:"trace"`
	Gray     bool   `yaml:"gray"`
}

// ClientConf 客户端配置，主要用于配置数据上报的服务名
type ClientConf struct {
	Service []ClientServiceConf `yaml:"service"`
}

// ClientServiceConf 客户端配置内容
type ClientServiceConf struct {
	Name   string `yaml:"name"`
	Target string `yaml:"target"`
}

// CommonReportMetrics 通用上报配置
type CommonReportMetrics struct {
	DimensionMap map[string]string `yaml:"dimension_map"` // 维度名和维度值
}

func init() {
	config.RegisterConfigWithPrefix(featuresConfigPrefix, featuresConfigPrefix, featuresConf, false)
	config.RegisterConfigWithPrefix(clientConfigPrefix, clientConfigPrefix, clientConf, true)
	config.RegisterConfigWithPrefix(commonReportConfigPrefix, commonReportConfigPrefix, commonReportConf, true)
}

// GetClientConf 获取客户端配置
func GetClientConf() *ClientConf {
	return clientConf
}

// GetFeaturesConf 获取collector额外配置
func GetFeaturesConf() *FeaturesConf {
	return featuresConf
}

// GetCommonReportConf 获取通用的指标上报配置
func GetCommonReportConf() *CommonReportMetrics {
	return commonReportConf
}
