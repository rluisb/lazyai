# Customizing ZMK/zmk-config folders | ZMK Firmware

Source: https://zmk.dev/docs/customization

---

After verifying you can successfully flash the default firmware, you will probably want to begin customizing your keymap and other keyboard options. In the initial setup tutorial, you created a Github repository called `zmk-config`. This repository is a discrete filesystem which works with the main `zmk` firmware repository to build your desired firmware. The main advantage of a discrete configuration folder is ensuring that the working components of ZMK are kept separate from your personal keyboard settings, reducing the amount of file manipulation in the configuration process. This makes flashing ZMK to your keyboard much easier, especially because you don't need to keep an up-to-date copy of zmk on your computer at all times.

## Flashing Your Changes

For normal keyboards, follow the same flashing instructions as before to flash your updated firmware.

For split keyboards, only the central (left) side will need to be reflashed if you are just updating your keymap. More troubleshooting information for split keyboards can be found here.

## Building Additional Keyboards

You can build additional keyboards with GitHub actions by appending them to `build.yaml` in your `zmk-config` folder. For instance assume that we have set up a Corne shield with nice!nano during initial setup and we want to add a Lily58 shield with nice!nano v2. The following is an example `build.yaml` file that would accomplish that.