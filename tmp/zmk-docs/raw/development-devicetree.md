# Devicetree Overview

The properties the node can have are listed under `properties`. Some additional properties are imported from zero_param.yaml. Bindings files are the authority on node properties, with our documentation of said properties sometimes omitting things like the `#binding-cells` property (imported from the previously mentioned file, describing the number of parameters that the behavior accepts). A full description of the bindings file syntax can be found in Zephyr's documentation.

Note that binding files can also specify properties for children, like the `zmk,keymap.yaml` bindings file specifying properties for layers in the keymap.

#### Status