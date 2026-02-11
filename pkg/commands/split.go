package commands

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
	"github.com/pepa65/horcrux/pkg/shamir"
)

func Split(path string, n int, m int, compress bool, force bool) error {
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
	payloadfull := b64full.String()
	keyparts, err := shamir.Split(key, n, m)
	if err != nil {
		return errors.New("error splitting the key")
	}

	partnames := make([]string, 0, n)
	timestamp := time.Now().Unix()
	for i, k := range keyparts {
		payload := payloadfull
		partname := fmt.Sprintf("%s_horcrux%dof%d.yml", filename, i+1, n)
		if m == n {
			size := towrite / int64(n-i)
			towrite -= size
			part := make([]byte, size)
			_, err := io.ReadFull(encReader, part)
			if err != nil && err != io.EOF {
				return err
			}
			payload = base64.StdEncoding.EncodeToString(part)
		}
		yaml := []byte(fmt.Sprintf("filename: %q\ntimestamp: %d\nindex: %d\ntotal: %d\nminimum: %d\nkeypart: %x\npayload: %s\n", filename, timestamp, i+1, n, m, k, payload))
		if compress {
			partname = fmt.Sprintf("%s_%dof%d.horcrux", filename, i+1, n)
		}
		if !force {
			_, err := os.Stat(partname)
			if err == nil {
				return fmt.Errorf("file '%s' already exists", partname)
			}
		}
		partfile, err := os.Create(partname)
		if err != nil {
    	return err
		}
		defer partfile.Close()
		if compress {
			zwriter, err := zstd.NewWriter(partfile, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
			if err != nil {
  	  	return err
			}
			defer zwriter.Close()
			_, err = zwriter.Write(yaml)
			if err != nil {
  	  	return err
			}
		} else {
			_, err = partfile.Write(yaml)
			if err != nil {
				return err
			}
		}
		partnames = append(partnames, partname)
	}
	fmt.Printf("Written: %s\n", strings.Join(partnames, " "))
	return nil
}
