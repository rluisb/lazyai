# ZMK Studio RPC Protocol

> danger
> 
> Before reading this section, it is vital that you read through our clean room policy.

## Overview

The ZMK Studio UI communicates with ZMK devices using a custom RPC protocol developed to be robust and reliable, while remaining simple and easy to extend with future enhancements.

The protocol consists of protocol buffer messages which are encoded/decoded using message framing, and then transmitted using an underlying transport. Two transports are currently implemented: a BLE transport using a custom GATT service and a serial port transport, which usually is used with CDC ADM devices over USB.

## Protobuf Messages

The messages for ZMK Studio are defined in a dedicated zmk-studio-messages repository. Fundamentally, the `Request` message is used to send any requests from the ZMK Studio client to the ZMK device, and the `Response` messages are sent from the ZMK device to the Studio client.

Responses can either be `RequestResponses` that are sent in response to an incoming `Request` or a `Notification` which is sent at any point from the ZMK device to the ZMK Studio client to inform the client about state changes on the device, e.g. that the device is unlocked.

## Message Framing

ZMK Studio uses a simple framing protocol to easily identify the start and end of a given message, with basic escaping to allow for unrestricted content.

### Example Encoding (Simple)

### Example Encoding (Escaping)

## Transports

### USB (Serial)

The USB transport is actually a basic serial/UART transport, that happens to use the CDC/ACM USB class for a serial connection. Framed messages are sent between ZMK Studio client and ZMK device using simple UART transmission.

### Bluetooth (GATT)

The bluetooth transport uses a custom GATT service to transmit/receive. The service has UUID `00000000-0196-6107-c967-c5cfb1c2482a` and has exactly one characteristic with UUID `00000001-0196-6107-c967-c5cfb1c2482a`. The characteristic accepts writes of framed client messages, and will use GATT Indications to send framed messages to the client.