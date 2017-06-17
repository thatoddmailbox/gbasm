# gbasm
Assembler for Gameboy games. Currently missing a few instructions, and only supports Gameboy Color-exclusive games with no mapping hardware and no cartridge RAM.

## FAQ
### why
because RGBDS is crappy especially for multi-platform use
### but this is even more crappy
that's not important
### how do I use this
1. Make a folder
2. Make a file called `info.toml`, put this in it
```
Name = "COOL GAME"
```
(name is limited to 15 characters, and this file will probably need more settings at some point)
3. Make a file called `main.s`, put assembly code in there.