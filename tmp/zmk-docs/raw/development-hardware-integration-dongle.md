# Keyboard Dongle

The dongle overlay file needs to define a physical layout with a mock kscan, which will be used by the central side to read input events and distribute them to the peripherals:

```
physical_layout0
default_transform
```

If there are multiple physical layouts in the file, you will need to copy over all of the remaining matrix transformations and assign them to their corresponding physical layout.

Copy the physical layout node into your `my_keyboard_dongle.overlay` file. Make sure the matrix transform is assigned to it, and select it in the `chosen` node.

## Building the Firmware

Add the appropriate lines to your `build.yaml` file to build the firmware for your dongle. Also add some CMake arguments using `cmake-args` to the existing parts of your keyboard, turning them into peripherals for your dongle.