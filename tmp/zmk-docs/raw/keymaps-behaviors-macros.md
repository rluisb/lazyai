# Macro Behavior

## Summary

The macro behavior allows configuring a list of other behaviors to invoke when the macro is pressed and/or released.

## Macro Definition

Each macro you want to use in your keymap gets defined first, then bound in your keymap.

A macro definition looks like:

```
/ {
    behaviors {
        my_macro: macro_p {
            compatible = "zmk,behavior-macro";
            #binding-cells = <0>;
            label = "MY_MACRO";

            bindings = <&kp S &kp H &kp I &kp F &kp T>;
        };
    };
};
```

The text before the colon (`:`) in the declaration of the macro node is the "node label", and is the text used to reference the macro in your keymap.

The macro can then be bound in your keymap by referencing it by the label `&my_macro`.

Note: For use cases involving sending a single keycode with modifiers, for instance ctrl+tab, the key press behavior with modifier functions can be used instead of a macro.

### Binding Activation Mode

Macros can be configured to handle binding activation differently:
- `interrupt`: The macro runs to completion but can be interrupted
- `buffer`: Key events are buffered and processed sequentially

### Processing Continuation on Release

Control whether macro processing continues when the binding is released.

### Wait Time

Add delays between behaviors in the macro:
```
wait-ms = <100>;
```

### Tap Time

Set the tap time for key press behaviors in the macro:
```
tap-ms = <50>;
```

### Behavior Queue Limit

Limit how many behaviors can be queued in the macro buffer.

## Parameterized Macros

Macros can accept parameters to make them reusable:

```
/ {
    behaviors {
        parameterized_macro: macro_p {
            compatible = "zmk,behavior-macro";
            #binding-cells = <2>;
            label = "PARAM_MACRO";

            bindings = <&kp A &kp B>;
        };
    };
};
```

Usage:
```
&parameterized_macro X Y
```

## Common Patterns

### Layer Activation

Switch layers with a macro:
```
&mo 1
```

### Keycode Sequences

Send sequences of keycodes:
```
&kp LSHIFT &kp A &kp B &kp C
```

### Unicode Sequences

Send Unicode characters via macros:
```
&kp U+00E9  // é
```

## Convenience C Macro

ZMK provides a C macro for defining simple tap-only macros more conveniently. See the full documentation for implementation details.