# gbasm
Assembler for Gameboy games. Currently missing a few instructions, and only supports Gameboy Color-exclusive games with no mapping hardware and no cartridge RAM.

## FAQ
### why
because RGBDS is crappy especially for multi-platform use
### but this is even more crappy
that's not important
### I don't like LDH/LDI/LDD instead of just LD/the usage of [] instead of ()/using 0x instead of $ for hexadecimal/something else
too bad
### does this support a normal Z80
not really. you might be able to get it to work, but none of the Z80-only instructions are supported, and the LR35902-only instructions are, so you should probably use an actual Z80 assembler
### how do I use this
1. Make a folder
2. Make a file called `info.toml`, put this in it: (name is limited to 15 characters, and this file will probably need more settings at some point)
```
Name = "COOL GAME"
```
3. Make a file called `main.s`, put assembly code in there.
4. Run `gbasm` in that folder.
5. Do stuff with the `out.gb` file it creates.

## Known issues
* the expression parser likes to assume parentheses and do weird things. for example, `2 - 3 + 4` gets interpreted as `2 - (3 + 4)`, which is probably not what you want
* no MBCs are supported, and rom sizes are assumed to be 32 KiB
* all numbers are assumed to be unsigned -- putting in `-1` will give you an error
* you can cause weird unhelpful errors to occur with the dot instructions if you mess with their expected parameters

## Assembler instructions
things that aren't actual LR35902 instructions but that do useful things
* `ascii "<string>"`
  inserts that string, encoded using ASCII, into the output
* `asciz "<string>"`
  same as `ascii` but terminates the string with a null byte
* `db <byte>`
  inserts that byte into the output
* `dw <word>`
  inserts that word into the output
* `.def <something> <value>`
  defines `<something>` as equal to `<value>`. useful for registers and things like that
* `.org <address>`
  sets the origin from that point on to the given address
* `.incasm "<file>.s"`
  includes everything from that assembly file

## Things that are different from other assemblers
* the checksums are automatically calculated, you don't need some other program to fix them for you
* the `0b` prefix can be used to make a binary number (for example, `0b10101010` == `170`)
* the `0x` prefix can be used to make a hexadecimal number (for example, `0x2A` == `42`)
* the `%` and `$` for binary and hexadecimal numbers are not currently supported because I'm lazy