# Module Creation

ZMK modules are the recommended method of adding content to ZMK, whenever it is possible. This page will guide you through creating a module for ZMK. Distinction is made between modules used for different purposes:

- Modules containing one or more keyboard-related definitions (boards, shields, interconnects, etc.)
- Modules containing behaviors & features
- Modules containing drivers
- Modules containing other features, such as visual effects

See also Zephyr's page on modules.

> tip
> 
> For open source hardware designs, it can be convenient to use Git submodules to have the ZMK module also be a Git submodule of the repository hosting the hardware design.

## Module Setup

ZMK has a template to make creating a module easier. Navigate to the ZMK module template repository and select "Use this template" followed by "Create a new repository" in the top right.

The below table reminds of the purpose of each of these files and folders:

| File or Folder | Description |
|---|---|
| `boards/` | Folder containing definitions for boards, shields and interconnects |
| `dts/` | Folder containing devicetree bindings and includes with devicetree nodes (.dtsi) |
| `CMakeLists.txt` | CMake configuration to specify source files to build |
| `Kconfig` | Kconfig file for the module |
| `include/` | Folder for C header files |
| `src/` | Folder for C source files |
| `snippets/` | Folder for snippets |