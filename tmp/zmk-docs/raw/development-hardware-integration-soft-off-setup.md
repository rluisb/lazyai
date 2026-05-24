# Adding Soft Off To A Keyboard

`(GPIO_ACTIVE_LOW | GPIO_PULL_UP)`
`col2row`

GPIO keys are defined using child nodes under the `gpio-keys` compatible node. Each child needs just one property defined.

## KScan sideband behavior

The kscan sideband behavior driver will be used to trigger the soft off behavior "out of band" from the normal keymap processing. To do so, it will decorate/wrap an underlying kscan driver.

With a simple direct pin setup, the direct kscan driver can be used with a GPIO key, to make a small "side matrix":

With that in place, the kscan sideband behavior will wrap the new driver:

Critically, the `column` and `row` values would correspond to the location of the added entry. The properties of the `kscan-sideband-behaviors` node can be found in the appropriate configuration section.

Alternatively, if you wish to integrate a dedicated GPIO pin into a key switch combination using a direct kscan, tie all of the MCU pins that you wish to combine to the dedicated GPIO pin through an OR gate.

To integrate the dedicated GPIO pin into your matrix, you will need to tie multiple switch outputs in the matrix together through AND gates and connect the result to the dedicated GPIO pin. This way you can use a (hardware defined) key combination in your existing keyboard matrix to trigger soft on/off.