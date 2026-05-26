# Getting Started

We recommend reading through the setup process before following it step by step, to ensure that you are happy with installing the required dependencies.

## Environment Setup

There are two ways to set up the ZMK development environment:

1. **Docker or Podman**: A self-contained development environment. It uses the same Docker image which is used by the GitHub action for local development. Beyond the benefits of dev/prod parity, this approach may be easier to set up for some operating systems. No toolchain or dependencies are necessary when using a container; the image has the toolchain installed and set up to use.

2. **Native**: This uses your operating system directly. Usually runs slightly faster than the container approach, and can be preferable for users who already have the dependencies on their system.