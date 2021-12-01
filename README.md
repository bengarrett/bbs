# Package bbs

Package bbs is a [Go module]() that interacts with legacy textfiles encoded with
[Bulletin Board Systems]() (BBS) color codes to reconstruct them into HTML documents.

BBSes were popular in the 1980s and 1990s, and allowed computer users to chat,
message, and share files over the landline telephone network. The commercialization
and ease of access to the Internet eventually replaced BBSes, as did the world-wide-web.

These centralized systems, termed boards, used a text-based interface, and their
owners often applied colorization, text themes, and art to differentiate themselves.

While in the 1990s, [ANSI control codes]() were in everyday use on the [PC/MS-DOS](), the
standard comes from mainframe equipment. Home microcomputers often had difficulty
interpreting it. So BBS developers created their own, more straightforward methods
to colorize and theme the text output to solve this.

*Please note, while many PC/MS-DOS boards used ANSI control codes for colorizations,
this library does not support the standard.

## Quick usage

[Go Package with docs and examples.](https://pkg.go.dev/github.com/bengarrett/bbs)

```go
// open text file
file, err := os.Open("textfile.pcb")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// transform the MS-DOS legacy text to Unicode
decoder := charmap.CodePage437.NewDecoder()
reader := transform.NewReader(file, decoder)

// create the HTML equivalent of BBS color codes
var htm bytes.Buffer
match, err := bbs.HTML(&htm, reader)
if err != nil {
    log.Fatal(err)
}

// fetch CSS
var css bytes.Buffer
if err := match.CSS(&css); err != nil {
    log.Fatal(err)
}

// print partial html and css
fmt.Println(htm.String())
fmt.Println(css.String())
```

## Known codes

### PCBoard

One of the most well-known applications for hosting a PC/MS-DOS BBS, PCBoard
pioneered the file_id.diz file descriptor, as well as being endlessly expandable
through software plugins known as PPEs. It developed the popular @X color code and
@ control syntax.

### Celerity

Another PC/MS-DOS application that was very popular with the hacking, phreaking,
and pirate communities in the early 1990s. It introduced a unique | pipe code
syntax in late 1991 that revised the code syntax in version 2 of the software.

### Renegade

A PC/MS-DOS application that was a derivative of the source code of Telegard BBS.
Surprisingly there was a new release of this software in 2021. Renegade had two
methods to implement color, and this library uses the Pipe Bar Color Codes.

### Telegard

A PC/MS-DOS application became famous due to a source code leak or release by
one of its authors back in an era when most developers were still highly
secretive with their code. The source is incorporated into several other projects.

### WVIV

A mainstay in the PC/MS-DOS BBS scene of the 1980s and early 1990s, it became well
known for releasing its source code to registered users. It allowed them to expand
the code to incorporate additional software such as games or utilities and port it
to other platforms. The source is now Open Source and is still updated.
Confusingly WWIV has three methods of colorizing text, 10 Pipe colors, two-digit
pipe colors, and its original Heart Codes.

### Wildcat

WILDCAT! was a popular, propriety PC/MS-DOS application from the late 1980s that
later migrated to Windows. It was one of the few BBS applications that sold at
retail in a physical box. It extensively used @ color codes throughout later
revisions of its software.