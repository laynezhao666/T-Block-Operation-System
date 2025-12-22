package controller

import (
	"strings"
)

// Define protocol constants
const (
	PROTOCOL_SYS_CLASS = "sys_class"
	PROTOCOL_UAL       = "ual"
	PROTOCOL_HYIOT     = "hyiot"
	PROTOCOL_TBOX      = "tbox"
)

// GPIOControllerUal, GPIOControllerSysclass, GPIOControllerHYIOT, and GPIOControllerTbox
var (
	gpioControllerUal      = NewGPIOControllerUal("/ual")
	gpioControllerSysclass = NewGPIOControllerSysclass("/sys/class")
	gpioControllerHYIOT    = NewGPIOControllerHYIOT("/sys/class/tbox")
	gpioControllerTbox     = NewGPIOControllerTbox("/usr/dev")
)

// IGPIOController is an interface that all GPIO controllers should implement
type IGPIOController interface {
	Export(pin string) int
	Unexport(pin string) int
	Direction(pin string, dir int) int
	Write(pin string, value int) int
	Read(pin string) int
}

// GetController returns the appropriate GPIO controller based on the protocol version
func GetController(protocolVer string) IGPIOController {
	temp := strings.ToLower(protocolVer)
	switch temp {
	case PROTOCOL_SYS_CLASS:
		return gpioControllerSysclass
	case PROTOCOL_UAL:
		return gpioControllerUal
	case PROTOCOL_HYIOT:
		return gpioControllerHYIOT
	default:
		return gpioControllerTbox
	}
}
