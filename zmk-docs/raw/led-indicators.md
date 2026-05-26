# LED Indicators Configuration

## Kconfig

Definition files: zmk/app/src/indicators/Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_INDICATOR_LEDS_INIT_PRIORITY` | int | Indicator LED device driver initialization priority | 91 |

`CONFIG_ZMK_INDICATOR_LEDS_INIT_PRIORITY` must be set to a larger value than `CONFIG_LED_INIT_PRIORITY`.

## Indicator LED Driver

Applies to: `compatible = "zmk,indicator-leds"`

| Property | Type | Description | Default |
| --- | --- | --- | --- |
| `indicator` | int | The `HID_INDICATOR_*` value to indicate |  |
| `leds` | phandles | One or more LED devices to control |  |
| `active-brightness` | int | LED brightness in percent when the indicator is active | 100 |
| `inactive-brightness` | int | LED brightness in percent when the indicator is not active | 0 |
| `disconnected-brightness` | int | LED brightness in percent when the keyboard is not connected | 0 |
| `on-while-idle` | bool | Keep LEDs enabled even when the keyboard is idle and on battery power | false |