// Package controller is a package for controllers
package controller

import (
	"fmt"
	"os"
	"strconv"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	IN            = 0
	OUT           = 1
	VALUE_STR_MAX = 10
)

// GPIOBaseController GPIO控制器
type GPIOBaseController struct {
	rootPath string
	//readLogger  *Logger
	//writeLogger *Logger
	pinNameMap map[string]string
}

// NewGPIOBaseController 创建GPIO控制器
func NewGPIOBaseController(rootPath string) *GPIOBaseController {
	return &GPIOBaseController{
		rootPath: rootPath,
		//readLogger:  NewLogger(60),
		//writeLogger: NewLogger(60),
		pinNameMap: make(map[string]string),
	}
}

// IsExport 判断是否已导出
func (c *GPIOBaseController) IsExport(pin string) bool {
	return true
}

// Export 导出
func (c *GPIOBaseController) Export(pin string) int {
	return -1
}

// Unexport 取消导出
func (c *GPIOBaseController) Unexport(pin string) int {
	return -1
}

// Direction 设置GPIO方向
func (c *GPIOBaseController) Direction(pin string, dir int) int {
	dirStr := []string{"in", "out"}
	path := fmt.Sprintf("%s/%s/direction", c.rootPath, c.pinNameMap[pin])

	fd, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return -1
	}
	defer fd.Close()

	_, err = fd.WriteString(dirStr[dir])
	if err != nil {
		return -1
	}

	return 0
}

// Write 写入GPIO
func (c *GPIOBaseController) Write(pin string, value int) int {
	valuesStr := []string{"0", "1"}
	path := fmt.Sprintf("%s/gpio/%s/value", c.rootPath, c.pinNameMap[pin])

	//flag := c.writeLogger.NotExist(pin)

	fd, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		//if flag {
		log.Errorf("GPIO open error: %v", err)
		//}
		return -1
	}
	defer fd.Close()

	if value > 1 || value < 0 {
		log.Errorf("GPIO value not in (0,1):%v", value)
		return -1
	}

	_, err = fd.WriteString(valuesStr[value])
	if err != nil {
		//if flag {
		log.Errorf("GPIO write error: %v", err)
		//}
		return -1
	}

	return 0
}

// Read 读取GPIO
func (c *GPIOBaseController) Read(pin string) int {
	path := fmt.Sprintf("%s/gpio/%s/value", c.rootPath, c.pinNameMap[pin])

	fd, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return -1
	}
	defer fd.Close()

	valueStr := make([]byte, VALUE_STR_MAX) // Assume VALUE_STR_MAX is defined
	n, err := fd.Read(valueStr)
	if err != nil || n <= 0 || n > VALUE_STR_MAX {
		//if c.readLogger.NotExist(pin) { // Assume NotExist is defined in Logger
		log.Errorf("GPIO read error: %v", err)
		//}
		return -1
	}

	// Parse the value
	result := -1
	begin := 0
	for begin < n && !isDigit(valueStr[begin]) {
		begin++
	}
	end := begin + 1
	for end < n && isDigit(valueStr[end]) {
		end++
	}
	if end < n {
		valueStr[end] = 0
	}

	if parsedValue, err := strconv.Atoi(string(valueStr[begin:end])); err == nil {
		if parsedValue != 0 {
			result = 1
		} else {
			result = 0
		}
	} else {
		result = -1
	}

	return result
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
