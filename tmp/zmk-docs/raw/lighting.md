# Lighting

ZMK supports two distinct systems in order to control lighting hardware integrated into keyboards.
Your keyboard likely uses only one type, depending on the type of LED hardware it supports:

- RGB underglow system controls LED strips composed of addressable RGB LEDs. Most keyboards that have multi-color lighting utilizes these.
- Backlight system controls parallel-connected, non-addressable, single color LEDs. These are found on keyboards that have a single color backlight that only allows for brightness control.

warning

Although the naming of the systems might imply it, which system you use typically does not depend on the physical location of the LEDs. Instead, you should use the one that supports the LED hardware type that your keyboard has, as described above.

## RGB Underglow

RGB underglow is a feature used to control "strips" of RGB LEDs. Most of the time this is called underglow and creates a glow underneath the board using a ring of LEDs around the edge, hence the name. However, this can be extended to be used to control anything from a single LED to a long string of LEDs anywhere on the keyboard.

info

RGB underglow can also be used for per-key lighting. If you have RGB LEDs on your keyboard, this is what you want. For PWM/single color LEDs, see Backlight section below.

For split keyboards, set `chain-length` to the number of LEDs installed on each half.

### Configuring RGB Underglow

See RGB underglow configuration.

### Adding RGB Underglow Support to a Keyboard

See RGB underglow hardware integration page on adding underglow support to a ZMK keyboard.

## Backlight

Backlight is a feature used to control an array of LEDs, usually placed through or under switches.

info

Unlike RGB underglow, backlight can only control single color LEDs. Additionally, because backlight LEDs all receive the same power, it's not possible to dim individual LEDs.

### Enabling Backlight