# Display Configuration

## Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_DISPLAY` | bool | Enable support for displays | n |
| `CONFIG_ZMK_DISPLAY_BLANK_ON_IDLE` | bool | Blank display on idle | y if SSD1306 |
| `CONFIG_ZMK_DISPLAY_TICK_PERIOD_MS` | int | Period (in ms) between display task execution | 10 |
| `CONFIG_ZMK_DISPLAY_INVERT` | bool | Invert display colors | n |
| `CONFIG_ZMK_WIDGET_LAYER_STATUS` | bool | Enable a widget to show the highest, active layer | y |
| `CONFIG_ZMK_WIDGET_BATTERY_STATUS` | bool | Enable a widget to show battery charge information | y |
| `CONFIG_ZMK_WIDGET_BATTERY_STATUS_SHOW_PERCENTAGE` | bool | Show percentage instead of icons | n |
| `CONFIG_ZMK_WIDGET_OUTPUT_STATUS` | bool | Enable a widget to show output status | y |
| `CONFIG_ZMK_WIDGET_WPM_STATUS` | bool | Enable a widget to show WPM | y |

## Status Screen

If `CONFIG_ZMK_DISPLAY` is enabled, exactly zero or one of the following options must be set to `y`:

| Config | Description |
| --- | --- |
| `CONFIG_ZMK_DISPLAY_STATUS_SCREEN_BUILT_IN` | Use the built-in status screen |
| `CONFIG_ZMK_DISPLAY_STATUS_SCREEN_CUSTOM` | Use a custom status screen |

## Work Queue

| Config | Description |
| --- | --- |
| `CONFIG_ZMK_DISPLAY_WORK_QUEUE_SYSTEM` | Use the system main thread for UI updates |
| `CONFIG_ZMK_DISPLAY_WORK_QUEUE_DEDICATED` | Use a dedicated thread for UI updates |