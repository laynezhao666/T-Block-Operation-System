package controller

// GPIOControllerTbox is a specific implementation of GPIOBaseController
type GPIOControllerTbox struct {
	GPIOBaseController
}

// NewGPIOControllerTbox creates a new GPIOControllerTbox
func NewGPIOControllerTbox(rootPath string) *GPIOControllerTbox {
	controller := &GPIOControllerTbox{
		GPIOBaseController: *NewGPIOBaseController(rootPath), // Assuming NewGPIOBaseController is defined
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
		"DO1":  "do1",
		"DO2":  "do2",
		"DO3":  "do3",
		"DO4":  "do4",
	}
	return controller
}
