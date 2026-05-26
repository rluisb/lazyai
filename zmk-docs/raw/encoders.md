# Encoder Configuration

## EC11 Encoders

### Kconfig

Definition file: zmk/app/module/drivers/sensor/ec11/Kconfig

| Config | Type | Description | Default |
| --- | --- | --- | --- |
| `CONFIG_EC11` | bool | Enable EC11 encoders | n |
| `CONFIG_EC11_THREAD_PRIORITY` | int | Priority of the encoder thread | 10 |
| `CONFIG_EC11_THREAD_STACK_SIZE` | int | Stack size of the encoder thread | 1024 |

If `CONFIG_EC11` is enabled, exactly one of the following options must be set to `y`:

| Config | Type | Description |
| --- | --- | --- |
| `CONFIG_EC11_TRIGGER_NONE` | bool | No trigger (encoders are disabled) |
| `CONFIG_EC11_TRIGGER_GLOBAL_THREAD` | bool | Process encoder interrupts on the global thread |
| `CONFIG_EC11_TRIGGER_OWN_THREAD` | bool | Process encoder interrupts on their own thread |

### Devicetree

Applies to: `compatible = "alps,ec11"`

Definition file: zmk/app/module/dts/bindings/sensor/alps,ec11.yaml

| Property | Type | Description | Default |
| --- | --- | --- | --- |
| `a-gpios` | GPIO array | GPIO connected to the encoder's A pin |  |
| `b-gpios` | GPIO array | GPIO connected to the encoder's B pin |  |
| `steps` | int | Number of encoder pulses per complete rotation |  |