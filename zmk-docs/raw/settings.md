# Persistent Settings

ZMK stores persistent settings in flash memory including:
- Bluetooth: Stores pairing keys and MAC addresses associated with hosts
- Split keyboards: Stores pairing keys and MAC addresses for wireless connection between parts
- Output selection: Stores last selected preferred endpoint
- ZMK Studio: Stores runtime keymap modifications and selected physical layouts
- Lighting: Stores current brightness/color/effects for underglow and backlight
- Power management: Stores the state of the external power toggle

## Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_SETTINGS` | bool | Enable settings subsystem | y |
| `CONFIG_ZMK_SETTINGS_SAVE_DEBOUNCE` | int | Milliseconds to wait before saving to flash | 300 |

## Clearing Persisted Settings

For end users, it is recommended to use a special shield named `settings_reset` to build a new firmware file. Regular firmware will need to be flashed afterwards for normal operation.

Note: Since pairing information between split keyboards are also cleared with this process, you will need to clear settings on all parts of a split keyboard.

ZMK Studio-specific settings can be cleared using the "Restore Stock Settings" button in the header of the Studio client.