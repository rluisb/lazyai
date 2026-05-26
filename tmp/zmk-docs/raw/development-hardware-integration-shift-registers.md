# Configuring Shift Registers

Shift registers are the recommended method of adding additional GPIO pins to MCUs and boards, when a standard matrix results in an insufficient number of keys. They are recommended because they simultaneously have very low power consumption and are quite cheap. This page serves as a (brief) introduction to shift registers, how to use them in your design, and how to configure ZMK to use them correctly.

> note
> 
> This page assumes that you are using a SIPO shift register with the part number 74HC595. Other shift registers can work as well but this is the most commonly used one.

> tip
> 
> Design Guidelines, Configuration, Enable SPI, Shift Register SPI Device, Using Shift Register Pins In Kscan