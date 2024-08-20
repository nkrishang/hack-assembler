# Hack Assembler in Go

This repository contains an assembler for the Hack machine language, as specified in [Nand2Tetris](https://www.nand2tetris.org/). The assembler is implemented in Go and converts Hack Assembly Language (`.asm`) files into machine code (`.hack`).

1. Clone this repository:

```bash
   git clone https://github.com/nkrishang/hack-assembler.git
   cd hack-assembler
```

2. Run the assembler:

```bash
go run main.go path/to/your/file.asm
```

This will generate a corresponding .hack file in an output directory, containing the machine code that can be executed on the Hack computer.
