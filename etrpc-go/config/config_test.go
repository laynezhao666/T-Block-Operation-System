package config

import (
	"etrpc-go/config/cache"
	"etrpc-go/config/loader"
	"etrpc-go/config/loader/file"
	"etrpc-go/config/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// TestLoadConfig 测试本地配置及七彩石配置是否加载正常
func TestLoadConfig(t *testing.T) {
	file.LocalConfigPath = "./trpc_go.yaml"
	loader.LoadConfig()
	assert.NotEmpty(t, cache.GetCfgMap())
	strVal, ok := util.GetByKey(cache.GetCfgMap(), "etrpc.string.val")
	assert.Equal(t, true, ok)
	assert.Equal(t, "aasdasd", strVal)
}

// TestRegisterConfig 测试注册配置是否正常读写
func TestRegisterConfig(t *testing.T) {
	file.LocalConfigPath = "./trpc_go.yaml"
	cfg := &struct {
		Etrpc struct {
			ServiceName string `yaml:"service_name"`
			ServicePort int    `yaml:"service_port"`
		} `yaml:"etrpc"`
	}{}

	RegisterConfig("test", cfg, false)
	loader.LoadConfig()
	assert.Equal(t, "idc-public-demo", cfg.Etrpc.ServiceName)
	assert.Equal(t, 8080, cfg.Etrpc.ServicePort)
}

func TestRegisterConfigDynamicLoad(t *testing.T) {
	file.LocalConfigPath = "./trpc_go.yaml"
	cfg := &struct {
		DynamicKey string `yaml:"dynamicKey"`
	}{}
	dynamicCfg := &struct {
		DynamicKey string `yaml:"dynamicKey"`
	}{}
	RegisterConfig("test", cfg, false)
	RegisterConfig("test-dynamic", dynamicCfg, true)
	loader.LoadConfig()
	assert.Equal(t, "222", cfg.DynamicKey)
	assert.Equal(t, "222", dynamicCfg.DynamicKey)
	time.Sleep(time.Second * 10)
	assert.Equal(t, "222", cfg.DynamicKey)
	//assert.Equal(t, "333", dynamicCfg.DynamicKey)
}

func TestGet(t *testing.T) {
	file.LocalConfigPath = "./trpc_go.yaml"
	loader.LoadConfig()
	etrpc, ok := Get("etrpc")
	assert.True(t, ok)
	assert.Equal(t, "idc-public-demo", etrpc.(map[string]any)["service_name"])
	assert.Equal(t, 8080, etrpc.(map[string]any)["service_port"])

	// int32 key exist
	int32Val, ok := GetInt32("test.int32")
	assert.Equal(t, int32(1111111), int32Val)
	// int32-str key exist
	int32Val, ok = GetInt32("test.int32-str")
	assert.Equal(t, int32(1111111), int32Val)
	// int32 key exits bad type
	int32Val, ok = GetInt32("test")
	assert.Equal(t, false, ok)
	// int32 key exits bad type, but default
	int32Val = GetInt32OrDefault("test", 111)
	assert.Equal(t, int32(111), int32Val)
	// int32 key not exits
	int32Val, ok = GetInt32("test.int32-1")
	assert.Equal(t, false, ok)
	// int32 key not exits, but default
	int32Val = GetInt32OrDefault("test.int32-1", 111)
	assert.Equal(t, int32(111), int32Val)

	// int64 key exist
	int64Val, ok := GetInt64("test.int64")
	assert.Equal(t, int64(1111111111111111111), int64Val)
	// int64-str key exist
	int64Val, ok = GetInt64("test.int64-str")
	assert.Equal(t, int64(1111111111111111111), int64Val)
	// int64 key exits with compatible type
	int64Val, ok = GetInt64("test.int32")
	assert.Equal(t, int64(1111111), int64Val)
	// int64 key exits bad type
	int64Val, ok = GetInt64("test")
	assert.Equal(t, false, ok)
	// int64 key exits bad type, but default
	int64Val = GetInt64OrDefault("test", 1111111111111111111)
	assert.Equal(t, int64(1111111111111111111), int64Val)
	// int64 key not exits
	int64Val, ok = GetInt64("test.int64-1")
	assert.Equal(t, false, ok)
	// int64 key not exits, but default
	int64Val = GetInt64OrDefault("test.int64-1", 1111111111111111111)
	assert.Equal(t, int64(1111111111111111111), int64Val)

	// float32 key exist
	float32Val, ok := GetFloat32("test.float32")
	assert.Equal(t, float32(0.1111), float32Val)
	// float32-str key exist
	float32Val, ok = GetFloat32("test.float32-str")
	assert.Equal(t, float32(0.1111), float32Val)
	// float32 key exits bad type
	float32Val, ok = GetFloat32("test")
	assert.Equal(t, false, ok)
	// float32 key exits bad type, but default
	float32Val = GetFloat32OrDefault("test", 0.1111)
	assert.Equal(t, float32(0.1111), float32Val)
	// float32 key not exits
	float32Val, ok = GetFloat32("test.float32-1")
	assert.Equal(t, false, ok)
	// float32 key not exits, but default
	float32Val = GetFloat32OrDefault("test.float32-1", 0.1111)
	assert.Equal(t, float32(0.1111), float32Val)

	// float64 key exist
	float64Val, ok := GetFloat64("test.float64")
	assert.Equal(t, 1111111.111111111111111111111111111111111111111111111111111111111111111111111, float64Val)
	// float64-str key exist
	float64Val, ok = GetFloat64("test.float64-str")
	assert.Equal(t, 1111111.111111111111111111111111111111111111111111111111111111111111111111111, float64Val)
	// float64 key exits with compatible type
	float64Val, ok = GetFloat64("test.float32")
	assert.Equal(t, 0.1111, float64Val)
	// float64 key exits bad type
	float64Val, ok = GetFloat64("test")
	assert.Equal(t, false, ok)
	// float64 key exits bad type, but default
	float64Val = GetFloat64OrDefault("test", 1111111.111111111111111111111111111111111111111111111111111111111111111111111)
	assert.Equal(t, 1111111.111111111111111111111111111111111111111111111111111111111111111111111, float64Val)
	// float64 key not exits
	float64Val, ok = GetFloat64("test.float64-1")
	assert.Equal(t, false, ok)
	// float64 key not exits, but default
	float64Val = GetFloat64OrDefault("test.float64-1", 1111111.111111111111111111111111111111111111111111111111111111111111111111111)
	assert.Equal(t, 1111111.111111111111111111111111111111111111111111111111111111111111111111111, float64Val)

	// string key exist
	strVal, ok := GetString("test.string")
	assert.Equal(t, "string", strVal)
	// string key exist with compatible type
	strVal, ok = GetString("test.int64")
	assert.Equal(t, "1111111111111111111", strVal)
	// string key not exist
	strVal, ok = GetString("test.string-1")
	assert.Equal(t, false, ok)
	// string key exist with empty
	strVal, ok = GetString("test.string-empty")
	assert.Equal(t, "", strVal)
	// string key exist with default
	strVal = GetStringOrDefault("test.string", "str")
	assert.Equal(t, "string", strVal)
	// string key not exist with default
	strVal = GetStringOrDefault("test.string-1", "str")
	assert.Equal(t, "str", strVal)

	// bool key exist
	boolVal, ok := GetBool("test.bool")
	assert.Equal(t, true, boolVal)
	// bool-str key exist
	boolVal, ok = GetBool("test.bool-str")
	assert.Equal(t, true, boolVal)
	// bool key not exist
	boolVal, ok = GetBool("test.bool-1")
	assert.Equal(t, false, ok)
	// bool key exist with bad type
	boolVal, ok = GetBool("test")
	assert.Equal(t, false, ok)
	// bool key not exist, but default
	boolVal = GetBoolOrDefault("test.bool-1", true)
	assert.Equal(t, true, boolVal)
	// bool key exist with bad type, but default
	boolVal = GetBoolOrDefault("test", true)
	assert.Equal(t, true, boolVal)

	// date key exist
	timeVal, ok := GetTime("test.date")
	assert.Equal(t, "2024-05-20", timeVal.Format("2006-01-02"))
	// date-str key exist
	timeVal, ok = GetTime("test.date-str")
	assert.Equal(t, "2024-05-20", timeVal.Format("2006-01-02"))
	// date key not exist
	timeVal, ok = GetTime("test.date-1")
	assert.Equal(t, false, ok)
	// date key exist with bad type
	timeVal, ok = GetTime("test")
	assert.Equal(t, false, ok)
	// date key not exist, but default
	timeVal = GetTimeOrDefault("test.date-1", time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC), timeVal)
	// date key exist with bad type, but default
	timeVal = GetTimeOrDefault("test", time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC))
	assert.Equal(t, time.Date(2024, 5, 20, 0, 0, 0, 0, time.UTC), timeVal)

	// time key exist
	timeVal, ok = GetTime("test.time")
	assert.Equal(t, "2024-05-20 00:00:00", timeVal.Format("2006-01-02 15:04:05"))
	// date-str key exist
	timeVal, ok = GetTime("test.time-str")
	assert.Equal(t, "2024-05-20 00:00:00", timeVal.Format("2006-01-02 15:04:05"))

}
