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
