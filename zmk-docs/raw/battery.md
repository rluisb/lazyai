# Battery Level

See the battery level feature page for more details on configuring a battery sensor.

## Kconfig

Definition file: zmk/app/Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_BATTERY_REPORTING` | bool | Enables/disables all battery level detection/reporting | n |
| `CONFIG_ZMK_BATTERY_REPORT_INTERVAL` | int | Battery level report interval in seconds | 60 |

## Battery Voltage Divider Sensor

Applies to: `compatible = "zmk,battery-voltage-divider"`

See Zephyr's voltage divider documentation.

## nRF VDDH Battery Sensor

Applies to: `compatible = "zmk,battery-nrf-vddh"`

Definition file: zmk/app/module/dts/bindings/sensor/zmk,battery-nrf-vddh.yaml

This driver has no configuration.