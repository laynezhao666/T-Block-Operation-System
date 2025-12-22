// Package util 用于配置合并/变量替换的工具类
package util

import (
	"etrpc-go/util/arrayutil"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
)

var (
	varRegx = regexp.MustCompile(`\$\{([^\s'"${]*)}`)
)

// ExpandSystemEnv  替换字符串种的环境变量引用，格式${xxx}
//
//	@param replaceBlank	获取环境变量为""，是否替换
func ExpandSystemEnv(s []byte, replaceBlank bool) []byte {
	var buf []byte
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+2 < len(s) && s[j+1] == '{' { // only ${var} instead of $var is valid
			if buf == nil {
				buf = make([]byte, 0, 2*len(s))
			}
			buf = append(buf, s[i:j]...)
			name, w := getEnvName(s[j+1:])
			if name == nil && w > 0 {
				// invalid matching, remove the $
			} else if name == nil {
				buf = append(buf, s[j]) // keep the $
			} else {
				if val := os.Getenv(string(name)); val != "" || replaceBlank {
					buf = append(buf, val...)
				} else {
					buf = append(buf, fmt.Sprintf("${%s}", string(name))...)
				}
			}
			j += w
			i = j + 1
		}
	}
	if buf == nil {
		return s
	}
	return append(buf, s[i:]...)
}

// getEnvName gets env name, that is, var from ${var}.
// The env name and its len will be returned.
func getEnvName(s []byte) ([]byte, int) {
	// look for right curly bracket '}'
	// it's guaranteed that the first char is '{' and the string has at least two char
	for i := 1; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\n' || s[i] == '"' { // "xx${xxx"
			return nil, 0 // encounter invalid char, keep the $
		}
		if s[i] == '}' {
			if i == 1 { // ${}
				return nil, 2 // remove ${}
			}
			return s[1:i], i + 1
		}
	}
	return nil, 0 // no }，keep the $
}

// ExpandEnv 替换配置中的变量引用
//
//	@param cfg			原始配置项,用于递归处理
//	@param rawCfg		原始配置项,用于保存底层配置结构
//	@return error		替换过程中出现的错误,如变量循环依赖、替换后yaml不符合格式要求等
func ExpandEnv(cfg map[string]any) (map[string]any, error) {
	yamlCfg, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "替换配置变量引用错误,map配置转yaml配置错误")
	}
	yamlCfgStr := string(yamlCfg)
	matches := varRegx.FindAllStringSubmatch(yamlCfgStr, -1)
	replacedVars := make(map[string]any)
	for _, match := range matches {
		variable := match[1]
		if _, ok := replacedVars[variable]; ok {
			continue
		}
		varVal, _, err := getVariableVal(cfg, variable, []string{variable})
		if err != nil {
			return nil, errors.Wrapf(err, "替换配置变量引用[%s]错误", variable)
		}
		yamlCfgStr = strings.ReplaceAll(yamlCfgStr, fmt.Sprintf("${%s}", variable), varVal)
		replacedVars[variable] = 1
	}
	replacedCfg := make(map[string]any)
	err = yaml.Unmarshal([]byte(yamlCfgStr), replacedCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "替换配置变量引用错误,替换后yaml格式异常,yaml:\n%s", yamlCfgStr)
	}
	return replacedCfg, nil
}

// getVariableVal 获取配置/环境变量中键为key的值,如果变量引用的是复杂类型(map，slice),则返回空字符串
//
//	@param cfg				所有配置，map格式
//	@param variable			需要查询的配置变量,格式a.b.c
//	@param refPath			变量引用列表,用于判断是否出现循环依赖
//	@return string			变量对应的值,复杂类型返回空字符串，其他类型强转字符串
//	@return bool			对应变量是否存在
//	@return error			查找过程中出现循环依赖错误
func getVariableVal(cfg map[string]any, variable string, refPath []string) (string, bool, error) {
	// 先尝试从环境变量获取
	envVal := os.Getenv(variable)
	if envVal != "" {
		return envVal, true, nil
	}
	// 再从配置中获取
	if val, ok := GetByKey(cfg, variable); ok {
		switch val.(type) {
		// 变量值为string类型,可能还包含变量引用，递归处理
		case string:
			rawStr := val.(string)
			matches := varRegx.FindAllStringSubmatch(rawStr, -1)
			for _, match := range matches {
				refVariable := match[1]
				// 出现循环,直接抛出错误后退出
				if arrayutil.Find(refPath, refVariable) >= 0 {
					return "", false, fmt.Errorf("变量[%s]出现循环依赖", refVariable)
				}
				// 递归处理
				varVal, _, err := getVariableVal(cfg, refVariable, append(refPath, refVariable))
				if err != nil {
					return "", false, err
				}
				// 变量值替换
				rawStr = strings.ReplaceAll(rawStr, fmt.Sprintf("${%s}", refVariable), varVal)
			}
			return rawStr, true, nil
			// 复杂类型，直接返回空字符串
		case map[string]any, []any:
			return "", false, nil
			// 其他类型,直接转化为字符串返回
		default:
			return fmt.Sprint(val), true, nil
		}
	}
	return "", false, nil
}

// GetByKey 从map从查找指定的key的值
//
//	@param cfg			配置map
//	@param key			配置key,格式a.b.c,不支持[0]的方式引用数组中的项
//	@return any			返回对应的值
//	@return bool		key是否存在
func GetByKey(cfg map[string]any, key string) (any, bool) {
	// key为空，返回所有配置
	if len(key) == 0 {
		return cfg, true
	}
	// 对key进行拆分
	keys := strings.Split(key, ".")
	lens := len(keys)
	res := cfg
	// 循环递归查找每段key
	for idx, k := range keys {
		if d, ok := res[k]; ok {
			// key的最后面一层
			if idx == lens-1 {
				return d, true
			}
			// 递归到深层的map
			if res, ok = d.(map[string]any); ok {
				continue
			}
			return nil, false
		}
	}
	return nil, false
}

// MergeAllConfig 合并七彩石配置及本地配置
//
//	@param rainbowCfg			七彩石配置
//	@param localCfg				本地配置
//	@return map[string]any		合并后的配置
func MergeAllConfig(rainbowCfg, localCfg map[string]any) map[string]any {
	mergeCfg := make(map[string]any)
	// 七彩石上etrpc开头的配置优先级最低,后续配置直接覆盖优先级低的配置
	for key, val := range rainbowCfg {
		if strings.HasPrefix(key, "etrpc") {
			mergeRainbowCfg(key, val, mergeCfg)
		}
	}
	// 七彩石上非etrpc开头的配置，优先级次之
	for key, val := range rainbowCfg {
		if !strings.HasPrefix(key, "etrpc") {
			mergeRainbowCfg(key, val, mergeCfg)
		}
	}
	// 本地配置优先级最高
	mergeYamlMap(mergeCfg, localCfg)
	return mergeCfg
}

func mergeRainbowCfg(key string, val any, cfg map[string]any) {
	// 如果val是string类型，则尝试解析成yaml格式,并进行合并
	if valStr, ok := val.(string); ok {
		// 因为json使用yaml也能解析成功，故trim后开头是"{"则认为是json,不进行yaml解析
		if !strings.HasPrefix(strings.Trim(valStr, " \n\t"), "{") {
			yamlMap := map[string]any{}
			if err := yaml.Unmarshal([]byte(valStr), yamlMap); err == nil {
				mergeYamlMap(cfg, yamlMap)
				return
			}
		}
	}
	// 保留七彩石上的KV
	mergeYamlKey(cfg, key, val)
}

// mergeYamlMap
//
//	@Description:		合并
//	@param target
//	@param source
func mergeYamlMap(target, source map[string]any) {
	for sk, sv := range source {
		if tv, ok := target[sk]; ok {
			// 已经存在的且target中对应的key是map类型
			if tvMap, ok := tv.(map[string]any); ok {
				// source中对应key也是map类型,则进行递归合并
				// 否则直接使用source中的val覆盖
				if svMap, ok := sv.(map[string]any); ok {
					mergeYamlMap(tvMap, svMap)
					continue
				}
			}
		}
		// target中对应key不存在,或者不同时为map或者list类型,直接覆盖
		target[sk] = sv
	}
}

// mergeYamlKey
//
//	@Description:    将key-val合并到yaml配置中,覆盖模式
//	@param target	 原始yaml配置,map格式
//	@param key		 key,格式为a.b.c
//	@param val		 key对应的值
func mergeYamlKey(target map[string]any, key string, val any) {
	if len(key) == 0 {
		return
	}
	keys := strings.Split(key, ".")
	lens := len(keys)
	for idx, k := range keys {
		// 最后一层直接将key写入Map
		if idx == lens-1 {
			target[k] = val
			return
		}
		// key存在,且是map类型,则继续合并
		if tv, ok := target[k]; ok {
			if tvMap, ok := tv.(map[string]any); ok {
				target = tvMap
				continue
			}
		}
		// key不存在直接创建map，key存在但非map类型,则覆盖原始值并创建新map
		ntv := make(map[string]any)
		target[k] = ntv
		target = ntv
	}
}

// SetDefaultIfAbsent 设置配置默认值,如果配置不存在或为类型默认值,则设置为默认值
//
//	@param cfg			配置map
//	@param key			配置key，格式a.b.c
//	@param defaultVal	配置默认值
func SetDefaultIfAbsent(cfg map[string]any, key string, defaultVal any) {
	if val, ok := GetByKey(cfg, key); ok {
		if isZero(val) {
			mergeYamlKey(cfg, key, defaultVal)
		}
		return
	}
	mergeYamlKey(cfg, key, defaultVal)
}

// SetOSEnvIfAbsent 设置环境变量人如果缺失的话
func SetOSEnvIfAbsent(key string, defaultVal any) {
	if env := os.Getenv(key); env == "" {
		_ = os.Setenv(key, fmt.Sprint(defaultVal))
	}
}

// isZero
//
//	@Description: 		判断值是否为类型默认值
//	@param val			需要判断的值
//	@return bool		true表示值为类型默认值
func isZero(val any) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64:
		return val == 0
	case float32, float64:
		return val == 0.0
	case string:
		return len(val.(string)) == 0
	default:
		return val == nil
	}
}
