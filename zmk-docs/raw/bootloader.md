# Bootloader Integration Configuration

## Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_BOOTMODE_BOOTLOADER_MAGIC_VALUE` | hex | The magic value to place into retained memory when the bootloader boot mode is set | none |
| `CONFIG_ZMK_BOOTMODE_MAGIC_VALUE_BOOTLOADER_TYPE_TINYUF2` | bool | Default the bootloader magic value for tinyuf2 bootloader | false |
| `CONFIG_ZMK_BOOTMODE_MAGIC_VALUE_BOOTLOADER_TYPE_ADAFRUIT_BOSSA` | bool | Default the bootloader magic value for Adafruit BOSSA (SAMD21) bootloader | false |
| `CONFIG_ZMK_BOOTMODE_MAGIC_VALUE_BOOTLOADER_TYPE_ADAFRUIT_NRF52` | bool | Default the bootloader magic value for Adafruit nRF52 bootloader | false |
| `CONFIG_ZMK_DBL_TAP_BOOTLOADER` | bool | Enable the double-tap to enter bootloader functionality | y if STM32 or RP2040/RP2350 |
| `CONFIG_ZMK_DBL_TAP_BOOTLOADER_TIMEOUT_MS` | int | Duration (in ms) to wait for a second reset to enter the bootloader | 500 |

## STM32 nBOOT_SEL Option Byte Setup

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_ZMK_BOOT_STM32_ENFORCE_NBOOT_SEL` | bool | Ensure the `nBOOT_SEL` bit is not set | y if STM32CO or STM32G0 |