// Package metric provide api to simply metric report
package metric

import (
	"fmt"
	"sync"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/metrics"
)

// IMetric 指标上报接口
type IMetric interface {
	// Report 上报指定值v，按照设定的聚合policy自动聚合
	Report(v float64)
	// ReportWithDim 上报指定值v，并携带额外的维度信息
	ReportWithDim(v float64, dims map[string]string)
	// ReportBatch 批量上报多个值，按照设定的聚合policy自动聚合
	ReportBatch(v []float64)
	// ReportBatchWithDim 批量上报多个值，并携带额外的维度信息
	ReportBatchWithDim(v []float64, dims map[string]string)
}

// 指标实例对象
type metric struct {
	name string   // 指标名称
	opt  *options // 指标相关配置
}

// Report 上报指定值v
func (m *metric) Report(v float64) {
	m.ReportBatchWithDim([]float64{v}, nil)
}

// ReportWithDim 上报指定值v，并携带额外的维度信息
func (m *metric) ReportWithDim(v float64, dims map[string]string) {
	m.ReportBatchWithDim([]float64{v}, dims)
}

// ReportBatch 批量上报多个值
func (m *metric) ReportBatch(vals []float64) {
	m.ReportBatchWithDim(vals, nil)
}

// ReportBatchWithDim 批量上报多个值，并携带额外的维度信息
func (m *metric) ReportBatchWithDim(vals []float64, dims map[string]string) {
	if len(vals) == 0 {
		return
	}
	m.opt.once.Do(func() {
		if !m.opt.disableDefaultDims {
			// 放这里初始化是因为，如果Metric被定义成全局变量，trpc的配置还未初始化
			cfg := trpc.GlobalConfig()
			m.opt.comDims = append(m.opt.comDims, []*metrics.Dimension{
				{Name: "server", Value: cfg.Server.Server},
				{Name: "env", Value: cfg.Global.EnvName},
				{Name: "ip", Value: cfg.Global.LocalIP},
				{Name: "container", Value: cfg.Global.ContainerName},
			}...)
			if cfg.Global.EnableSet == "Y" {
				m.opt.comDims = append(m.opt.comDims, &metrics.Dimension{Name: "set_name", Value: cfg.Global.FullSetName})
			}
		}
	})
	dimensions := make([]*metrics.Dimension, 0, len(m.opt.comDims)+len(dims))
	dimensions = append(dimensions, m.opt.comDims...)
	for k, val := range dims {
		dimensions = append(dimensions, &metrics.Dimension{Name: k, Value: val})
	}
	allMetrics := make([]*metrics.Metrics, 0, len(m.opt.policy))
	for _, p := range m.opt.policy {
		var metricName string
		if m.opt.nameConvertFunc != nil {
			metricName = m.opt.nameConvertFunc(m.name, p)
		} else {
			metricName = defaultNameConvertFunc(m.name, p)
		}
		for _, v := range vals {
			allMetrics = append(allMetrics, metrics.NewMetrics(metricName, v, p))
		}
	}
	rec := metrics.NewMultiDimensionMetricsX(m.opt.metricGroup, dimensions, allMetrics)
	_ = metrics.Report(rec)
}

// NameConvertFunc 自定义指标名称转换函数
type NameConvertFunc = func(string, metrics.Policy) string

// 默认的指标名称转换函数
var defaultNameConvertFunc = func(name string, policy metrics.Policy) string {
	switch policy {
	case metrics.PolicyAVG:
		return fmt.Sprintf("%s_avg", name)
	case metrics.PolicyMAX:
		return fmt.Sprintf("%s_max", name)
	case metrics.PolicyMIN:
		return fmt.Sprintf("%s_min", name)
	case metrics.PolicyMID:
		return fmt.Sprintf("%s_mid", name)
	default:
		return name
	}
}

// SetDefaultNameConvertFunc 设置全局默认的指标名称转换函数
func SetDefaultNameConvertFunc(convertor NameConvertFunc) {
	defaultNameConvertFunc = convertor
}

// NewMetric 新建一个指标
func NewMetric(name string, opts ...Option) IMetric {
	optObj := &options{
		policy: []metrics.Policy{metrics.PolicySUM},
	}
	for _, opt := range opts {
		opt(optObj)
	}
	return &metric{
		name: name,
		opt:  optObj,
	}
}

// options 指标相关配置
type options struct {
	metricGroup        string
	policy             []metrics.Policy
	nameConvertFunc    NameConvertFunc
	disableDefaultDims bool

	once    sync.Once // 用于控制
	comDims []*metrics.Dimension
}

// Option sets metric options.
type Option func(*options)

// WithPolicy 设置聚合策略，默认SUM
func WithPolicy(policy ...metrics.Policy) Option {
	return func(o *options) {
		if len(policy) > 0 {
			o.policy = policy
		}
	}
}

// WithNameConvertFunc 设置指标名称转换函数
func WithNameConvertFunc(convertor NameConvertFunc) Option {
	return func(o *options) {
		o.nameConvertFunc = convertor
	}
}

// WithMetricGroup 设置指标组，默认为空，自动选择配置文件中的默认指标组
func WithMetricGroup(metricGroup string) Option {
	return func(o *options) {
		o.metricGroup = metricGroup
	}
}

// WithDimensions 设置默认的携带的维度信息
func WithDimensions(dims map[string]string) Option {
	dimensions := make([]*metrics.Dimension, 0, len(dims))
	for k, v := range dims {
		dimensions = append(dimensions, &metrics.Dimension{Name: k, Value: v})
	}
	return func(o *options) {
		o.comDims = append(o.comDims, dimensions...)
	}
}

// WithoutDefaultDims 设置无需PodName/Env/IP/App这些维度
func WithoutDefaultDims() Option {
	return func(o *options) {
		o.disableDefaultDims = true
	}
}
