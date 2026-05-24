# Keymap Configuration

## Keymap

### Devicetree

Applies to: `compatible = "zmk,keymap"`

Definition file: zmk/app/dts/bindings/zmk,keymap.yaml

The `zmk,keymap` node itself has no properties. It should have one child node per layer of the keymap, starting with the default layer (layer 0).

Each child node can have the following properties:

| Property | Type | Description |
| --- | --- | --- |
| `display-name` | string | Name for the layer in ZMK Studio and on displays |
| `bindings` | phandle-array | List of key behaviors, one per key |
| `sensor-bindings` | phandle-array | List of sensor behaviors, one per sensor |

Items for `bindings` must be listed in the order the keys are defined in the keyboard scan configuration.

Items for `sensor-bindings` must be listed in the order the sensors are defined.

## Keymap Sensors

### Devicetree

Applies to: `compatible = "zmk,keymap-sensors"`

| Property | Type | Description |
| --- | --- | --- |
| `sensors` | phandles | List of sensor nodes |

The following types of nodes can be used as a sensor:
- `alps,ec11`