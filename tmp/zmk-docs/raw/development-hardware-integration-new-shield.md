# New Keyboard Shield

The keymap should match the order of the keys in the matrix transform exactly, left to right, top to bottom (they are both 1 dimensional arrays rearranged with newline characters for better legibility). See Keymaps for information on defining keymaps in ZMK. If you wish to use ZMK Studio with your keyboard, make sure to assign the ZMK Studio unlocking behavior to a key in your keymap.

## Metadata

ZMK makes use of an additional metadata YAML file for all boards and shields to provide high level information about the hardware to be incorporated into setup scripts/utilities, website hardware list, etc.

## Testing

Once you've defined everything as described above, you can build your firmware to make sure everything is working. If you wish to test that your keyboard works with ZMK Studio, you'll also need to follow the instructions for enabling Studio.

### GitHub Actions

To use GitHub Actions to test, push the files defining the keyboard to GitHub. Next, update the `build.yaml` of your `zmk-config` to build your keyboard.

### Local Toolchain

## Kscan

## Matrix Transform

## Keymap