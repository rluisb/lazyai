# Behaviors Overview

"Behaviors" are bindings that are assigned to and triggered by key positions on keymap layers, sensors (like an encoder) or combos. They describe what happens e.g. when a certain key position is pressed or released, or an encoder triggers a rotation event. They can also be recursively invoked by other behaviors, such as macros.

Below is a summary of pre-defined behavior bindings and user-definable behaviors available in ZMK, with references to documentation pages describing them.

## Key Press Behaviors

| Binding | Behavior | Description |
| --- | --- | --- |
| `&kp` | Key Press | Send keycodes to the connected host when a key is pressed |
| `&mt` | Mod Tap | Sends a different key press depending on whether a key is held or tapped |
| `&kt` | Key Toggle | Toggles the press of a key. If the key is not currently pressed, key toggle will press it, holding it until the key toggle is pressed again or the key is released in some other way. If the key is currently pressed, key toggle will release it |
| `&sk` | Sticky Key | Stays pressed until another key is pressed, then is released. It is often used for modifier keys like shift, which allows typing capital letters without holding it down |
| `&gresc` | Grave Escape | Sends Grave Accent ` keycode if shift or GUI is held, sends Escape keycode otherwise |

## Layer Behaviors

| Binding | Behavior | Description |
| --- | --- | --- |
| `&mo` | Momentary Layer | Activates a layer while held |
| `&mt` | Mod Tap | Sends a modifier when held, or a tap key when released |
| `&lt` | Layer Tap | Activates a layer when held, or sends a tap key when released |
| `&tg` | Layer Toggle | Toggles a layer on or off |
| `&tt` | Layer Tap-Toggle | Toggles a layer if tapped, or activates it if held |
| `&sl` | Layer Slide | Momentarily activates a layer when pressed, returns to previous layer when released |

## Encoding & Sensor Behaviors

| Binding | Behavior | Description |
| --- | --- | --- |
| `&enc` | Encoder | Handles encoder rotation and press events |

## Special Behaviors

| Binding | Behavior | Description |
| --- | --- | --- |
| `&bt` | Bluetooth | Connect/disconnect Bluetooth profiles |
| `&out` | Output Selection | Select USB or Bluetooth output |
| `&ext_power` | External Power | Control external power states |
| `&sleep` | Sleep | Put the keyboard into sleep mode |

## User-Defined Behaviors

| Behavior | Description |
| --- | --- |
| Macros | Allows configuring a list of other behaviors to invoke when the key is pressed and/or released |
| Hold-Tap | Invokes different behaviors depending on key press duration or interrupting keys. |
| Tap Dance | Invokes different behaviors corresponding to how many times a key is pressed |
| Mod-Morph | Invokes different behaviors depending on whether a specified modifier is held during a key press |
| Sensor Rotation | Invokes different behaviors depending on whether a sensor is rotated clockwise or counter-clockwise |