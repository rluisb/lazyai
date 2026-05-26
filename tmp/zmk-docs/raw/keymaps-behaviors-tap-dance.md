# Tap-Dance Behavior

## Summary

A tap-dance key invokes a different behavior (e.g. `kp`) corresponding to how many times it is pressed. For example, you could configure a tap-dance key that acts as `LSHIFT` if tapped once, or Caps Lock if tapped twice. The expandability of the number of `bindings` attached to a particular tap-dance is a great way to add more functionality to a single key, especially for keyboards with a limited number of keys. Tap-dances are completely custom, so for every unique tap-dance key, a new tap-dance must be defined in your keymap's `behaviors`.

## Behavior Definition

Tap-dances are user-defined behaviors that must be defined in the behaviors node of your keymap:

```
#include <behaviors.dtsi>
#include <dt-bindings/zmk/keys.h>

/ {
    behaviors {
        td_mt: tap_dance_mod_tap {
            compatible = "zmk,behavior-tap-dance";
            #binding-cells = <0>;
            tapping-term-ms = <200>;
            bindings = <&mt LSHIFT CAPSLOCK>, <&kp LCTRL>;
        };
    };

    keymap {
        compatible = "zmk,keymap";

        default_layer {
            bindings = <&td_mt>;
        };
    };
};
```

## Tap-Dance with 3+ Bindings

You can create tap-dances with any number of bindings:

```
#include <behaviors.dtsi>
#include <dt-bindings/zmk/keys.h>

/ {
    behaviors {
        td0: tap_dance_0 {
            compatible = "zmk,behavior-tap-dance";
            #binding-cells = <0>;
            tapping-term-ms = <200>;
            bindings = <&kp N1>, <&kp N2>, <&kp N3>;
        };
    };

    keymap {
        compatible = "zmk,keymap";

        default_layer {
            bindings = <&td0>;
        };
    };
};
```

## Configuration Options

### tapping-term-ms

The time in milliseconds that determines the difference between a tap and a hold.

### bindings

An array of behaviors to cycle through based on the number of taps.

## Usage

After defining a tap-dance behavior with a label (e.g., `td_mt`), you can reference it in your keymap using `&td_mt`. Each press of the key cycles to the next behavior in the bindings array, wrapping back to the first after the last.