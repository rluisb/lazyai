# Encoders

Existing support for encoders in ZMK is focused around the five pin EC11 rotary encoder with push button design used in the majority of current keyboard and macropad designs.

## Enabling EC11 Encoders

To enable encoders for boards that have existing encoder support, uncomment the `CONFIG_EC11=y` and `CONFIG_EC11_TRIGGER_GLOBAL_THREAD=y` lines in your board's .conf file in your `zmk-config/config` folder. Save and push your changes, then download and flash the new firmware.

## Customizing EC11 Encoder Behavior

Encoder behavior in ZMK is configured in two different locations as the push button and rotation behaviors are handled in two separate ways.

### Push Button

Keyboards and macropads with encoder support will typically take the two EC11 pins responsible for the push button and include them as part of the matrix for the keys. To configure what is sent by the push button, find the encoder's position in the keyboard matrix and assign it a behavior the same as you would any other key.

### Rotation

Rotation is handled separately as a type of sensor. The behavior for this is set in `sensor-bindings`. See Sensor Rotation for customizing this behavior.

Here, the left encoder is configured to control volume up and down while the right encoder sends either Page Up or Page Down.

## Adding Encoder Support

See the Hardware Integration page for encoders for instructions on adding them to your keyboard.