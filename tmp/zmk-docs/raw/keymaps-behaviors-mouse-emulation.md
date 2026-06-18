# Mouse Emulation Behaviors

## Summary

ZMK provides several behaviors for mouse emulation, including mouse button presses, mouse movement, and mouse scrolling.

## Mouse Button Press

This behavior can press/release up to 5 mouse buttons.

### Behavior Binding

Reference: `&mkp`

Parameter: A `uint8` with bits 0 through 4 each referring to a button.

### Button Defines

| Define | Action |
| --- | --- |
| `MB1`, `LCLK` | Left click |
| `MB2`, `RCLK` | Right click |
| `MB3`, `MCLK` | Middle click |
| `MB4` | Mouse button 4 |
| `MB5` | Mouse button 5 |

Mouse buttons 4 and 5 typically map to "back" and "forward" actions in most applications.

### Examples

Send a left click:
```
&mkp LCLK
```

Send press of the fourth mouse button:
```
&mkp MB4
```

### Input Processors

To apply input processors to `&mkp`, reference `&mkp_input_listener`:

```
&mkp_input_listener {
    input-processors = <&zip_temp_layer 2 2000>;
};
```

## Mouse Move

This behavior sends mouse X/Y movement events to the connected host.

### Behavior Binding

Reference: `&mmv`

Parameter: A `uint32` with 16-bits each used for vertical and horizontal max velocity.

### Predefined Move Values

| Define | Action |
| --- | --- |
| `MOVE_UP` | Move up |
| `MOVE_DOWN` | Move down |
| `MOVE_LEFT` | Move left |
| `MOVE_RIGHT` | Move right |

### Example

Move the mouse:
```
&mmv MOVE_RIGHT
```

### Input Processors

For applying input processors to `&mmv`:

```
&mmv_input_listener {
    input-processors = <&zip_temp_layer 2 2000>;
};
```

## Mouse Scroll

This behavior sends vertical and horizontal scroll events to the connected host.

### Behavior Binding

Reference: `&msc`

Parameter: A `uint32` with 16-bits each used for vertical and horizontal velocity.

### Scroll Defines

| Define | Action |
| --- | --- |
| `SCRL_UP` | Scroll up |
| `SCRL_DOWN` | Scroll down |
| `SCRL_LEFT` | Scroll left |
| `SCRL_RIGHT` | Scroll right |

### Example

Scroll down:
```
&msc SCRL_DOWN
```