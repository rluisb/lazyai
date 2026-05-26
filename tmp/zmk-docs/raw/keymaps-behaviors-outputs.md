# Output Selection Behavior

## Summary

The output behavior allows selecting whether keyboard output is sent to the USB or bluetooth connection when both are connected. This allows connecting a keyboard to USB for power but outputting to a different device over bluetooth.

By default, output is sent to USB when both USB and BLE are connected.
Once you select a different output, it will be remembered until you change it again.

## Output Command Defines

Output command defines are provided through the `dt-bindings/zmk/outputs.h` header:

```
#include <dt-bindings/zmk/outputs.h>
```

### Command Defines

| Define | Action |
| --- | --- |
| `OUT_USB` | Prefer sending to USB |
| `OUT_BLE` | Prefer sending to the current bluetooth profile |
| `OUT_TOG` | Toggle between USB and BLE |
| `OUT_NONE` | Prevent from sending any output |

## Output Selection Behavior

The output selection behavior changes the preferred output on press.

### Behavior Binding

Reference: `&out`

Parameter #1: Command, e.g. `OUT_BLE`

### Examples

1. Prefer sending keyboard output to USB:
```
&out OUT_USB
```

2. Prefer sending keyboard output to Bluetooth:
```
&out OUT_BLE
```

3. Toggle between USB and Bluetooth:
```
&out OUT_TOG
```

4. Disable all output:
```
&out OUT_NONE
```

## Powering the Keyboard via USB

ZMK is not always able to detect if the other end of a USB connection accepts keyboard input or not.
So if you are using USB only to power your keyboard (for example with a charger or a portable power bank), you will want to select the BLE output through below behavior to be able to send keystrokes to the selected bluetooth profile.

## Output Selection Persistence

The endpoint that is selected by the `&out` behavior will be saved to flash storage and hence persist across restarts and firmware flashes.
However it will only be saved after `CONFIG_ZMK_SETTINGS_SAVE_DEBOUNCE` milliseconds in order to reduce potential wear on the flash memory.