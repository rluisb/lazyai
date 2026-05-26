# Behavior Configuration

## Two Axis Input

Applies to: `compatible = "zmk,behavior-input-two-axis"`

Definition file: zmk/app/dts/bindings/behaviors/zmk,behavior-input-two-axis.yaml

| Property | Type | Description | Default |
| --- | --- | --- | --- |
| `#binding-cells` | int | Must be `<1>` |  |
| `x-input-code` | int | The relative event code for generated input events for the X-axis |  |
| `y-input-code` | int | The relative event code for generated input events for the Y-axis |  |
| `trigger-period-ms` | int | How many milliseconds between generated input events based on the current speed/direction | 16 |
| `delay-ms` | int | How many milliseconds to delay any processing or event generation when first pressed | 0 |
| `time-to-max-speed-ms` | int | How many milliseconds it takes to accelerate to the current max speed | 0 |
| `acceleration-exponent` | int | The acceleration exponent to apply: `0` - uniform speed, `1` - uniform acceleration, `2` - linear acceleration | 1 |

## Tap Dance

Applies to: `compatible = "zmk,behavior-tap-dance"`

| Property | Type | Description | Default |
| --- | --- | --- | --- |
| `#binding-cells` | int | Must be `<0>` |  |
| `bindings` | phandle array | A list of behaviors from which to select |  |
| `tapping-term-ms` | int | The maximum time (in milliseconds) between taps before an item from `bindings` is triggered | 200 |

## Caps Word

Applies to: `compatible = "zmk,behavior-caps-word"`

Definition file: zmk/app/dts/bindings/behaviors/zmk,behavior-caps-word.yaml

| Property | Type | Description | Default |
| --- | --- | --- | --- |
| `#binding-cells` | int | Must be `<0>` |  |
| `continue-list` | array | List of keycodes which do not deactivate caps lock | `<UNDERSCORE BACKSPACE DELETE>` |
| `mods` | int | A bit field of modifiers to apply | `<MOD_LSFT>` |