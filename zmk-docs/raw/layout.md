# Layout Configuration

## Matrix Transform

Applies to: `compatible = "zmk,matrix-transform"`

| Property | Type | Description | Default |
| --- | --- | --- | --- |
| `rows` | int | Number of rows in the transformed matrix |  |
| `columns` | int | Number of columns in the transformed matrix |  |
| `row-offset` | int | Adds an offset to all rows before looking them up in the transform | 0 |
| `col-offset` | int | Adds an offset to all columns before looking them up in the transform | 0 |
| `map` | array | A list of position transforms |  |

The `map` array should be defined using the `RC()` macro from dt-bindings/zmk/matrix_transform.h. It should have one item per logical position in the keymap.