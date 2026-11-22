[![Go Reference](https://pkg.go.dev/badge/github.com/bengarrett/bbs.svg)](https://pkg.go.dev/github.com/bengarrett/bbs)
![Coverage](https://img.shields.io/badge/Coverage-93.0%25-brightgreen)

# Package bbs

Package bbs is a [Go module](https://go.dev/), terminal utility, and WebAssembly package that detects and converts legacy [Bulletin Board System](https://en.wikipedia.org/wiki/Bulletin_board_system) (BBS) color-coded textfiles into ANSI escape sequences, HTML documents, PNG images, or stripped plain text.

[Check out the live web app.](https://bengarrett.github.io/bbs/)

BBSes were popular in the 1980s and 1990s and allowed computer users to chat,
message, and share files over the landline telephone network. The commercialization
and ease of access to the Internet eventually replaced BBSes, as did the worldwide-web. These centralized systems, termed _boards_, used a text-based interface, and their
owners often applied colorization, text themes, and art to differentiate themselves.

While in the 1990s, [ANSI control codes](https://en.wikipedia.org/wiki/ANSI_escape_code) were in everyday use on the [PC/MS-DOS](https://en.wikipedia.org/wiki/MS-DOS), the
standard comes from mainframe equipment. Home microcomputers often had difficulty
interpreting it. So, BBS developers created their own, more straightforward methods
to colorize and theme the text output to solve this.

*Please note that many microcomputer, PC and MS-DOS based boards used ANSI control codes for colorizations that this library does not convert.

## Terminal utility

[Download the latest release](https://github.com/bengarrett/bbs/releases/latest)

```bash
# Convert to ANSI
bbs display.pcb output.ans

# Convert to a HTML file with embedded VGA font & CSS
bbs -f html -css display.pcb display.html

# Render as a PNG image
bbs -f png display.pcb

# Strip BBS color codes to plain text
bbs -f strip display.pcb

# Detect the BBS color format
bbs -check display.pcb

# Disable 80-column wrapping
bbs -unlock display.pcb
```

## Go package quick usage

[Go Package documentation and examples.](https://pkg.go.dev/github.com/bengarrett/bbs)

```go
// open the text file
file, err := os.Open("pcboard.txt")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// render directly to a PNG image
var pngBuf bytes.Buffer
kind, err := bbs.PNG(&pngBuf, file)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Detected:", kind.Name())

const hello = "@X1EHello world"
// convert to ANSI escape sequences
var ansiBuf bytes.Buffer
if _, err := bbs.ToANSI(&ansiBuf, strings.NewReader(hello)); err != nil {
    log.Fatal(err)
}
fmt.Println(ansiBuf.String())

// convert to HTML
var htmlBuf bytes.Buffer
if _, err := bbs.HTML(&htmlBuf, strings.NewReader(hello)); err != nil {
    log.Fatal(err)
}
fmt.Println(htmlBuf.String())
```

## Known codes

### PCBoard

One of the most well-known applications for hosting a PC/MS-DOS BBS, PCBoard
pioneered the `file_id.diz` file descriptor, and being endlessly expandable
through software plugins known as PPEs. It developed the popular **@X** color code and
**@** control syntax.

### Celerity

Another PC/MS-DOS application was very popular with the hacking, phreaking,
and pirate communities in the early 1990s. It introduced a unique **|** pipe code
syntax in late 1991 that revised the code syntax in version 2 of the software.

### Renegade

A PC/MS-DOS application that was a derivative of the source code of Telegard BBS.
Surprisingly, there was a new release of this software in 2021. Renegade had two
methods to implement color, and this library uses the Pipe Bar Color Codes.

### Telegard

A PC/MS-DOS application became famous due to a source code leak or release by
one of its authors in an era when most developers were still highly
secretive with their code. The source is in use in several other BBS applications.

### WWIV

A mainstay in the PC/MS-DOS BBS scene of the 1980s and early 1990s, the software became well-known for releasing its source code to registered users. It allowed owners to expand the code to incorporate additional software, such as games or utilities, and port it to other platforms. The source is now Open Source and is still updated. Confusingly, WWIV has three methods of colorizing text: 10 **|** pipe colors, two-digit pipe colors, and its original **♥** Heart Codes.

### Wildcat

WILDCAT! was a popular, propriety PC/MS-DOS application from the late 1980s that later migrated to Windows. It was one of the few BBS applications sold at retail in a physical box. It extensively used **@** color codes throughout later revisions of the software.
