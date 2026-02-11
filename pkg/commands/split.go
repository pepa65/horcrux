package commands

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pepa65/horcrux/pkg/shamir"
)

func Split(path string, n int, m int) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.New("error opening the file")
	}

	info, _ := file.Stat()
	towrite := info.Size()
	filename := info.Name()

	key := make([]byte, 32)
	_, err = rand.Read(key)
	if err != nil {
		return errors.New("error generating a random key")
	}

	encReader := cryptoReader(file, key)
	var b64full bytes.Buffer
	if n > m {
		b64enc := base64.NewEncoder(base64.StdEncoding, &b64full)
		_, err := io.Copy(b64enc, encReader)
		if err != nil {
			return err
		}
		b64enc.Close()
	}
	keyparts, err := shamir.Split(key, n, m)
	if err != nil {
		return errors.New("error splitting the key")
	}

	timestamp := time.Now().Unix()
	for i, k := range keyparts {
		yaml := fmt.Sprintf("filename: %q\ntimestamp: %d\nindex: %d\ntotal: %d\nminimum: %d\nkeypart: %x\npayload: ", filename, timestamp, i+1, n, m, k)
		partname := fmt.Sprintf("%s_horcrux%dof%d.yml", filename, i+1, n)
		fmt.Printf("creating: %s\n", partname)
		// Overwriting any existing file, perhaps should be forced with a flag?
		_ = os.Truncate(partname, 0)
		partfile, err := os.OpenFile(partname, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return errors.New("error opening horcrux-file for writing: " + partname)
		}
		defer partfile.Close()
		partfile.WriteString(yaml)

		if m == n {
			size := towrite / int64(n-i)
			towrite -= size
			b64 := base64.NewEncoder(base64.StdEncoding, partfile)
			_, err := io.CopyN(b64, encReader, size)
			if err != nil && err != io.EOF {
				return err
			}
			b64.Close()
		} else { // m < n
			partfile.Write(b64full.Bytes())
		}
		partfile.Write([]byte("\n"))
	}
	return nil
}
