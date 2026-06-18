# Hold-Tap Behavior

## Summary

The hold-tap behavior invokes different behaviors depending on whether a key is held or tapped. It is the foundation for mod-tap and layer-tap behaviors.

## Mod-Tap

The "mod-tap" behavior sends a different key press, depending on whether it's held or tapped.

If you hold the key for longer than 200ms or press any other key while it is held, the first keycode ("mod") is sent.
If you tap the key (release before 200ms), the second keycode ("tap") is sent.

By default, mod-tap is configured with the "hold-preferred" `flavor`.

### Behavior Binding

Reference: `&mt`

Parameter #1: The keycode to be sent when held (usually a modifier), e.g. `LSHIFT`
Parameter #2: The keycode to sent when used as a tap, e.g. `A`, `B`.

Example:
```
&mt LSHIFT A
```

### Configuration

You can adjust the default properties of the mod-tap behavior using its node like so:

```
&mt {
    tapping-term-ms = <200>;
};
```

## Layer-Tap

The "layer-tap" behavior works similarly to the mod-tap behavior, but instead of outputting one of two keys, it activates a specified layer as its "hold" action.

By default, layer-tap is configured with the "tap-preferred" `flavor`.

### Behavior Binding

Reference: `&lt`

Parameter: The layer number to enable while held, e.g. `1`
Parameter: The keycode to send when tapped, e.g. `A`

Example:
```
&lt 1 A
```

### Configuration

For layer-tap, you can also configure the tapping term:

```
&lt {
    tapping-term-ms = <200>;
};
```

## Hold-Tap Flavors

The hold-tap behavior supports different "flavors" that determine which binding is preferred:

- **hold-preferred**: When another key is pressed while holding, the hold binding is used
- **tap-preferred**: When another key is pressed while holding, the tap binding is used
- **balanced**: Prefers neither, uses timing to decide

## Custom Hold-Tap Definitions

You can define custom hold-tap behaviors with specific configurations:

```
/ {
    behaviors {
        my_modtap: hold_tap_modtap {
            compatible = "zmk,behavior-hold-tap";
            #binding-cells = <2>;
            flavor = "hold-preferred";
            tapping-term-ms = <200>;
            hold-trigger-key-positions = <...>;
            hold-trigger-on-release;
        };
    };
};
```

## Advanced Configuration

### hold-trigger-key-positions

For home-row mods, it is recommended to use this property with `hold-trigger-on-release` so that modifiers on the same hand can be combined.

`hold-trigger-key-positions` is an array of key position indexes. Key positions are numbered sequentially according to your keymap, starting with 0.

### hold-trigger-on-release

When set, the hold binding is only triggered when the hold-tap key is released, and only if no other key was pressed during the hold.

## Use Cases

- **Home-row mods**: Using mod-tap with hold-trigger-on-release for comfortable modifier keys on home row
- **Layer-tap**: Using layer-tap to access a layer when holding a key
- **Space-cadet shift**: Using hold-tap for shift-on-hold, space-on-tap patterns