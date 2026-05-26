# Low Power States

## Idle

In the idle state, peripherals such as displays and lighting are disabled, but the keyboard remains connected to Bluetooth so it can immediately respond when you press a key. Idle state is entered automatically after a timeout period that is 30 seconds by default.

## Deep Sleep

In the deep sleep state, the keyboard enters a software power-off state. Among others, this:

- Disconnects the keyboard from all Bluetooth connections
- Disables any peripherals such as displays and lighting
- If possible, disables external power output
- Clears the contents of RAM, including any unsaved Studio changes

## Soft Off

The feature is intended as an alternative to using a hardware switch to physically cut power from the battery to the keyboard. This can be useful for existing PCBs not designed for wireless that don't have a power switch, or for new designs that favor a push button on/off like found on other devices. It yields power savings comparable to the deep sleep state.

note

The device enters the same software power-off state as in deep sleep, but is significantly more restrictive in the sources which can wake it. Power is not technically removed from the entire system, unlike a hardware switch.

A device can be put in the soft off state by:

- Triggering a hardware-defined dedicated GPIO pin, if one exists;
- Triggering the soft off behavior from the keymap.

Once in the soft off state, the device can only be woken up by:

- Triggering any GPIO pin specified to enable waking from sleep, if one exists;
- Pressing a reset button found on the device.

The GPIO pin used to wake from sleep can be a hardware-defined one, such as for a dedicated on-off push button, or it can be a single specific key switch reused for waking up (which may be accidentally pressed, e.g. while the device is being carried in a bag). To allow the simultaneous pressing of multiple key switches to trigger and exit soft off, some keyboards make use of additional hardware to integrate the dedicated GPIO pin into the keyboard matrix.

### Config

Soft off must be enabled via its corresponding config before it can be used.

### Using Soft Off