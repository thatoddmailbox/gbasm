# gbasm
Assembler for Gameboy games. Currently missing a few instructions, and only supports Gameboy Color-exclusive games with no mapping hardware and no cartridge RAM.

## FAQ
### why
because RGBDS is crappy especially for multi-platform use
### but this is even more crappy
that's not important
### I don't like LDH/LDI/LDD instead of just LD/the usage of [] instead of ()/something else
too bad
### how do I use this
1. Make a folder
2. Make a file called `info.toml`, put this in it: (name is limited to 15 characters, and this file will probably need more settings at some point)
```
Name = "COOL GAME"
```
3. Make a file called `main.s`, put assembly code in there.

## Known issues
* the expression parser likes to assume parentheses and do weird things. for example, `2 - 3 + 4` gets interpreted as `2 - (3 + 4)`, which is probably not what you want