# New Board

### Add Variant to board.yml

Edit your existing `board.yml` file to add the variant:

### Add Kconfig for ZMK Variant

Add a file to your board's folder called `<your-board>_zmk_defconfig`. This file will be used to set Kconfig flags specific to the ZMK variant.

Make sure you know what each Kconfig flag does before you enable it. Some flags may be incompatible with certain hardware, or have other adverse effects.

These flags are typically a subset of the following:
- Enable MPU
- Enable clock control
- Enable reset by default
- Enable bootloader support
- Enable settings support
- Enable USB HID
- Enable BLE HID