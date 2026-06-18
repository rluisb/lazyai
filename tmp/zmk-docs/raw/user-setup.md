# Installing ZMK | ZMK Firmware

Source: https://zmk.dev/docs/user-setup

---

## Summary

The following steps can be taken to obtain an installable firmware image for your device, without the need to install any compiler or specialized toolchain. This is possible by leveraging GitHub Actions to build your firmware for you in the cloud, which you can then download and flash to your device.

Following the steps in this guide, you will:

- `build.yaml`
- `.keymap`
- `.conf`

## Prerequisites

You will need to install a few tools before you can get started.

Many instructions in this guide use commands that need to be run in a terminal application. On most operating systems, the program is simply named "Terminal".

On Windows, get Windows Terminal from the Microsoft Store if it isn't already installed.

### Install Git

Open a terminal and run the following command. If Git is already installed, it will print out a version number.

`git --version`

If it prints an error instead, install Git from . Close and reopen your terminal and run the `git --version` command again to check if it installed correctly.

### Set Up a GitHub Account

You will need a GitHub account. Create an account if you don't have one yet.

### Commit and Push to GitHub

After you've changed your keymap to your liking, you need to commit and push your changes.

Run the following commands to go to the repo directory, mark all your changed files to be added to the commit, create the commit, and push it to GitHub.

`zmk cd  
git add .  
git commit -m "Initial commit"  
git push`

The first time you run this command, it will try to identify which text editors you have installed and ask you which one to use. If you want to change this later, change the `core.editor` setting.