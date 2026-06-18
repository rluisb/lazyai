# Split Keyboards

## Split Transports

### Bluetooth

Bluetooth is the most well tested and flexible transport available in ZMK. Using Bluetooth, a central can connect to multiple peripherals, enabling the use of a dongle to improve battery life, or allowing for multi-part split keyboards.

This transport will be enabled for designs that set `CONFIG_ZMK_SPLIT=y` and have `CONFIG_ZMK_BLE=y` set by a supported MCU/controller.

### Full-Duplex Wired (UART)

The full-duplex wired UART transport is a recent addition, and is intended for testing by early adopters. It allows for fully functional communication between one central and one peripheral. Unlike the Bluetooth transport, the central cannot currently have more than one peripheral connected via wired split.

This transport will be enabled for designs that set `CONFIG_ZMK_SPLIT=y` and have a node with `compatible = "zmk,wired-split";` present in their devicetree configuration.

Full Duplex vs Half Duplex

Full-duplex UART requires the use of two wires connecting the halves. Future half-duplex (single-wire) UART support, which is planned, will allow using wired ZMK with designs such as the Corne, Sweep, etc. that use only a single GPIO pin for bidirectional communication between split sides.

Until half-duplex support is completed, those particular designs will not work with the wired split transport, and can only be used with the Bluetooth transport.

### Runtime Switching

## Latency Considerations

Since peripherals communicate through centrals, the key and sensor events originating from them will naturally have a larger latency, especially with a wireless split communication protocol. For the currently used BLE-based transport, split communication increases the average latency by 3.75ms with a worst case increase of 7.5ms.

Also see the reference section on split keyboards configuration where the relevant symbols include `CONFIG_ZMK_SPLIT` that enables the feature, `CONFIG_ZMK_SPLIT_ROLE_CENTRAL` which sets the central role and `CONFIG_ZMK_SPLIT_BLE_CENTRAL_PERIPHERALS` that sets the number of peripherals.