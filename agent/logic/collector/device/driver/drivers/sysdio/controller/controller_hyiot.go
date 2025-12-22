package controller

// GPIOControllerHYIOT is a specific implementation of IGPIOController for HYIOT protocol
type GPIOControllerHYIOT struct {
	GPIOBaseController
	pinNameMap map[string]string
	doPinMap   map[string]int
}

// NewGPIOControllerHYIOT creates a new instance of GPIOControllerHYIOT
func NewGPIOControllerHYIOT(rootPath string) *GPIOControllerHYIOT {
	controller := &GPIOControllerHYIOT{
		GPIOBaseController: *NewGPIOBaseController(rootPath),
		pinNameMap: map[string]string{
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
		},
		doPinMap: map[string]int{
			"do1": 1,
			"do2": 2,
			"do3": 3,
			"do4": 4,
		},
	}
	return controller
}

// IsExport checks if a pin is exported
func (c *GPIOControllerHYIOT) IsExport(pin string) bool {
	return c.GPIOBaseController.IsExport(pin)
}

// Export exports a pin
func (c *GPIOControllerHYIOT) Export(pin string) int {
	return 0
}

// Unexport unexports a pin
func (c *GPIOControllerHYIOT) Unexport(pin string) int {
	return 0
}

// Write writes a value to a GPIO pin
func (c *GPIOControllerHYIOT) Write(pin string, value int) int {
	pinName, exists := c.pinNameMap[pin]
	if !exists {
		return -1 // Pin not found
	}

	pinChannel, exists := c.doPinMap[pinName]
	if !exists {
		return -1 // Pin not found in doPinMap
	}

	kValuesStr := "01"
	if value < 0 || value >= len(kValuesStr) {
		return -1 // Invalid value
	}

	// Call the control function
	if err := ctrlDo(pinChannel, string(kValuesStr[value])); err != nil {
		return -1 // Error in control
	}

	return 0
}

// ctrlDo is a placeholder for the actual control function
func ctrlDo(pinChannel int, value string) error {
	// dummy implement
	return nil
}
