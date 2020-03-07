// +build stm32

package gpio

type (
	Pin       int8
	PinMode   uint8
	PinPull   uint8
	PinConfig struct {
		Mode PinMode
		Pull PinPull
	}
)

const NoPin = Pin(-1)

const (
	portA Pin = iota * 16
	portB
	portC
	portD
	portE
	portF
	portG
	portH
)

func (p Pin) High() {
	p.Set(true)
}

func (p Pin) Low() {
	p.Set(false)
}
