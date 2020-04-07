package main

import (
	"fmt"
	"strconv"
	"strings"
	"os"
	"github.com/pepa65/horcrux/pkg/commands"
)

var self = ""

func main() {
	n, m, path,	narg, marg, split, anypath := 2, 0, "", 0, 0, false, false
	for _, arg := range os.Args {
		if self=="" {
			selves := strings.Split(arg, "/")
			self = selves[len(selves)-1]
			continue
		}
		if narg==1 { // after -n
			narg = 2
			n, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Println(err)
				usage("Argument of -n should be an integer: '"+arg+"'")
			}
			if n<2 {
				usage("Argument of -n should be at least 2")
			}
			continue
		}
		if marg==1 { // after -m
			marg = 2
			m, err := strconv.Atoi(arg)
			if err != nil {
				fmt.Println(err)
				usage("Argument of -m should be an integer: '"+arg+"'")
			}
			if m<2 {
				usage("Argument of -m should be at least 2")
			}
			continue
		}
		switch arg {
		case "-h", "--help":
			usage("")
		case "-n", "--num":
			split = true
			if narg>0 {
				usage("Multiple '-n' flags")
			}
			narg = 1
		case "-m", "--min":
			split = true
			if marg>0 {
				usage("Multiple '-m' flags")
			}
			marg = 1
		default:
			if arg=="" {
				usage("Empty argument")
			}
			if arg=="--" {
				anypath = true
			} else {
				if !anypath && arg[0]=='-' {
					usage("Unknown flag: "+arg)
				}
				if len(path)>0 {
					usage("Redundant argument '"+arg+"' after '"+path+"'")
				}
				path = arg
			}
		}
	}
	if len(path)==0 { // No file/directory given
		if split {
			usage("No file specified")
		} else {
			path = "."
		}
	} else { // Path specified: file or directory
		fi, err := os.Stat(path)
		if err==nil { // The path exists
			if fi.IsDir() { // Directory
				if split {
					usage("Can't split a directory: "+path)
				}
			} else { // File
				split = true
			}
		} else {
			if split { // -n and/or -m given
				usage("Not a file: "+path)
			} else {
				usage("Not a file/directory: "+path)
			}
		}
	}
	if split {
		if m>n {
			usage("Argument of -m should less or equal to "+string(n))
		}
		if m==0 { // default minimum is all
			m = n
		}
		if err := commands.Split(path, n, m); err != nil {
			fmt.Println(err)
			usage("Splitting file '"+path+"' failed")
		}
	} else { // Merge
		if commands.Merge(path) != nil {
			usage("Merge in directory '"+path+"' failed")
		}
	}
}

func usage(err string) {
	fmt.Println(self+" - Split file into encrypted 'horcruxes', remergeable without password")
	fmt.Printf("Usage:  %s [-n|--num <num>][-m|--min <min>] <file> | [<dir>] | -h|--help\n", self)
	fmt.Println("  <num>:   Number of horcruxes to produce [2..255, default: 2]")
	fmt.Println("  <min>:   Minimum number of horcruxes needed for merge [2..num, default: num]")
	fmt.Println("  <file>:  Original file to split up")
	fmt.Println("  <dir>:   Directory with horcruxes to merge [default: current]")
	if err!="" {
		fmt.Println("Abort: "+err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
