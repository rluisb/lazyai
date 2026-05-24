# Power Management Behaviors

## Summary

ZMK provides external power control behaviors to manage power to connected devices or peripherals.

## External Power Control Command Defines

External power control command defines are provided through the `dt-bindings/zmk/ext_power.h` header:

```
#include <dt-bindings/zmk/ext_power.h>
```

### Command Defines

| Define | Action | Alias |
| --- | --- | --- |
| `EXT_POWER_OFF_CMD` | Disable the external power. | `EP_OFF` |
| `EXT_POWER_ON_CMD` | Enable the external power. | `EP_ON` |
| `EXT_POWER_TOGGLE_CMD` | Toggle the external power. | `EP_TOG` |

## Behavior Binding

Reference: `&ext_power`

Parameter #1: Command, e.g `EP_ON`

### Examples

1. Enable the external power:
```
&ext_power EP_ON
```

2. Disable the external power:
```
&ext_power EP_OFF
```

3. Toggle the external power:
```
&ext_power EP_TOG
```

## External Power State Persistence

The on/off state that is set by the `&ext_power` behavior will be saved to flash storage and hence persist across restarts and firmware flashes.
However it will only be saved after `CONFIG_ZMK_SETTINGS_SAVE_DEBOUNCE` milliseconds in order to reduce potential wear on the flash memory.

## Split Keyboards

Power management behaviors are global: This means that when triggered, they affects both the central and peripheral side of split keyboards.