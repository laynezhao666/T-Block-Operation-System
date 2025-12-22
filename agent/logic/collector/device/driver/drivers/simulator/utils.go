package simulator

import (
	"encoding/json"
	"strconv"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/logic/collector/device/driver/drivers/simulator/generator"
)

type generatorData struct {
	Name  string `json:"name"`
	Min   string `json:"min"`
	Max   string `json:"max"`
	Type  string `json:"type"`
	Step  string `json:"step"`
	Value string `json:"value"`
}

// NewDefaultGenerator 创建默认的生成器
func NewDefaultGenerator() *generator.Generator {
	return generator.New(generator.NewRandomImpl(-99999.0, -99990.0))
}

// CreateStaticGenerator 创建静态生成器
func CreateStaticGenerator(d *generatorData) *generator.Generator {
	x, err := strconv.ParseFloat(d.Value, 64)
	if err != nil {
		return nil
	}
	return generator.New(generator.NewStaticImpl(x))
}

// CreateMonotoneGenerator 创建单调递增生成器
func CreateMonotoneGenerator(d *generatorData) *generator.Generator {
	var (
		err  error
		min  float64
		max  float64
		step float64
	)

	if err = json.Unmarshal([]byte(d.Min), &min); err != nil {
		return nil
	}
	if err = json.Unmarshal([]byte(d.Max), &max); err != nil {
		return nil
	}
	if err = json.Unmarshal([]byte(d.Step), &step); err != nil {
		return nil
	}

	return generator.New(generator.NewMonotoneImpl(min, max, step))
}

// CreateRandomGenerator 创建随机生成器
func CreateRandomGenerator(d *generatorData) *generator.Generator {
	var (
		err error
		min float64
		max float64
	)

	if err = json.Unmarshal([]byte(d.Min), &min); err != nil {
		return nil
	}
	if err = json.Unmarshal([]byte(d.Max), &max); err != nil {
		return nil
	}

	return generator.New(generator.NewRandomImpl(min, max))
}

// CreateGenerator 创建模拟器生成器
func CreateGenerator(s string) *generator.Generator {
	var (
		err  error
		data generatorData
		g    *generator.Generator
	)

	for {
		if err = json.Unmarshal([]byte(s), &data); err != nil {
			log.Warnf("unmarshal simulator rule \"%v\" error: %v", s, err)
			break
		}

		switch data.Name {
		case "random":
			g = CreateRandomGenerator(&data)
		case "monotone":
			g = CreateMonotoneGenerator(&data)
		case "static":
			g = CreateStaticGenerator(&data)
		}

		if g == nil {
			break
		}

		return g
	}

	return NewDefaultGenerator()
}
