# Board Pin Control

Many boards will already have some pins configured for particular protocols. We recommend using these pins whenever possible.

The following node labels are available:

- I2C bus: `&arduino_i2c`
- SPI bus: `&arduino_spi`
- UART: `&arduino_serial`
- ADC: `&arduino_adc`

Predefined Nodes for various platforms including Arduino Uno Rev3 Shields, BlackPill Shields, Pro Micro Shields, and Seeed XIAO Shields.

## Using Pinctrl

In the main file:

```
#include "-pinctrl.dtsi"

&uart0 {
  pinctrl-0 = <&uart0_default>;
  pinctrl-1 = <&uart0_sleep>;
  pinctrl-names = "default", "sleep";
};
```

On designs using wired split on nRF52840, using asynchronous UART APIs with DMA will help ensure that the interrupts used to handle timing sensitive BT interactions can respond when needed.

This can be accomplished by overwriding the `compatible` property to `"nordic,nrf-uarte"`.