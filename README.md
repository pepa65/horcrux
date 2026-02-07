[![Go Report Card](https://goreportcard.com/badge/github.com/pepa65/horcrux)](https://goreportcard.com/report/github.com/pepa65/horcrux)
[![GoDoc](https://godoc.org/github.com/pepa65/horcrux?status.svg)](https://godoc.org/github.com/pepa65/horcrux)
<img src="https://raw.githubusercontent.com/pepa65/horcrux/master/horcrux.png" width="96" alt="horcrux icon" align="right">
# horcrux v0.5.1
**Split file into horcrux-files, reconstitutable without key**

* Repo: https://github.com/pepa65/horcrux
* After https://github.com/jesseduffield/horcrux
* The technique used is Hashicorp's Shamir's secret sharing, based on Age encryption.

## Function
The program `horcrux` can split a file into a predefined number of encrypted horcrux-files,
or reconstitute a (predefinable) sufficient number of constituent horcrux-files in a directory
back into the original file.

### Split
To split up a file into horcrux-files, call `horcrux` with the filename and
optionally flags `-n`/`--num` with the number of desired horcrux-files and/or
`-m`/`--min` followed by the minimum number of horcrux-files needed to reconstitute the original file.
Example:

`horcrux -n 5 -m 3 secret.txt`

The resulting horcrux-files can be renamed, dispersed (given to different people or
stored at different locations) and later be used to reconstitute the original file
if the minimum number of needed horcrux-files is met (in this case: 3 out of the 5 are needed).

### Reconstitute
To merge horcrux-files back into the original file, call `horcrux` in the
directory containing the horcruxes, or pass that directory as an argument:
`horcrux directory/with/horcruxes`

Only files ending in `.horcrux` will be taken into account, other files will be ignored.
(There should not be any horcrux-files in that directory that were produced at a different time or from a
different file!)

### Query
To display information about a horcrux-file, call `horcrux` with the `-q`/`--query`
flag followed by the filename of the horcrux-file, like:

`horcrux -q file.horcrux`

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

(Needs upx installed.)

## Usage
```
horcrux v0.5.1 - Split file into 'horcrux-files', reconstitutable without key
Usage:
  - Split & encrypt:  horcrux [-n|--number N] [-m|--minimum M] FILE
        N:     Number of horcrux-files to produce [2..255, default: 2]
        M:     Min.number of horcrux-files to reconstitute [2..n, default: n]
        FILE:  Original file to split up and encrypt
  - Reconstitute file:  horcrux [DIR]
       DIR:  Directory with horcrux-files to reconstitute [default: current]
  - Query horcrux-file:  horcrux -q|--query FILE.horcrux
       FILE.horcrux:  The horcrux-file to query for information
  - Get help or version:  horcrux -h|--help | -V|--version
```
