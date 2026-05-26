# Keyboard Scan Configuration

## Mock Driver

Applies to: `compatible = "zmk,kscan-mock"`

| Property | Type | Description | Default |
| --- | --- | --- | --- |
| `event-period` | int | Milliseconds between each generated event |  |
| `events` | array | List of key events to simulate |  |
| `rows` | int | The number of rows in the composite matrix |  |
| `columns` | int | The number of columns in the composite matrix |  |
| `exit-after` | bool | Exit the program after running all events | false |

## GPIO Input

Applies to: `compatible = "zmk,kscan-gpio-direct"`

| Property | Type | Description | Default |
| --- | --- | --- | --- |
| `input-gpios` | GPIO array | Input GPIOs (one per key) |  |
| `debounce-press-ms` | int | Debounce time for key press in milliseconds | 5 |
| `debounce-release-ms` | int | Debounce time for key release in milliseconds | 5 |
| `debounce-scan-period-ms` | int | Time between reads in milliseconds when any key is pressed | 1 |
| `poll-period-ms` | int | Time between reads in milliseconds when no key is pressed | 10 |
| `toggle-mode` | bool | Use toggle switch mode | n |
| `wakeup-source` | bool | Mark this kscan instance as able to wake the keyboard | n |

## GPIO Matrix

Applies to: `compatible = "zmk,kscan-gpio-matrix"`

| Property | Type | Description | Default |
| --- | --- | --- | --- |
| `row-gpios` | GPIO array | Matrix row GPIOs in order |  |
| `col-gpios` | GPIO array | Matrix column GPIOs in order |  |
| `debounce-press-ms` | int | Debounce time for key press in milliseconds | 5 |
| `debounce-release-ms` | int | Debounce time for key release in milliseconds | 5 |
| `debounce-scan-period-ms` | int | Time between reads in milliseconds when any key is pressed | 1 |
| `diode-direction` | string | The direction of the matrix diodes | `"row2col"` |
| `poll-period-ms` | int | Time between reads in milliseconds when no key is pressed | 10 |