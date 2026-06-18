# Pointing Devices

ZMK's pointing device support builds upon the Zephyr input API to offer pointing/mouse functionality with various hardware. A limited number of input drivers are available in the Zephyr version currently used by ZMK, but additional drivers can be found in external modules for a variety of hardware.

Pointing devices are also supported on split peripherals, with some additional configuration using the input split device. The configuration details will thus vary depending on if you are adding a pointing device to a split peripheral as opposed to a unibody keyboard or split central part.

## Input Device

Enable `CONFIG_ZMK_POINTING=y` in your configuration.

Second, the input listener that is used by the central side is added here but disabled by default. This is so that keymaps (which are included for both central and peripheral builds) can reference the listener to add input processors without failing with an undefined reference error.

Input splits need to be nested under a parent node that properly sets `#address-cells = <1>` and `#size-cells = <0>`. These settings are what allow us to use a single integer number for the `reg` value.