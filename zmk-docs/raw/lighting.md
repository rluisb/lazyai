# Lighting Configuration

## RGB Underglow

### Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_RGB_UNDERGLOW` | bool | Enable RGB underglow | n |
| `CONFIG_ZMK_RGB_UNDERGLOW_EXT_POWER` | bool | Underglow toggling also controls external power | y |
| `CONFIG_ZMK_RGB_UNDERGLOW_AUTO_OFF_IDLE` | bool | Turn off RGB underglow when keyboard goes into idle state | n |
| `CONFIG_ZMK_RGB_UNDERGLOW_AUTO_OFF_USB` | bool | Turn off RGB underglow when USB is disconnected | n |
| `CONFIG_ZMK_RGB_UNDERGLOW_HUE_STEP` | int | Hue step in degrees (0-359) used by RGB actions | 10 |
| `CONFIG_ZMK_RGB_UNDERGLOW_SAT_STEP` | int | Saturation step in percent used by RGB actions | 10 |
| `CONFIG_ZMK_RGB_UNDERGLOW_BRT_STEP` | int | Brightness step in percent used by RGB actions | 10 |
| `CONFIG_ZMK_RGB_UNDERGLOW_HUE_START` | int | Default hue in degrees (0-359) | 0 |
| `CONFIG_ZMK_RGB_UNDERGLOW_SAT_START` | int | Default saturation percent (0-100) | 100 |
| `CONFIG_ZMK_RGB_UNDERGLOW_BRT_START` | int | Default brightness in percent (0-100) | 100 |
| `CONFIG_ZMK_RGB_UNDERGLOW_SPD_START` | int | Default effect speed (1-5) | 3 |
| `CONFIG_ZMK_RGB_UNDERGLOW_EFF_START` | int | Default effect index from the effect list | 0 |
| `CONFIG_ZMK_RGB_UNDERGLOW_ON_START` | bool | Default on state | y |
| `CONFIG_ZMK_RGB_UNDERGLOW_BRT_MIN` | int | Minimum brightness in percent (0-100) | 0 |
| `CONFIG_ZMK_RGB_UNDERGLOW_BRT_MAX` | int | Maximum brightness in percent (0-100) | 100 |

## Backlight

### Kconfig

| Option | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_BACKLIGHT` | bool | Enables LED backlight | n |
| `CONFIG_ZMK_BACKLIGHT_BRT_STEP` | int | Brightness step in percent | 20 |
| `CONFIG_ZMK_BACKLIGHT_BRT_START` | int | Default brightness in percent | 40 |
| `CONFIG_ZMK_BACKLIGHT_ON_START` | bool | Default backlight state | y |
| `CONFIG_ZMK_BACKLIGHT_AUTO_OFF_IDLE` | bool | Turn off backlight when keyboard goes into idle state | n |
| `CONFIG_ZMK_BACKLIGHT_AUTO_OFF_USB` | bool | Turn off backlight when USB is disconnected | n |