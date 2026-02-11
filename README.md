[![Go Report Card](https://goreportcard.com/badge/github.com/pepa65/horcrux)](https://goreportcard.com/report/github.com/pepa65/horcrux)
[![GoDoc](https://godoc.org/github.com/pepa65/horcrux?status.svg)](https://godoc.org/github.com/pepa65/horcrux)
<img src="https://raw.githubusercontent.com/pepa65/horcrux/master/horcrux.png" width="96" alt="horcrux icon" align="right">
# horcrux v1.2.0
**Split file into encrypted horcrux-files, reconstructable without key**

* Repo: https://github.com/pepa65/horcrux
* After https://github.com/jesseduffield/horcrux
* The technique used is Hashicorp's Shamir's secret sharing, based on 256 bit AES encryption with CTR.
  When not all split parts are required to reconstruct, every part contains the data for the whole file,
  but only part of the needed key to decrypt it! In case all parts are required, the original file data is split up too.
  It is possible to only require 1 part for decryption, but in that case only the horcrux binary and 1 file is needed..!
* Versions of `horcrux` before 1.0.0 (0.5.2 and below) used OFB, are less secure and should no longer be used.
  Version 1.0.0 and higher use CTR and a different horcrux-file format.

## Function
The program `horcrux` can split a file into a predefined number of encrypted horcrux-files,
and reconstruct an original file from a (predefinable) sufficient number of horcrux-files in a directory.

### Split
To split up a file into horcrux-files, call `horcrux` with the filename and
optionally flags `-n`/`--number` with the number of desired horcrux-files and/or
`-m`/`--minimum` followed by the minimum number of horcrux-files needed to reconstruct the original file.
Example:

`horcrux -n 5 -m 3 secret.txt`

The resulting horcrux-files can be renamed, dispersed (given to different people or
stored at different locations) and later be used to reconstruct the original file
if the minimum number of needed horcrux-files are present (in this case: 3 out of the 5 are needed).

### Reconstruct
To merge horcrux-files back into the original file, call `horcrux` in the directory containing the
horcrux-files (`.yml`, or in the case of `horcrux --zstd`: `.horcrux`).
Alternatively, that directory can be given as an argument: `horcrux directory/with/horcrux-files`

All other files with non-matching names will be ignored. There should not be any horcrux-files with the
same extention in that same directory that were produced with a different command!

### Query
To display information about a horcrux-file, call `horcrux` with the `-q`/`--query`
flag followed by the filename of the horcrux-file, like:

`horcrux -q file.horcrux`

Horcrux files ending in `.yml` can also just be opened as a text file to see all information about them.

## Installation
### Download
Download any of `horcrux` `horcrux_pi` `horcrux_bsd` `horcrux_osx` `horcrux.exe` through:

`wget -O BINARY https://gitlab.com/pepa65/horcrux/-/jobs/artifacts/master/raw/BINARY?job=building`

(replace BINARY with the desired binary name, e.g. `horcrux is for Linux amd64).

Then put it into a location on your PATH, or use it by specifying its path, like:
`./horcrux`

### Go install
Or do `go install github.com/pepa65/horcrux@latest` (needs Golang fully installed).

### Git clone and go install
Or, clone this repo by: `git clone https://github.com/pepa65/horcrux`
and do `cd horcrux` followed by `go install` (needs both Git and Golang).

#### Better binaries
```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install -ldflags="-s -w"
CGO_ENABLED=0 GOOS=linux GOARCH=arm go install -ldflags="-s -w" -o horcrux_pi
CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go install -ldflags="-s -w" -o horcrux_freebsd
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go install -ldflags="-s -w" -o horcrux_osx
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go install -ldflags="-s -w" -o horcrux.exe
```

#### More extreme shrinking
`upx --best --lzma horcrux`

(Needs `upx` or `upx-ucl` installed.)

## Usage
```
horcrux v1.2.0 - Split file into 'horcrux-files', reconstructable without key
Usage:
- Split:  horcrux [-f|--force] [-z|--zstd] [-n|--number N] [-m|--min M] FILE
  -f/--force:  Created horcrux-files will overwrite existing files
  -z/--zstd:   Compress each horcrux-file and give it the .horcrux extension
    N:     Number of horcrux-files to produce [1..255, default: 2]
    M:     Min.number of horcrux-files needed to reconstruct [1..N, default: N]
    FILE:  Original file to split up and encrypt
- Reconstruct file:  horcrux [DIR]
    DIR:  Directory with horcrux-files to reconstruct [default: current]
- Query horcrux-file:  horcrux -q|--query FILE
    FILE:  Horcrux-file to query for information (.yml files can be viewed too)
- Get help or version:  horcrux -h|--help | -V|--version
```
