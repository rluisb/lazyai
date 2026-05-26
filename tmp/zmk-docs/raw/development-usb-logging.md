# USB Logging

## Overview

If you are developing ZMK on a device that does not have a built in UART for debugging and log/console output, Zephyr can be configured to create a USB CDC ACM device and the direct all `printk`, console output, and log messages to that device instead.

> Battery Life Impact
> 
> Enabling logging increases the power usage of your keyboard, and can have a non-trivial impact to your time on battery. It is recommended to only enable logging when needed, and not leaving it on by default.

## USB Logging Snippet

The `zmk-usb-logging` snippet is used to enable logging.

### Additional Config

Logging can be further configured using Kconfig described in the Zephyr documentation. For instance, setting `CONFIG_LOG_PROCESS_THREAD_STARTUP_DELAY_MS` to a large value such as `8000` might help catch issues that happen near keyboard boot, before you can connect to view the logs.

## Viewing Logs

After flashing the updated ZMK image, the board should expose a USB CDC ACM device that you can connect to and view the logs.

### Linux

On Linux, this should be a device like `/dev/ttyACM0` and you can connect with `minicom` or `tio` as usual.

### Windows

### MacOS

## Enabling Logging on Older Boards

Previously, enabling logging required setting the `CONFIG_ZMK_USB_LOGGING` Kconfig symbol. If for whatever reason a custom board definition does not support the new `zmk-usb-logging` snippet, you can try setting this symbol at the keyboard level.