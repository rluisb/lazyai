# Debouncing

To prevent contact bounce (also known as chatter) and noise spikes from causing unwanted key presses, ZMK uses a cycle-based debounce algorithm, with each key debounced independently.

By default the debounce algorithm decides that a key is pressed or released after the input is stable for 5 milliseconds. You can decrease this to improve latency or increase it to improve reliability.

`debounce-scan-period-ms` determines how often the keyboard scans while debouncing. It defaults to 1 ms, but it can be increased to reduce power use. Note that the debounce press/release timers are rounded up to the next multiple of the scan period. For example, if the scan period is 2 ms and debounce timer is 5 ms, key presses will take 6 ms to register instead of 5.

## Eager Debouncing

Eager debouncing means reporting a key change immediately and then ignoring further changes for the debounce time. This eliminates latency but it is not noise-resistant.

If you are having problems with a single key press registering multiple inputs, you can try increasing the debounce press and/or release times to compensate. You should also check for mechanical issues that might be causing the bouncing, such as hot swap sockets that are making poor contact.

## Debounce Configuration

Currently the `zmk,kscan-gpio-matrix` and `zmk,kscan-gpio-direct` drivers supports these options, while `zmk,kscan-gpio-demux` driver does not.