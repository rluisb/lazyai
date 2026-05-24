# Power Management Configuration

## External Power Control

### Kconfig

Definition file: zmk/app/Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_EXT_POWER` | bool | Enable support to control external power output | y |

### Devicetree

Applies to: `compatible = "zmk,ext-power-generic"`

| Property | Type | Description |
| --- | --- | --- |
| `control-gpios` | GPIO array | List of GPIOs which should be active to enable external power |
| `init-delay-ms` | int | number of milliseconds to delay after initializing the driver |

## GPIO Key Wakeup Trigger

Applies to: `compatible = "zmk,gpio-key-wakeup-trigger"`

| Property | Type | Description |
| --- | --- | --- |
| `trigger` | phandle | Phandle to a GPIO key to be used to wake from soft off |
| `wakeup-source` | bool | Mark this device as able to wake the keyboard |
| `extra-gpios` | GPIO array | list of GPIO pins to set active before going into power off |

## Idle/Sleep Settings

### Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_IDLE_TIMEOUT` | int | Milliseconds of inactivity before entering idle state | 30000 |
| `CONFIG_ZMK_SLEEP` | bool | Enable deep sleep support | n |
| `CONFIG_ZMK_IDLE_SLEEP_TIMEOUT` | int | Milliseconds of inactivity before entering deep sleep | 900000 |
| `CONFIG_ZMK_PM_SOFT_OFF` | bool | Enable soft off functionality from the keymap or dedicated hardware | n |