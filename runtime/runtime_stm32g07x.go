// +build stm32,stm32g07x

package runtime

import (
	"github.com/tinygo-org/tinygo/src/device/arm"
	"github.com/tinygo-org/tinygo/src/device/stm32"
	"github.com/tinygo-org/tinygo/src/runtime/interrupt"
	"github.com/tinygo-org/tinygo/src/runtime/volatile"
)

const (
	TickMicros     = 1000
	AsyncScheduler = false
)

var (
	timestamp        TimeUnit // microseconds since boottime
	timerLastCounter uint64
	timerWakeup      volatile.Register8
	tickMS           TimeUnit
)

func DeviceInit() {
	initCLK()
	initTIM3()
}

func initCLK() {
	// enable clocks
	stm32.RCC.APBENR2.SetBits(stm32.RCC_APBENR2_SYSCFGEN)
	stm32.RCC.APBENR1.SetBits(stm32.RCC_APBENR1_PWREN)

	// set flash latency
	stm32.FLASH.ACR.ClearBits((0x7 << stm32.Flash_ACR_LATENCY_Pos))
	stm32.FLASH.ACR.SetBits((0x2 << stm32.Flash_ACR_LATENCY_Pos))

	// enable HSI
	stm32.RCC.CR.SetBits(stm32.RCC_CR_HSION)
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_HSIRDY) {
	}

	// configure PLL
	stm32.RCC.PLLSYSCFGR.ClearBits(stm32.RCC_PLLSYSCFGR_PLLSRC_Msk | stm32.RCC_PLLSYSCFGR_PLLM_Msk |
		stm32.RCC_PLLSYSCFGR_PLLN_Msk | stm32.RCC_PLLSYSCFGR_PLLR_Msk)
	stm32.RCC.PLLSYSCFGR.SetBits((0x1 << 1) | 0 | (8 << stm32.RCC_PLLSYSCFGR_PLLN_Pos) | 0x20000000)

	// enable PLL
	stm32.RCC.CR.SetBits(stm32.RCC_CR_PLLON)

	// enable SYSCLK domain
	stm32.RCC.PLLSYSCFGR.SetBits(stm32.RCC_PLLSYSCFGR_PLLREN)

	// wait for PLL ready
	for !stm32.RCC.CR.HasBits(stm32.RCC_CR_PLLRDY) {
	}

	// set AHB prescaler
	stm32.RCC.CFGR.ClearBits(stm32.RCC_CFGR_HPRE_Msk)
	stm32.RCC.CFGR.SetBits(0)

	// Sysclk activation on the main PLL
	stm32.RCC.CFGR.ClearBits(stm32.RCC_CFGR_SW_Msk)
	stm32.RCC.CFGR.SetBits(0x00000002)
	for (stm32.RCC.CFGR.Get() & stm32.RCC_CFGR_SWS_Msk) != 0x00000010 {
	}

	// set APB1 prescaler
	stm32.RCC.CFGR.ClearBits(stm32.RCC_CFGR_PPRE_Msk)
	stm32.RCC.CFGR.SetBits(0)
}

func initTIM3() {
	// enable tim3 clock
	stm32.RCC.APBENR1.SetBits(stm32.RCC_APBENR1_TIM3EN)

	const TIM_CR1_DIR_Msk = 0x00000010
	const TIM_CR1_CMS_Msk = 0x00000060
	tmpcr1 := stm32.TIM3.CR1

	// Set counter mode up
	tmpcr1.ClearBits(TIM_CR1_DIR_Msk | TIM_CR1_CMS_Msk)
	tmpcr1.SetBits(0)

	// Set clock division DIV1
	tmpcr1.ClearBits(stm32.TIM_CR1_CKD_Msk)
	tmpcr1.SetBits(0)

	// update CR1 reg
	stm32.TIM3.CR1.Set(tmpcr1.Get())

	// update_event = TIM_CLK/((PSC + 1)*(ARR + 1)*(RCR + 1))
	// tim_clk = 64000000; psc = 63; arr = 0; rcr = 0
	// update_event = 64000000/(64*1) = 1000000 hz = 1 us
	const sourceClock = 64000000
	stm32.TIM3.ARR.Set(sourceClock/1000 - 1) // set autoreload
	stm32.TIM3.PSC.Set(1 - 1)                // test set prescaler

	// generate an update event to reload the Prescaler
	stm32.TIM3.EGR.SetBits(stm32.TIM_EGR_UG)

	// enable the tim counter
	stm32.TIM3.CR1.SetBits(stm32.TIM_CR1_CEN)

	// enable the update interrupt
	stm32.TIM3.DIER.SetBits(stm32.TIM_DIER_UIE)

	intr := interrupt.New(stm32.IRQ_TIM3, handleTIM3)
	intr.SetPriority(0xc3)
	intr.Enable()

}

func handleTIM3(interrupt.Interrupt) {
	if stm32.TIM3.SR.HasBits(stm32.TIM_SR_UIF) {

		// disable timer (only use it when needed)
		//stm32.TIM3.CR1.ClearBits(stm32.TIM_CR1_CEN)

		// clear update flag
		stm32.TIM3.SR.ClearBits(stm32.TIM_SR_UIF)

		tickMS++

		timerWakeup.Set(1)
	}
}

func timerSleep(us uint64) {
	ms := TimeUnit(us / 1000)
	start := tickMS
	for (tickMS - start) < ms {
		arm.Asm("nop")
	}
}

// SleepTicks should sleep for specific number of microseconds.
func SleepTicks(d TimeUnit) {
	timerSleep(uint64(d))
}

// Number of ticks (microseconds) since start.
func Ticks() TimeUnit {
	return tickMS * TickMicros
}

func Putchar(c byte) {

}
