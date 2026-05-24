# New Behavior

Key files and APIs for creating new behaviors:

- `zephyr/device.h`: Zephyr Device APIs
- `drivers/behavior.h`: ZMK Behavior Functions (e.g. locality, `behavior_keymap_binding_pressed`, `behavior_keymap_binding_released`, `behavior_sensor_keymap_binding_triggered`)
- `zephyr/logging/log.h`: Zephyr Logging APIs
- `zmk/behavior.h`: ZMK Behavior Information (e.g. parameters, position and timestamp of events)

### Return values:

- `ZMK_BEHAVIOR_OPAQUE`: Used to terminate `on__binding_pressed` and `on__binding_released` functions that accept `(struct zmk_behavior_binding binding, struct zmk_behavior_binding_event event)` as parameters
- `ZMK_BEHAVIOR_TRANSPARENT`: Used in the `binding_pressed` and `binding_released` functions for the transparent (`&trans`) behavior