# ansiart2utf8

[![GoDoc](https://godoc.org/github.com/BourgeoisBear/ansiart2utf8?status.png)](http://godoc.org/github.com/BourgeoisBear/ansiart2utf8)

Processes legacy BBS-style ANSI art (ACiDDraw, PabloDraw, etc.) to UTF-8.<br/>
Escape codes and line endings are processed for terminal friendliness.

## BEFORE
![Before ansiart2utf8 processing][imgBefore]

## AFTER
![After ansiart2utf8 processing][imgAfter]

[imgBefore]: docs/before.png "ANSI in Terminal Before Processing"
[imgAfter]: docs/after.png "ANSI in Terminal After Processing"

## INSTALLATION

1. Install the latest Go compiler from https://golang.org/dl/
2. Install the program:

```sh
go install github.com/BourgeoisBear/ansiart2utf8/ansiart2utf8@latest
```

## USAGE

```sh

ansiart2utf8
  Converts ANSI art to UTF-8 encoding, expands cursor forward ESC sequences
  into spaces, wraps/resets at a specified line width, sends result to STDOUT.

  Leave the [FILE] parameter empty to read from STDIN.

USAGE: ansiart2utf8 [OPTION]... [FILE]...

OPTIONS
  -bytes uint
        MAXIMUM OUTPUT BYTES PER-ROW (0 = NO LIMIT)
  -debug
        DEBUG MODE: line numbering + pipe @ \n
  -w uint
        LINE WRAP WIDTH (default 80)
  -x    ANSI TO XTERM-256 COLOR SUBSTITUTION
          (to overcome strange terminal color scheme palettes)

```

## NOTES

**To see the result, make sure that your terminal font provides glyphs for the old CP437 box drawing characters.**

Here are a few fonts that will do:

- [Consolas](https://en.wikipedia.org/wiki/Consolas)
- [Courier New](https://www.microsoft.com/typography/fonts/family.aspx?FID=10)
- [DejaVu Sans Mono](https://github.com/dejavu-fonts/dejavu-fonts)
- [Envy Code R](https://damieng.com/blog/2008/05/26/envy-code-r-preview-7-coding-font-released)
- [Iosevka](https://be5invis.github.io/Iosevka/)

### Seeing Code Page 437 in Vim

`:e ++enc=cp437`

### Resources

- [PabloDraw](http://picoe.ca/products/pablodraw/), an ANSI drawing program for Windows
- [ACiDDraw](http://www.acid.org/apps/apps.html), an ANSI drawing program for DOS
- Lots of ANSI art to be seen here:<br/>http://blocktronics.org/artpacks/
- Ultimate guide to pimping your terminal:<br/>http://mewbies.com/acute_terminal_fun_table_of_contents.htm
- A very clear mapping of code page 437 characters to Unicode at Wikipedia:<br/>
  https://en.wikipedia.org/wiki/Code_page_437#Characters
- Helpful references on ANSI escape codes:<br/>
  https://en.wikipedia.org/wiki/ANSI_escape_code<br/>
  https://www.gnu.org/software/screen/manual/html_node/Control-Sequences.html

### Media

- http://artscene.textfiles.com/ansi/
- https://www.ansilove.org/bbs.html
- http://bbs.ninja/
- http://ascii-table.com/ansi-escape-sequences.php
- http://bluesock.org/~willg/dev/ansi.html
- https://github.com/k0kubun/go-ansi
