# Sticky Key Behavior

## Summary

A sticky key stays pressed until another key is pressed. It is often used for 'sticky shift'. By using a sticky shift, you don't have to hold the shift key to write a capital.

## Behavior Binding

Reference: `&sk`

Parameter #1: The keycode, e.g. `LSHIFT`

### Example
```
&sk LSHIFT
```

You can use any keycode that works for `&kp` as parameter to `&sk`:
```
&sk LG(LS(LA(LCTRL)))
```

## Configuration

### release-after-ms

By default, sticky keys stay pressed for a second if you don't press any other key. You can configure this with the `release-after-ms` setting.

### quick-release

When enabled, the sticky key releases immediately when another key is pressed rather than waiting for that key to be released.

### ignore-modifiers

When enabled, the sticky key won't be released by modifier keys, only by regular key presses.

### Example Configuration

```
&sk {
    release-after-ms = <2000>;
    quick-release;
    /delete-property/ ignore-modifiers;
};
```

This configuration would apply to all sticky keys. This may not be appropriate if using `quick-release` as you'll lose the ability to chain sticky key modifiers. A better approach for this use case would be to create a new behavior.

## Comparison to QMK

In QMK, sticky keys are known as 'one shot mods'.