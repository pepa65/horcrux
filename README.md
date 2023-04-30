# horcrux v0.3.2
**Split file into encrypted horcruxes, mergeable without key**

* Repo: https://github.com/pepa65/horcrux
* Contact: pepa65@passchier.net
* After https://github.com/jesseduffield/horcrux
* Install: `wget -qO- gobinaries.com/pepa65/horcrux |sh`

## Function
The program `horcrux` can split a file into a predefined number of horcruxes,
or merge a predefinable sufficient number of constituent horcruxes in a
directory back into the original file.

### Split
To split up a file into horcruxes, call `horcrux` with the filename and
optionally flags `-n`/`--num` with the number of desired horcruxes or
`-m`/`--min` followed by the minimum number of horcruxes needed to merge back
into the original file. Example:

`horcrux -n 5 -m 3 secret.txt`

The resulting horcruxes can be renamed, dispersed and later be used to put the
original file back together if only the minimum number is met.

### Merge
To merge horcruxes back into the original file, call `horcrux` in the
directory containing the horcruxes, or pass that directory as an argument:

`horcrux directory/with/horcruxes`

Only files ending in `.horcrux` will be taken into account. There should not be
any horcruxes in the directory that were produced at a different time or from a
different file!

### Query
To display information about a horcrux, call `horcrux` with the `-q`/`--query`
flag followed by the filename of the horcrux in question, like:

`horcrux -q file.horcrux`

## Installation
- Do `go install github.com/pepa65/horcrux@latest` anywhere when golang is
properly installed.
- Or, clone this repo by: `git clone https://github.com/pepa65/horcrux`
and do `cd horcrux` followed by `go install`.
- Or for Linux amd64 systems, download the `horcrux` binary in this repo at:
https://github.com/pepa65/horcrux/releases/download/v0.3.2/horcrux
and put it into a location on your PATH, or use it by specifying its path:
`~/horcrux`.

## Usage
```
horcrux v0.3.2 - Split file into encrypted 'horcruxes', mergeable without key
Usage:  horcrux [-n|--number <n>] [-m|--minimum <m>] <file>  |  [<dir>]  |
                -q|--query <horcrux>  |  -V|--version  |  -h|--help
  <n>:        Number of horcruxes to produce [2..255, default: 2]
  <m>:        Minimum number of horcruxes needed for merge [2..n, default: n]
  <file>:     Original file to split up
  <dir>:      Directory with horcruxes to merge [default: current]
  <horcrux>:  The horcrux file to query for information
```
