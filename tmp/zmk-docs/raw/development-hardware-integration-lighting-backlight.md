# Backlight

In this example, `PWM_MSEC(10)` is the period of the PWM waveform. This period can be altered if your drive circuitry requires different values or the frequency is audible.

If your board uses a P-channel MOSFET to control backlight instead of a N-channel MOSFET, you may want to change `PWM_POLARITY_NORMAL` for `PWM_POLARITY_INVERTED`.

The value inside `pwm_led_0` after `&pwm0` must be the channel number. Since `PWM_OUT0` is defined in the pinctrl node, the channel in this example is 0.

Finally you need to add backlight to the `chosen` element of the root devicetree node.

## Multiple Backlight LEDs

It is possible to control multiple backlight LEDs at the same time. This is useful if, for example, you have a Caps Lock LED connected to a different pin and you want it to be part of the backlight.

In order to do that, first configure PWM for each pin in the pinctrl node.

Pin numbers are handled differently depending on the MCU. On nRF MCUs pins are configured using `(PWM_OUTX, Y, Z)`, where `X` is the PWM channel used (usually 0), `Y` is the first part of the hardware port (PY.01) and `Z` is the second part of the hardware port (P1.Z).