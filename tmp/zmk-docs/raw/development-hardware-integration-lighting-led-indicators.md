# LED Indicators

### PWM Brightness Control

The above example only supports LEDs being off or on at full brightness. If you want to be able to reduce the brightness or use multiple brightness levels, you must use `pwm-leds` instead of `gpio-leds`. Note that this will increase power usage slightly when LEDs are enabled compared to using `gpio-leds`.

See the backlight hardware integration page for an example of configuring PWM LEDs.

## Indicator Definitions

Now that you have some LEDs defined, you can configure ZMK to use them.

In your shield's `.overlay` file or board's `.dts` file, add the following include to the top of the file:

```
#include <dt-bindings/zmk/hid_indicators.h>
```

Then, add a `zmk,indicator-leds` node. This node can have any number of child nodes. Each child maps an indicator state to one or more LEDs:

In your shield's `.overlay` file or board's `.dts` file, create a `gpio-leds` node and define any LEDs that you want to be able to control. The following example defines two LEDs on `&gpio0 1` and `&gpio0 2` which are enabled by driving the pins high: