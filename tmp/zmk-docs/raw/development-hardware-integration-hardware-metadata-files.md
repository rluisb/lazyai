# Hardware Metadata Files

## Overview

ZMK makes use of an additional metadata YAML file for all boards and shields to provide high level information about the hardware to be incorporated into setup scripts/utilities, website hardware list, etc.

The naming convention for metadata files is `{item_id}.zmk.yml`, where the `item_id` is the board/shield identifier, including version information but excluding any optional split `_left`/`_right` suffix, e.g. `corne.zmk.yml` or `nrfmicro_nrf52840.zmk.yml`.

## Example File

Here is a sample `corne.zmk.yml` file from the repository:

- `keys` - Any board or shield that contains keyboard keys should include this feature.
- `display` - Indicates the hardware includes a display for use with the ZMK display functionality.
- `encoder` - Indicates the hardware contains one or more rotary encoders.
- `underglow` - Indicates the hardware includes underglow LEDs.
- `backlight` - Indicates the hardware includes backlight LEDs.
- `studio` - Indicates the keyboard is ready to use with ZMK Studio.
- `pointer` (future) - Used to indicate the hardware includes one or more pointer inputs, e.g. joystick, touchpad, or trackpoint.

## Schema

ZMK uses a JSON Schema file to validate metadata files and sure all required properties are present, and all other optional properties provided conform to the expected format. You can validate all metadata files in the repository by running `west metadata check` from your configured ZMK repository.