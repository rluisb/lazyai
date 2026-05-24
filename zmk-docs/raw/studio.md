# ZMK Studio Configuration

## Keymaps

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_KEYMAP_LAYER_NAME_MAX_LEN` | int | Max allowable keymap layer display name | 20 |

## Locking

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_STUDIO_LOCKING` | bool | Enable/disable locking for ZMK Studio | y |
| `CONFIG_ZMK_STUDIO_LOCK_IDLE_TIMEOUT_SEC` | int | Seconds of inactivity before automatically locking | 500 |
| `CONFIG_ZMK_STUDIO_LOCK_ON_DISCONNECT` | bool | Whether to automatically lock again whenever ZMK Studio disconnects | y |

## Transport/Protocol Details

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_STUDIO_TRANSPORT_BLE_PREF_LATENCY` | int | Lower latency to request while ZMK Studio is active | 10 |
| `CONFIG_ZMK_STUDIO_RPC_THREAD_STACK_SIZE` | int | Stack size for the dedicated RPC thread | 1800 |
| `CONFIG_ZMK_STUDIO_RPC_RX_BUF_SIZE` | int | Number of bytes available for buffering incoming messages | 30 |
| `CONFIG_ZMK_STUDIO_RPC_TX_BUF_SIZE` | int | Number of bytes available for buffering outgoing messages | 64 |