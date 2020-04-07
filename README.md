# horcrux
**Split file into encrypted horcruxes, remergeable without password**
After https://github.com/jesseduffield/horcrux

## Function
The program `horcrux` can split a file into a predefined number of horcruxes,
or merge a predefinable sufficient number of constituent horcruxes in a
directory back into the original file.

### Splitting
To split up a file into horcruxes, call `horcrux` with the filename and
optionally flags -n/--num with the number of desired horcruxes or -m/--min
followed by the minimum number of horcruxes needed to merge back into the
original file. Example:

`horcrux -n 5 -m 3 secret.txt`

The resulting horcruxes can be dispersed and later be used to put the original
file back together if the minimum number is met.

### Merging
To merge horcruxes back into the original file, call `horcrux` in the
directory containing the horcruxes, or pass that directory as an argument:

`horcrux directory/with/horcruxes`

There should not be any horcruxes in there that were produced at a different
time or from a different file!

## Installation
Do `go get github.com/pepa65/horcrux` anywhere when golang is properly
installed, or clone this repo by: `git clone https://github.com/pepa65/horcrux`
and do `cd horcrux` followed by `go build`.
