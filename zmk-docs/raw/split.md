# Split Configuration

## Wired Split

### Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_SPLIT_WIRED_ASYNC_RX_TIMEOUT` | int | RX Timeout (in microseconds) before reporting received data | 20 |

### Polling Mode

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_SPLIT_WIRED_POLLING_RX_PERIOD` | int | Number of ticks between calls to poll for split data | 10 |

## BLE Split

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_SPLIT_BLE` | bool | Use BLE to communicate between split keyboard halves | y |
| `CONFIG_ZMK_SPLIT_BLE_CENTRAL_PERIPHERALS` | int | Number of peripherals that will connect to the central | 1 |
| `CONFIG_ZMK_SPLIT_BLE_CENTRAL_BATTERY_LEVEL_FETCHING` | bool | Enable fetching split peripheral battery levels to the central side | n |
| `CONFIG_ZMK_SPLIT_BLE_CENTRAL_BATTERY_LEVEL_PROXY` | bool | Enable central reporting of split battery levels to hosts | n |
| `CONFIG_ZMK_SPLIT_BLE_CENTRAL_BATTERY_LEVEL_QUEUE_SIZE` | int | Max number of battery level events to queue | `CONFIG_ZMK_SPLIT_BLE_CENTRAL_PERIPHERALS` |
| `CONFIG_ZMK_SPLIT_BLE_CENTRAL_POSITION_QUEUE_SIZE` | int | Max number of key state events to queue | 5 |
| `CONFIG_ZMK_SPLIT_BLE_CENTRAL_SPLIT_RUN_STACK_SIZE` | int | Stack size of the BLE split central write thread | 512 |
| `CONFIG_ZMK_SPLIT_BLE_CENTRAL_SPLIT_RUN_QUEUE_SIZE` | int | Max number of behavior run events to queue | 5 |
| `CONFIG_ZMK_SPLIT_BLE_PERIPHERAL_STACK_SIZE` | int | Stack size of the BLE split peripheral notify thread | 756 |
| `CONFIG_ZMK_SPLIT_BLE_PERIPHERAL_PRIORITY` | int | Priority of the BLE split peripheral notify thread | 5 |
| `CONFIG_ZMK_SPLIT_BLE_PERIPHERAL_POSITION_QUEUE_SIZE` | int | Max number of key state events to queue to send to the central | 10 |