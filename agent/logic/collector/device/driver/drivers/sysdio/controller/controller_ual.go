package controller

import (
	"fmt"
	"os"

	"trpc.group/trpc-go/trpc-go/log"
)

// GPIOControllerUal is a specific implementation of GPIOBaseController
type GPIOControllerUal struct {
	GPIOBaseController
}

// NewGPIOControllerUal creates a new GPIOControllerUal
func NewGPIOControllerUal(rootPath string) *GPIOControllerUal {
	controller := &GPIOControllerUal{
		GPIOBaseController: *NewGPIOBaseController(rootPath),
	}
	controller.pinNameMap = map[string]string{
		"DI1":  "di1",
		"DI2":  "di2",
		"DI3":  "di3",
		"DI4":  "di4",
		"DI5":  "di5",
		"DI6":  "di6",
		"DI7":  "di7",
		"DI8":  "di8",
		"DI9":  "di9",
		"DI10": "di10",
		"DI11": "di11",
		"DI12": "di12",
		"DO1":  "do1",
		"DO2":  "do2",
		"DO3":  "do3",
		"DO4":  "do4",
	}
	return controller
}

// Export exports a GPIO pin
func (c *GPIOControllerUal) Export(pin string) int {
	path := fmt.Sprintf("%s/gpio/export", c.rootPath)
	fd, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		log.Errorf("export error: %v", err)
		return -1
	}
	defer fd.Close()

	_, err = fd.WriteString(c.pinNameMap[pin] + "\n")
	if err != nil {
		log.Errorf("write export error: %v", err)
		return -1
	}

	return 0
}

// Unexport unexports a GPIO pin
func (c *GPIOControllerUal) Unexport(pin string) int {
	path := fmt.Sprintf("%s/gpio/unexport", c.rootPath)
	fd, err := os.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		log.Errorf("unexport error: %v", err)
		return -1
	}
	defer fd.Close()

	_, err = fd.WriteString(c.pinNameMap[pin] + "\n")
	if err != nil {
		log.Errorf("write unexport error: %v", err)
		return -1
	}

	return 0
}
