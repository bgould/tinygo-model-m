Replacement controller for Model M keyboard with firmware written in TinyGo
=================================================

This is an experimental project that I am using to familiarize myself with idioms and techniques for creating embedded firmware using [TinyGo][tinygo].

The architecture of the firmware code is a loose port of the excellent [TMK Keyboard][tmk] firmware libary from C to TinyGo.

Hardware
--------

To build the controller I de-soldered the connectors from the original controller and added the following parts:
 * (1) [Adafruit Feather M0 Express][feather-m0]
 * (1) [Bluefruit LE SPI Friend][spifriend]
 * (2) [MCP23008 I/O Port Expanders][mcp23008]
 * (2) 4.47k resistors

The setup looks like this:

<image src="docs/matrix_and_circuit_1024x768.png">

<image src="docs/circuit.jpg">

Firmware
--------

To build and install the firmware, first install TinyGo as per the website and make sure you've setup up the LLVM toolchain to be able to compile code for ARM Cortex-M according to the instructions.

Use `make` to compile a UF2 file and `make flash` to flash it onto the Feather board (assuming it is on `/dev/ttyACM0`, you may need to adjust the Makefile to suit your needs).  If you do not already have a TinyGo application running on the Feather board, you may need to press the reset button twice in order to put it into bootloader mode the first time.

Note: if your Bluefruit LE SPI Friend device is not brand new you should power up the device and hold the DFU pin low for 5 seconds and then release it to perform a factory reset, or use one of Adafruit's example Arduino sketches to do the same using the documented AT command.

Usage
-----

Once the firmware starts up, the Bluefruit device will need to be paired.  Perform a bluetooth scan with the device you want to pair and select "TinyGo Model M Keyboard".  Once paired, the blue light on the Bluefruit device will illuminate and stay on.  You can then start using the keyboard with the paired device.

The keymap used in the firmware is the "ANSI 101" layout that you'll find on most vintage US versions of the Model M keyboard.  The keys can be remapped by changing the <a href="pkg/modelm/keymap.go">pkg/modelm/keymap.go</a> file and recompiling.  A list of available keycodes can be found in <a href="pkg/keyboard/keycodes/keycodes.go">keycodes.go</a> (not all are supported yet, see Next Steps below).

If the firmware is compiled with the debug variable set to true in main.go, you can connect to the serial port of the Feather device to see debugging information as you type, for exampe:

    screen /dev/ttyACM0 115200

Make sure you set the baud rate explicitly other is seems `screen` will trigger a reset to the bootloader on the Feather board.

For development, you might find it useful to set the following build tags in your editor for better Go completion:

    :GoBuildTags tinygo arm atsamd21g18 atsamd21 sam atsamd21g18a feather_m0

Next steps
----------

I plan to take this a little further before perhaps refactoring and spinning off a separate project that is a more generic port (friendly fork) of the TMK firmware library.  I would definitely like to get layer switching working with multiple keymaps, as well as mouse keys and media keys.

In addition, to the keyboard framework code, this project probably has a couple of driver libraries that could be extracted and built upon, specifically the code for the MCP23008 chips and the code for interfacing with the Bluefruit SPI Friend device.

Contributions & License
-----------------------

Pull requests are welcome.

For now license is GPLv2 or later, same as TMK keyboard firmware since I'm not sure if this is considered a derivative work or not, so better safe than sorry.  At this point this is just a prototype/experiment anyhow.

[tinygo]: https://tinygo.org/
[tmk]: https://github.com/tmk/tmk_keyboard

[feather-m0]: https://www.adafruit.com/product/3403
[spifriend]: https://www.adafruit.com/product/2633
[mcp23008]: https://www.adafruit.com/product/593