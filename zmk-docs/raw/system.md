# System Configuration

## Bluetooth

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_BT` | bool | Enable Bluetooth support |  |
| `CONFIG_BT_BAS` | bool | Enable the Bluetooth BAS (battery reporting service) | y |
| `CONFIG_BT_MAX_CONN` | int | Maximum number of simultaneous Bluetooth connections | 5 |
| `CONFIG_BT_MAX_PAIRED` | int | Maximum number of paired Bluetooth devices | 5 |
| `CONFIG_ZMK_BLE` | bool | Enable ZMK as a Bluetooth keyboard |  |
| `CONFIG_ZMK_BLE_CLEAR_BONDS_ON_START` | bool | Clears all bond information from the keyboard on startup | n |
| `CONFIG_ZMK_BLE_CONSUMER_REPORT_QUEUE_SIZE` | int | Max number of consumer HID reports to queue for sending over BLE | 5 |
| `CONFIG_ZMK_BLE_KEYBOARD_REPORT_QUEUE_SIZE` | int | Max number of keyboard HID reports to queue for sending over BLE | 20 |
| `CONFIG_ZMK_BLE_INIT_PRIORITY` | int | BLE init priority | 50 |
| `CONFIG_ZMK_BLE_THREAD_PRIORITY` | int | Priority of the BLE notify thread | 5 |
| `CONFIG_ZMK_BLE_THREAD_STACK_SIZE` | int | Stack size of the BLE notify thread | 768 |
| `CONFIG_ZMK_BLE_PASSKEY_ENTRY` | bool | Experimental: require typing passkey from host to pair BLE connection | n |

## USB

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_USB` | bool | Enable USB drivers |  |
| `CONFIG_USB_DEVICE_VID` | int | The vendor ID advertised to USB | `0x1D50` |
| `CONFIG_USB_DEVICE_PID` | int | The product ID advertised to USB | `0x615E` |
| `CONFIG_USB_DEVICE_MANUFACTURER` | string | The manufacturer name advertised to USB | `"ZMK Project"` |
| `CONFIG_USB_HID_POLL_INTERVAL_MS` | int | USB polling interval in milliseconds | 1 |
| `CONFIG_ZMK_USB` | bool | Enable ZMK as a USB keyboard |  |
| `CONFIG_ZMK_USB_BOOT` | bool | Enable USB Boot protocol support | n |
| `CONFIG_ZMK_USB_INIT_PRIORITY` | int | USB init priority | 50 |