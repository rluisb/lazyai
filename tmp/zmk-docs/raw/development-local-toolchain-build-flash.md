# Building and Flashing

From here on, building and flashing ZMK should all be done from the `app/` subdirectory of the ZMK checkout:

```
cd app
```

> warning
> 
> If this is not done, you will encounter errors such as: `ERROR: source directory "." does not contain a CMakeLists.txt; is this really what you want to build?`

## Building

Building a particular keyboard is done using the `west build` command. Its usage slightly changes depending on if your build is for a keyboard with an onboard MCU or one that uses an MCU board add-on.

## Multi-CPU and Dual-Chip Bluetooth Boards

Zephyr supports running the Bluetooth host and controller on separate processors. In such a configuration, ZMK always runs on the host processor, but you may need to build and flash separate firmware for the controller.

### nRF5340

To build and flash the firmware for the nRF5340 development kit's network core, run the following command from the root of the ZMK repo.

## Flashing