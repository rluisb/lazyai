# Bootloader Integration

The information on this page is only relevant for **boards**, not **shields**.

The `&bootloader` behavior requires properly set up boot mode support to function properly. The behavior operates by setting the boot mode, resetting, and then relies on an SoC/bootloader specific early init hook to enter the bootloader when the boot mode is found to have been set.

Most of the SoCs actively supported by ZMK rely on a generic retained memory driver to store the boot mode between restarts, and additional configuration is required when using a second stage bootloader like the Adafruit nRF52 Bootloader or tinyuf2.

## Magic Value Bootloaders

Most "second stage" bootloaders will enter bootloader mode on startup when a specific magic value is found in a specific reserved location in memory. For those bootloaders, an extra mapping layer is used to map the Zephyr "bootloader mode" retained value to the magic value expected by the bootloader.

The following bootloaders of this type are supported, see those pages for details on the additional configuration needed:

## Jump-To Bootloaders

Several SoCs use bootloaders that can be directly jumped to from early init code in the firmware. For these situations, the only setup required is a retained mem instance that can retain the set boot mode after the reset, in order for the early initailization code to check the value and then jump to the bootloader.

The following bootloaders of this type are supported, see those pages for details on the additional configuration needed: