# EC11 Encoders

EC11 encoder support can be added to your board or shield by adding the appropriate lines to your board/shield's configuration (.conf), devicetree (.dtsi/.overlay), and keymap (.keymap) files.

## Configuration File

In your configuration file you will need to add the following lines so that the encoders can be enabled/disabled:

```
# Uncomment to enable encoder
# CONFIG_EC11=y
# CONFIG_EC11_TRIGGER_GLOBAL_THREAD=y
```

These should be commented by default for encoders that are optional/can be swapped with switches, but can be uncommented if encoders are part of the default design.

In this example, a `left_encoder` and `right_encoder` are both added. Additional encoders can be added with spaces separating each, and the order they are added here determines the order in which you define their behavior in your keymap.

In addition, a default value for the number of times the sensors trigger the bound behavior per full rotation is set via the `triggers-per-rotation` property. See Encoders Config for more details.

Here you need to replace `PIN_A` and `PIN_B` with the appropriate pins that your PCB utilizes for the encoder(s). See shield overlays section in the new shield guide on the appropriate node label and pin number to use for GPIOs.