# Battery Sensing

If your keyboard is using one of the boards supported in ZMK it will already be configured to sense and report battery levels. Below instructions are only intended for users defining and using a custom board.

To enable a battery sensor on a new board, add the driver for the sensor to your board's `.dts` file. ZMK provides two drivers for estimating the battery level using its voltage:

- `zmk,battery-voltage-divider`
- `zmk,battery-nrf-vddh`

See the battery level configuration page for the configuration supported by each driver provided by ZMK.

Zephyr also provides some drivers for fuel gauge ICs such as the TI bq274xx series and Maxim MAX17xxx series. If you use a battery sensor that does not have an existing driver, you will need to write a new driver that supports the `SENSOR_CHAN_GAUGE_STATE_OF_CHARGE` sensor channel and contribute it to Zephyr or ZMK.

Once you have the sensor driver defined, add a `zmk,battery` property to the `chosen` node and set it to reference the sensor node.