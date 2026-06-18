# Keymaps & Behaviors

ZMK uses a declarative approach to keymaps, using devicetree syntax to configure them in a `<keyboard>.keymap` file.
This declarative configuration defines the key mappings, behaviors used within them and the configuration of certain features.
These keymaps can then be updated dynamically/at runtime using the ZMK Studio feature,
which allows keymap modifications over USB or BLE.

## Stock and User Keymaps

All keyboard definitions (complete boards or shields) include the default keymap for that keyboard,
so ZMK can produce a "stock" firmware for that keyboard without any further modifications.
When users complete the user setup, the stock `.keymap` file is copied to the
user config directory, which can be used to customize the keymap to each user's liking.

## Keycode Header Files

Keymaps use devicetree includes to bring in keycodes:

```
#include <dt-bindings/zmk/keys.h>
#include <dt-bindings/zmk/modifiers.h>
```

The first include brings in the defines for all the keycodes (e.g. `A`, `N1`, `C_PLAY`) and the modifiers (e.g. `LSHIFT`) used for various behavior bindings.

### Root Devicetree Node

All the remaining keymap nodes will be nested inside of the root devicetree node, like so:

```
/ {
    // Everything else goes here!
};
```

### Keymap Node

Nested under the devicetree root, is the keymap node. The node name itself is not critical, but the node MUST have a property `compatible = "zmk,keymap"` in order to be used by ZMK.

```
keymap {
    compatible = "zmk,keymap";

    // Layer nodes go here!
};
```

### Layers

Layers are nodes nested inside the keymap node. Layer nodes require a `bindings` property, which is an array of behavior bindings for each key position on that layer.

Layers also support an optional `label` property, which is a human-readable string for the layer.

```
keymap {
    compatible = "zmk,keymap";

    default_layer {
        label = "base";
        bindings = <&kp Q &kp W &kp E &kp R &kp T ...>;
    };
};
```

## Behaviors

ZMK implements the concept of "behaviors", which can be bound to a certain key position, sensor (encoder), or layer, to perform certain actions when events occur for that binding (e.g. when a certain key position is pressed or released, or an encoder triggers a rotation event).