ansiart2utf8
============
VERSION 0.1 BETA
----------------
Processes legacy BBS-style ANSI art (ACiDDraw, PabloDraw, etc.) to UTF-8.<br/>
Escape codes and line endings are processed for terminal friendliness.

```
USAGE: ansiart2utf8 [OPTION]...

OPTIONS
  -f string
    	INPUT FILENAME, OR "-" FOR STDIN (default "-")
  -w uint
    	LINE WIDTH (default 80)
```

BEFORE & AFTER
--------------
**BEFORE**
![Before ansiart2utf8 processing][imgBefore]

**AFTER**
![After ansiart2utf8 processing][imgAfter]

[imgBefore]: ansiart2utf8-before.gif "ANSI in Terminal Before Processing"
[imgAfter]: ansiart2utf8-after.gif "ANSI in Terminal After Processing"

NOTES
-----
To build:

1. Install the latest Go compiler from https://golang.org/dl/
2. Change to project folder: `cd ./ansiart2utf8`
3. Build executable: `go build ./ansiart2utf8.go`

**To see the result, make sure that your terminal font provides glyphs for the old CP437 box drawing characters.**<br/>Here are a few fonts that will do:

- [DejaVu Sans Mono](https://github.com/dejavu-fonts/dejavu-fonts)
- [Envy Code R](https://damieng.com/blog/2008/05/26/envy-code-r-preview-7-coding-font-released)
- [Courier New](https://www.microsoft.com/typography/fonts/family.aspx?FID=10)
- [Consolas](https://en.wikipedia.org/wiki/Consolas)


RESOURCES
---------
- [PabloDraw](http://picoe.ca/products/pablodraw/), an ANSI drawing program for Windows
- [ACiDDraw](http://www.acid.org/apps/apps.html), an ANSI drawing program for DOS
- Lots of ANSI art to be seen here:<br/>http://blocktronics.org/artpacks/
- Ultimate guide to pimping your terminal:<br/>http://mewbies.com/acute_terminal_fun_table_of_contents.htm
- A very clear mapping of code page 437 characters to Unicode at Wikipedia:<br/>
  https://en.wikipedia.org/wiki/Code_page_437#Characters
- Helpful references on ANSI escape codes:<br/>
  https://en.wikipedia.org/wiki/ANSI_escape_code<br/>
  https://www.gnu.org/software/screen/manual/html_node/Control-Sequences.html

http://artscene.textfiles.com/ansi/
https://www.ansilove.org/bbs.html
http://bbs.ninja/
http://ascii-table.com/ansi-escape-sequences.php
http://bluesock.org/~willg/dev/ansi.html
https://github.com/k0kubun/go-ansi