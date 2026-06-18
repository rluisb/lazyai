# Bluetooth Configuration

## Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_BLE_PASSKEY_ENTRY` | bool | Enable passkey entry during pairing for enhanced security | n |
| `CONFIG_BT_GATT_ENFORCE_SUBSCRIPTION` | bool | Low level setting for GATT subscriptions | y |
| `CONFIG_BT_DEVICE_APPEARANCE` | int | Bluetooth device appearance value | 961 |
| `CONFIG_ZMK_BLE_EXPERIMENTAL_CONN` | bool | Enables settings to improve connection stability (disables 2M PHY) | n |
| `CONFIG_ZMK_BLE_EXPERIMENTAL_SEC` | bool | Enables BT Secure Connection passkey entry and key overwrite | n |
| `CONFIG_ZMK_BLE_EXPERIMENTAL_FEATURES` | bool | Aggregate config enabling both EXPERIMENTAL_CONN and EXPERIMENTAL_SEC | n |