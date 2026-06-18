# RGB Underglow

For example: the `kyria` shield has a `boards/nice_nano_nrf52840_zmk.overlay` and a `boards/nrfmicro_nrf52840_zmk_1_3_0.overlay`, which configure a WS2812 LED strip for the `nice_nano/nrf52840/zmk` and `nrfmicro@1.3.0/nrf52840/zmk` boards respectively.

### nRF52-Based Boards

Using an SPI-based LED strip driver on the `&spi3` interface is the simplest option for nRF52-based boards. If possible, avoid using pins which are limited to low-frequency I/O for this purpose. The resulting interference may result in poor wireless performance.