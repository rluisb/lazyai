# List of Keycodes

This is the reference page for keycodes used by behaviors. Use the table of contents (on the right or the top) for easy navigation.

Take extra notice of the spelling of the keycodes, especially the shorthand spelling.
Otherwise, it will result in an elusive parsing error!

In the below tables, there are keycode pairs with similar names where one variant has a `K_` prefix and another `C_`.
These variants correspond to similarly named usages from different HID usage pages,
namely the "keyboard/keypad" and "consumer" ones respectively.

`K_` = Keyboard/Keypad usage page
`C_` = Consumer usage page

In practice, some OS and applications might listen to only one of the variants.
You can use the values in the compatibility columns below to assist you in selecting which one to use.

Columns: W = Windows, L = Linux, A = macOS, m = Android, i = iOS

## Keyboard

### Letters

| Names | Description |
| --- | --- |
| `A` - `Z` | Letter keys |

### Numbers

| Names | Description |
| --- | --- |
| `N1` - `N0` | Number keys 1-0 |

### Modifiers

| Names | Description |
| --- | --- |
| `LSHIFT`, `RSHIFT` | Left/Right Shift |
| `LCTRL`, `RCTRL` | Left/Right Control |
| `LALT`, `RALT` | Left/Right Alt |
| `LGUI`, `RGUI` | Left/Right GUI (Windows/Command) |
| `LS(LA(LCTRL))` | Chaining modifiers example |

### Navigation Keys

| Names | Description |
| --- | --- |
| `UP` | Up arrow |
| `DOWN` | Down arrow |
| `LEFT` | Left arrow |
| `RIGHT` | Right arrow |
| `HOME` | Home |
| `END` | End |
| `PGUP` | Page Up |
| `PGDN` | Page Down |

### Function Keys

| Names | Description |
| --- | --- |
| `F1` - `F24` | Function keys |

### System Keys

| Names | Description |
| --- | --- |
| `SYS_RESET` | System Reset |
| `SYS_DFU` | System DFU |
| `SYS_REBOOT` | System Reboot |
| `CLR_BOND` | Clear Bluetooth bonds |

### Media Keys

| Names | Description | W | L | A | m | i |
| --- | --- | --- | --- | --- | --- | --- |
| `C_VOLUME_UP` `C_VOL_UP` | Volume Up Consumer | ⭐ | ⭐ | ⭐ | ⭐ | ⭐ |
| `K_VOLUME_UP` `K_VOL_UP` | Volume Up Keyboard | ❌ | ⭐ | ⭐ | ⭐ | ❔ |
| `C_VOLUME_DOWN` `C_VOL_DN` | Volume Down Consumer | ⭐ | ⭐ | ⭐ | ⭐ | ⭐ |
| `K_VOLUME_DOWN` `K_VOL_DN` | Volume Down Keyboard | ❌ | ⭐ | ⭐ | ⭐ | ❔ |
| `C_MUTE` | Mute Consumer | ⭐ | ⭐ | ⭐ | ⭐ | ⭐ |
| `K_MUTE` | Mute Keyboard | ❌ | ⭐ | ⭐ | ⭐ | ❔ |

### Editing Keys

| Names | Description |
| --- | --- |
| `C_AC_CUT` | Cut Consumer AC |
| `K_CUT` | Cut Keyboard |
| `C_AC_COPY` | Copy Consumer AC |
| `K_COPY` | Copy Keyboard |
| `C_AC_PASTE` | Paste Consumer AC |
| `K_PASTE` | Paste Keyboard |

### Numpad

| Names | Description |
| --- | --- |
| `KP_N1` - `KP_N0` | Numpad 1-0 |
| `KP_PLUS` | Numpad + |
| `KP_MINUS` | Numpad - |
| `KP_MULTIPLY` | Numpad * |
| `KP_DIVIDE` | Numpad / |
| `KP_DOT` | Numpad . |
| `KP_EQUAL` | Numpad = |
| `KP_LEFT_PARENTHESIS` `KP_LPAR` | Numpad ( |
| `KP_RIGHT_PARENTHESIS` `KP_RPAR` | Numpad ) |
| `KP_ENTER` | Numpad Enter |

### Special Keys

| Names | Description |
| --- | --- |
| `ENTER` | Enter |
| `ESC` | Escape |
| `TAB` | Tab |
| `SPACE` | Space |
| `BACKSPACE` | Backspace |
| `CAPSLOCK` | Caps Lock |
| `PRINTSCREEN` `PSCRN` | Print Screen |
| `PAUSE_BREAK` | Pause / Break |
| `DELETE` | Delete |
| `INSERT` | Insert |

## Layers

| Names | Description |
| --- | --- |
| `&mo` | Momentary layer |
| `&lt` | Layer tap |
| `&tg` | Layer toggle |
| `&tt` | Layer tap-toggle |

## Modifiers in Bindings

Modifiers can be combined using the following format:
- `LSHIFT` - Left Shift
- `RSHIFT` - Right Shift
- `LCTRL` - Left Control
- `RCTRL` - Right Control
- `LALT` - Left Alt
- `RALT` - Right Alt
- `LGUI` - Left GUI
- `RGUI` - Right GUI

Chained modifiers example: `&kp LG(LS(LA(LCTRL)))` = Ctrl+Alt+Shift+Gui

## Mouse Keycodes

| Names | Description |
| --- | --- |
| `MB1` `LCLK` | Left click |
| `MB2` `RCLK` | Right click |
| `MB3` `MCLK` | Middle click |
| `MB4` | Mouse button 4 |
| `MB5` | Mouse button 5 |
| `MOVE_UP` | Move up |
| `MOVE_DOWN` | Move down |
| `MOVE_LEFT` | Move left |
| `MOVE_RIGHT` | Move right |
| `SCRL_UP` | Scroll up |
| `SCRL_DOWN` | Scroll down |
| `SCRL_LEFT` | Scroll left |
| `SCRL_RIGHT` | Scroll right |

## Consumer Keycodes (C_ prefix)

These send HID consumer usage page codes instead of keyboard codes:
- `C_PLAY` / `C_PP` - Play/Pause
- `C_STOP` - Stop
- `C_NEXT` - Next track
- `C_PREV` - Previous track
- `C_RECORD` - Record
- `C_EJECT` - Eject
- `C_MUTE` - Mute
- `C_VOLUME_UP` / `C_VOL_DOWN` - Volume
- And many more application-specific codes

## Notes

For the complete list of keycodes, please refer to the official ZMK documentation at https://zmk.dev/docs/keymaps/list-of-keycodes