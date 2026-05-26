# ZMK Events

ZMK makes use of events to decouple individual components such as behaviors and peripherals from the core functionality. For this purpose, ZMK has implemented its own event manager. This page is a (brief) overview of the functionality and methods exposed by the event manager, documenting its API. Its purpose is to aid module developers and contributors, such as for the development of new behaviors or new features. There is no value in reading this page as an end-user.

The priority of the listeners is determined by the order in which the linker links the files. Within ZMK, this is the order of the corresponding files in `CMakeLists.txt`. External modules targeting `app` are linked prior to any files within ZMK itself, making them the highest priority. It is thus the module maintainer's responsibility to both ensure that their module does not cause issues by being first in the listener queue. For example, hold-tap is the first listener to `position_state_changed`, and may behave inconsistently if a behavior defined in a module listens to `position_state_changed` and invokes a `hold-tap` (e.g. by calling `zmk_behavior_invoke_event` with a `hold-tap` as the binding).

Once you have a listener set up, you can subscribe to individual events by calling the `ZMK_SUBSCRIPTION` macro:

```
ZMK_SUBSCRIPTION(combo, zmk_keycode_state_changed);
ZMK_SUBSCRIPTION(combo,  zmk_keycode_state_changed);
```

The first parameter is the name of the listener created with `ZMK_LISTENER`, while the second is the name of the struct that defines the event's data, which was declared in the corresponding header file. By convention the header file for an event will be named `specific_thing_happened`, with the struct named `zmk_specific_thing_happened`.

Of course, you will also need to import the corresponding event header at the top of your file.

### Listener Callback