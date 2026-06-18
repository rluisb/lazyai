# FAQs | ZMK Firmware

Source: https://zmk.dev/docs/faq

---

### Can I contribute?

Of course! Please use the developer documentation to get started!

### I have an idea! What should I do?

Please join us on Discord and discuss it with us!

### I want to add a new keyboard! What should I do?

The exact process for the management of all the possible hardware is still being finalized, but any developer looking to contribute new keyboard definitions should chat with us on Discord to get started.

### Does ZMK have a Code of Conduct?

Yes, it does have a Code of Conduct! Please give it a read!

### What does "ZMK" mean?

ZMK was originally coined as a quasi-acronym of "Zephyr Mechanical Keyboard" and also taking inspiration from the amazing keyboard firmware projects, TMK and QMK.

### Is ZMK related to TMK or QMK?

### Does ZMK support wired split?

Currently, ZMK only supports wireless split keyboards. Experimental wired split support for some specific hardware designs is available for advanced users to test.

### How is the latency?

The latency of ZMK is comparable to other firmware offerings. ZMK is equipped with a variety of scanning methods and debounce algorithms that can affect the final measured latency. This video shows a latency comparison of ZMK and other keyboard firmwares.

### Any chance for 2.4GHz dongle implementation?

### How do I get started?

ZMK is still in its infancy, so there's a learning curve involved. But if you'd like to try it out, please check out the development documentation and the other FAQs. Please keep in mind that the project team is still small, so our support capability is limited whilst we focus on development. But we'll try our best! Interested developers are also very welcome to contribute!

### What are boards and shields? Why not just "keyboard"?

ZMK uses the Zephyr concepts of "boards" and "shields" to refer to different parts of a keyboard build, that in turn get combined during a firmware build.
This provides the modularity to be able to use composite keyboards with different compatible controllers.
Please see the explainer on boards & shields for more details.