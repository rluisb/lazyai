# Physical Layouts

A physical layout is a devicetree entity that aggregates all details about a certain possible keyboard layout. It contains:

By convention, physical layouts and any position maps are defined in a separate file called `<your keyboard>-layouts.dtsi`. This file should then be imported by the appropriate file, such as an `.overlay`, `.dts`, or a `.dtsi`.

## Basic Physical Layout

A bare physical layout without the `keys` property looks like this:

```
/ {
  physical_layout0: physical_layout_0 {
    compatible = "zmk,physical-layout";
    display-name = "Default Layout";
  };
};
```

## Keys Property

The `keys` property contains a 2D array of key positions. Each key can have attributes like width, height, x position, y position, rotation, and rotation origin.

You can specify negative values in devicetree using parentheses around it, e.g. `(-3000)` for a 30 degree counterclockwise rotation.

## Position Maps

Position maps allow multiple layouts to share the same physical layout while mapping keys to different positions.