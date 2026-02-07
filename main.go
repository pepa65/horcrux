package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pepa65/horcrux/pkg/commands"
)

const version = "0.5.1"

var self = ""

func main() {
	path, narg, marg, qarg, split, anypath := "", 0, 0, 0, false, false
	var err error
	var n, m int
	for _, arg := range os.Args {
		if self == "" {
			selves := strings.Split(arg, "/")
			self = selves[len(selves)-1]
			continue
		}
		if narg == 1 { // after -n
			if qarg > 0 {
				usage(nil, "Flag -q/--query can't be used with other flags")
			}
			narg = 2
			n, err = strconv.Atoi(arg)
			if err != nil {
				//fmt.Println(err)
				usage(err, "Argument of -n/--num should be an integer: '"+arg+"'")
			}
			if n < 2 {
				usage(nil, "Argument of -n/--num should be 2 or more")
			}
			continue
		}
		if marg == 1 { // after -m
			if qarg > 0 {
				usage(nil, "Flag -q/--query can't be used with other flags")
			}
			marg = 2
			m, err = strconv.Atoi(arg)
			if err != nil {
				//fmt.Println(err)
				usage(err, "Argument of -m/--min should be an integer: '"+arg+"'")
			}
			if m < 2 {
				usage(nil, "Argument of -m/--min should be 2 or more")
			}
			continue
		}
		if qarg == 1 { // after -q
			if marg > 0 || narg > 0 {
				usage(nil, "Flag -q/--query can't be used with other flags")
			}
			qarg = 2
			path = arg
			continue
		}
		switch arg {
		case "-V", "--version":
			fmt.Printf("%s v%s\n", self, version)
			return
		case "-h", "--help":
			usage(nil, "")
		case "-n", "--num":
			split = true
			if narg > 0 {
				usage(nil, "Multiple '-n/--num' flags")
			}
			narg = 1
		case "-m", "--min":
			split = true
			if marg > 0 {
				usage(nil, "Multiple '-m/--min' flags")
			}
			marg = 1
		case "-q", "--query":
			if qarg > 0 {
				usage(nil, "Multiple '-q/--query' flags")
			}
			qarg = 1
		default:
			if arg == "" {
				usage(nil, "Empty argument")
			}
			if arg == "--" {
				anypath = true
			} else {
				if !anypath && arg[0] == '-' {
					usage(nil, "Unknown flag: "+arg)
				}
				if len(path) > 0 {
					usage(nil, "Redundant argument '"+arg+"' after '"+path+"'")
				}
				path = arg
			}
		}
	}
	if path == "" { // No file/directory given
		if split || qarg > 0 {
			usage(nil, "No file specified")
		}
		path = "."
	} else { // Path specified: file or directory
		fi, err := os.Stat(path)
		if err == nil { // The path exists
			if fi.IsDir() { // Directory
				if split {
					usage(nil, "Can't split a directory")
				}
				if qarg > 0 {
					usage(nil, "A horcrux can't be a directory")
				}
			} else { // File
				if qarg > 0 { // Query
					if err = commands.Query(path); err != nil {
						usage(err, "Query of file '"+path+"' failed")
					}
					return
				}
				split = true
			}
		} else {
			if split { // -n and/or -m given
				usage(nil, "Not a file: "+path)
			}
			usage(nil, "Not a file/directory: "+path)
		}
	}
	if split {
		if n == 0 {
			n = 2
		}
		if m > n {
			usage(nil, "Argument of -m should be less or equal to "+fmt.Sprintf("%d", n))
		}
		if m == 0 { // default minimum is all
			m = n
		}
		if err := commands.Split(path, n, m); err != nil {
			usage(err, "Splitting file '"+path+"' failed")
		}
		return
	}
	// Merge
	if err = commands.Merge(path); err != nil {
		usage(err, "Merge in directory '"+path+"' failed")
	}
}

func usage(e error, err string) {
	fmt.Println(self + " v" + version + " - Split file into 'horcrux-files', reconstitutable without key")
	fmt.Println("Usage:")
	fmt.Println("  - Split & encrypt:  " + self + " [-n|--number N] [-m|--minimum M] FILE")
	fmt.Println("        N:     Number of horcrux-files to produce [2..255, default: 2]")
	fmt.Println("        M:     Min.number of horcrux-files to reconstitute [2..N, default: N]")
	fmt.Println("        FILE:  Original file to split up and encrypt")
	fmt.Println("  - Reconstitute file:  " + self + " [DIR]")
	fmt.Println("       DIR:  Directory with horcrux-files to reconstitute [default: current]")
	fmt.Println("  - Query horcrux-file:  " + self + " -q|--query FILE.horcrux")
	fmt.Println("       FILE.horcrux:  The horcrux-file to query for information")
	fmt.Println("  - Get help or version:  " + self + " -h|--help | -V|--version")
	if e != nil {
		fmt.Println(e)
	}
	if err != "" {
		fmt.Println("\nAbort: " + err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
