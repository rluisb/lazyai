# Supported Hardware | ZMK Firmware

Source: https://zmk.dev/docs/hardware

---

### Seeed XIAO Interconnect

The Seeed Studio XIAO is a popular smaller format micro-controller, that has gained popularity as an alternative to the SparkFun Pro Micro. Since its creation, several pin compatible controllers, such as the Seeed Studio XIAO nRF52840 (also known as XIAO BLE), Adafruit QT Py and Adafruit QT Py RP2040, have become available.

#### Boards

- Adafruit QT Py RP2040 (Board: `adafruit_qt_py_rp2040//zmk`)
- Seeed Studio XIAO SAMD21 (Board: `seeeduino_xiao//zmk`)
- Seeed Studio XIAO nRF52840 (Board: `xiao_ble//zmk`)
- Seeed Studio XIAO RP2040 (Board: `xiao_rp2040//zmk`)

#### Shields

- Hummingbird (Shield: `hummingbird`)
- TesterXiao (Shield: `tester_xiao`)

### Arduino Uno Rev3 Interconnect

The Arduino Uno Rev3 is a board who's popularity lead to countless shields being developed for it. Note: ZMK doesn't support boards with AVR 8-bit processors, such as the ATmega32U4, because Zephyr™ only supports 32-bit and 64-bit platforms.

#### Boards

- Nordic nRF52840 DK (Board: `nrf52840dk/nrf52840/zmk`)
- Nordic nRF5340 DK (Board: `nrf5340dk/nrf5340/cpuapp`)

#### Shields

- ZMK Uno (Shield: `zmk_uno`)

## Onboard Controller Keyboards

Keyboards with onboard controllers are single PCBs that contain all the components of a keyboard, including the controller chip, switch footprints, etc.

- Advantage 360 Pro (Boards: `adv360pro_left//zmk`, `adv360pro_right//zmk`)
- BDN9 (Rev2) (Board: `bdn9//zmk`)
- BT60 V1 Hotswap (Board: `bt60_hs//zmk`)
- BT60 V2 (Board: `bt60//zmk`)
- BT65 (Board: `bt65//zmk`)
- BT75 V1 (Board: `bt75//zmk`)
- Corneish Zen (Boards: `corneish_zen_left//zmk`, `corneish_zen_right//zmk`)
- Ferris 0.2 (Board: `ferris//zmk`)
- Glove80 (Boards: `glove80_lh`, `glove80_rh`)
- KBDfans Tofu65 2.0 (Board: `tofu65//zmk`)
- nice!60 (Board: `nice60//zmk`)
- Planck (Rev6) (Board: `planck//zmk`)
- Preonic Rev3 (Board: `preonic//zmk`)
- S40NC (Board: `s40nc//zmk`)

## Composite Keyboards