// +build stm32, stm32g07x

package gpio

import (
	"github.com/tinygo-org/tinygo/src/device/stm32"
)

const (
	// Mode Flag
	PinInput    PinMode = 0x00
	PinOutputPP PinMode = 0x01
	PinOutputOD PinMode = 0x11
	PinAltPP    PinMode = 0x02
	PinAltOD    PinMode = 0x12
	PinAnalog   PinMode = 0x03

	// Pull Flag
	PinNoPull   PinPull = 0x00
	PinPullUp   PinPull = 0x01
	PinPullDown PinPull = 0x02

	// GPIOx_OSPEEDR
	GPIO_SPEED_LOW     = 0
	GPIO_SPEED_MID     = 1
	GPIO_SPEED_HI      = 2
	GPIO_SPEED_VERY_HI = 3

	// GPIO Mask
	GPIO_OUTPUT_TYPE = 0x00000010
	GPIO_MODE        = 0x00000003
)

const (
	PA0  = portA + 0
	PA1  = portA + 1
	PA2  = portA + 2
	PA3  = portA + 3
	PA4  = portA + 4
	PA5  = portA + 5
	PA6  = portA + 6
	PA7  = portA + 7
	PA8  = portA + 8
	PA9  = portA + 9
	PA10 = portA + 10
	PA11 = portA + 11
	PA12 = portA + 12
	PA13 = portA + 13
	PA14 = portA + 14
	PA15 = portA + 15

	PB0  = portB + 0
	PB1  = portB + 1
	PB2  = portB + 2
	PB3  = portB + 3
	PB4  = portB + 4
	PB5  = portB + 5
	PB6  = portB + 6
	PB7  = portB + 7
	PB8  = portB + 8
	PB9  = portB + 9
	PB10 = portB + 10
	PB11 = portB + 11
	PB12 = portB + 12
	PB13 = portB + 13
	PB14 = portB + 14
	PB15 = portB + 15

	PC0  = portC + 0
	PC1  = portC + 1
	PC2  = portC + 2
	PC3  = portC + 3
	PC4  = portC + 4
	PC5  = portC + 5
	PC6  = portC + 6
	PC7  = portC + 7
	PC8  = portC + 8
	PC9  = portC + 9
	PC10 = portC + 10
	PC11 = portC + 11
	PC12 = portC + 12
	PC13 = portC + 13
	PC14 = portC + 14
	PC15 = portC + 15
)

func (p Pin) getPort() *stm32.GPIO_Type {
	switch p / 16 {
	case 0:
		return stm32.GPIOA
	case 1:
		return stm32.GPIOB
	case 2:
		return stm32.GPIOC
	case 3:
		return stm32.GPIOD
	case 5:
		return stm32.GPIOF
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPAEN)
	case 1:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPBEN)
	case 2:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPCEN)
	case 3:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPDEN)
	case 5:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPFEN)
	default:
		panic("machine: unknown port")
	}
}

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
	// Configure the GPIO pin.
	p.enableClock()
	port := p.getPort()
	pin := uint8(p) % 16
	gpioPin := uint32(0x1 << pin)

	port.MODER.ClearBits((gpioPin * gpioPin) * stm32.GPIO_MODER_MODER0_Msk)
	port.MODER.SetBits((gpioPin * gpioPin) * uint32(config.Mode&GPIO_MODE))

	if config.Mode != PinInput && config.Mode != PinAnalog {
		port.OSPEEDR.ClearBits((gpioPin * gpioPin) * stm32.GPIO_OSPEEDR_OSPEEDR0_Msk)
		port.OSPEEDR.SetBits((gpioPin * gpioPin) * GPIO_SPEED_HI)

		port.OTYPER.ClearBits(gpioPin)
		port.OTYPER.SetBits(gpioPin * uint32((config.Mode&GPIO_OUTPUT_TYPE)>>4))
	}

	port.PUPDR.ClearBits((gpioPin * gpioPin) * stm32.GPIO_PUPDR_PUPDR0_Msk)
	port.PUPDR.SetBits((gpioPin * gpioPin) * uint32(config.Pull))

}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	port := p.getPort()
	pin := uint8(p) % 16
	if high {
		port.BSRR.Set(1 << pin)
	} else {
		port.BRR.Set(1 << pin)
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	port := p.getPort()
	pin := uint8(p) % 16
	val := port.IDR.Get() & (1 << pin)
	return (val > 0)
}
