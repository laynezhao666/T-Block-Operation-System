// Package controller is a package for controllers
package controller

import (
	"fmt"
	"os"
)

// GPIOControllerSysclass is a specific implementation of IGPIOController for sysfs GPIO
type GPIOControllerSysclass struct {
	GPIOBaseController
	pinNameMap map[string]string
}

// NewGPIOControllerSysclass creates a new instance of GPIOControllerSysclass
func NewGPIOControllerSysclass(rootPath string) *GPIOControllerSysclass {
	controller := &GPIOControllerSysclass{
		GPIOBaseController: *NewGPIOBaseController(rootPath),
		pinNameMap: map[string]string{
			"DO1":  "gpio253",
			"DO2":  "gpio252",
			"DO3":  "gpio255",
			"DO4":  "gpio254",
			"DI1":  "gpio241",
			"DI2":  "gpio240",
			"DI3":  "gpio243",
			"DI4":  "gpio242",
			"DI5":  "gpio245",
			"DI6":  "gpio244",
			"DI7":  "gpio247",
			"DI8":  "gpio246",
			"DI9":  "gpio249",
			"DI10": "gpio248",
			"DI11": "gpio251",
			"DI12": "gpio250",
		},
	}
	return controller
}

// Export exports a GPIO pin
func (c *GPIOControllerSysclass) Export(pin string) int {
	pinName, exists := c.pinNameMap[pin]
	if !exists {
		return -1 // Pin not found
	}

	fd, err := os.OpenFile("/sys/class/gpio/export", os.O_WRONLY, 0644)
	if err != nil {
		return -1 // Failed to open export file
	}
	defer fd.Close()

	if _, err := fd.WriteString(pinName); err != nil {
		return -1 // Failed to write to export file
	}

	return 0
}

// Unexport unexports a GPIO pin
func (c *GPIOControllerSysclass) Unexport(pin string) int {
	pinName, exists := c.pinNameMap[pin]
	if !exists {
		return -1 // Pin not found
	}

	fd, err := os.OpenFile("/sys/class/gpio/unexport", os.O_WRONLY, 0644)
	if err != nil {
		return -1 // Failed to open unexport file
	}
	defer fd.Close()

	if _, err := fd.WriteString(pinName); err != nil {
		return -1 // Failed to write to unexport file
	}

	return 0
}

// IsExport checks if a GPIO pin is exported
func (c *GPIOControllerSysclass) IsExport(pin string) bool {
	pinName, exists := c.pinNameMap[pin]
	if !exists {
		return false // Pin not found
	}

	path := fmt.Sprintf("/sys/class/gpio/%s", pinName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false // Path does not exist
	}

	return true // Path exists
}
