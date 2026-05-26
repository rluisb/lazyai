# Hardware Integration

This section of the documentation describes steps necessary to get ZMK running on a keyboard, including basic keyboard functionality as well as additional features such as encoders.
Please see pages in the sidebar for guides and reference that describe different aspects of this integration.

The foundational elements needed to get a specific keyboard working with ZMK can be broken down into:

A physical layout that describes the electrical and physical structure of the keyboard, referring to:
  + A kscan driver, which most frequently uses `compatible = "zmk,kscan-gpio-matrix"` for GPIO matrix based keyboards, or `compatible = "zmk,kscan-gpio-direct"` for direct wires.
  + A matrix transform, which defines how the kscan row/column events are translated into logical "key positions".
  + An optional description of physical key positions and sizes, in order to visualize the keyboard accurately in ZMK Studio.
 A keymap, which binds each key position to a behavior, e.g. key press, mod-tap, momentary layer, in a set of layers.
 Other, optional configuration items to support features such as encoders or lighting systems.

Devicetree files:

 A `<shield_name>.overlay` file which is a devicetree overlay file, containing definitions including but not limited to:
  + Kscan, matrix transform and physical layout devicetree nodes as described above, where the kscan node uses the interconnect nexus node aliases such as `&pro_micro` for GPIO pins.
  + A chosen node including at least the `zmk,physical-layout` property, referring to the defined node.
 A `<keyboard_name>.keymap` file that includes the default keymap for that keyboard. Users will be able to override this keymap in their user configs.

And other miscellaneous ones:

 A `<keyboard_name>.zmk.yml` file containing metadata for the keyboard.