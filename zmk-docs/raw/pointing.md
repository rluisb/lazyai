# Pointing Device Configuration

## Advanced Settings

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_INPUT_THREAD_STACK_SIZE` | int | Stack size for the dedicated input event processing thread | 512 (1024 on split peripherals) |

## Input Listener

Applies to: `compatible = "zmk,input-listener"`

| Property | Type | Description |
| --- | --- | --- |
| `device` | phhandle | Input device handle |
| `input-processors` | phandle-array | List of input processors to apply to input events |

### Child Properties

| Property | Type | Description |
| --- | --- | --- |
| `layers` | array | List of layer indexes. This config will apply if any layer in the list is active |
| `input-processors` | phandle-array | List of input processors to apply to input events |
| `process-next` | bool | Whether to continue applying other input processors after this override if it takes effect |

## Input Split

Applies to: `compatible = "zmk,input-split"`

| Property | Type | Description |
| --- | --- | --- |
| `device` | handle | Input device handle |
| `input-processors` | phandle-array | List of input processors to apply to input events |