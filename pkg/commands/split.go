package commands

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pepa65/horcrux/pkg/multiplexing"
	"github.com/pepa65/horcrux/pkg/shamir"
)

func Split(path string, n int, m int) error {
	key, err := generateKey()
	if err != nil {
		return errors.New("Problem generating a random key")
	}
	keyFragments, err := shamir.Split(key, n, m)
	if err != nil {
		return errors.New("Problem splitting the key")
	}

	timestamp := time.Now().Unix()
	file, err := os.Open(path)
	if err != nil {
		return errors.New("Problem opening the file")
	}
	originalFilename := filepath.Base(path)
	horcruxFiles := make([]*os.File, n)

	for i := range horcruxFiles {
		index := i + 1
		headerBytes, err := json.Marshal(&horcruxHeader{
			OriginalFilename: originalFilename,
			Timestamp:        timestamp,
			Index:            index,
			Total:            n,
			KeyFragment:      keyFragments[i],
			Threshold:        m,
		})
		if err != nil {
			return errors.New("Problem making the header into JSON")
		}

		originalFilenameWithoutExt := strings.TrimSuffix(originalFilename, filepath.Ext(originalFilename))
		horcruxFilename := fmt.Sprintf("%s_%dof%d.horcrux", originalFilenameWithoutExt, index, n)
		fmt.Printf("creating %s\n", horcruxFilename)

		// Clearing file in case it already existed
		_ = os.Truncate(horcruxFilename, 0)
		horcruxFile, err := os.OpenFile(horcruxFilename, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return errors.New("Problem writing horcrux file " + horcruxFilename)
		}
		defer horcruxFile.Close()

		horcruxFile.WriteString(header(originalFilename, index, n, m, headerBytes))
		horcruxFiles[i] = horcruxFile
	}

	// Wrap file reader in an encryption stream
	var fileReader io.Reader = file
	reader := cryptoReader(fileReader, key)
	var writer io.Writer
	if m == n {
		// All horcruxes are needed to resurrect the original, so use multiplexer
		// to divide the encrypted content evenly between the horcruxes
		writer = &multiplexing.Demultiplexer{Writers: horcruxFiles}
	} else {
		writers := make([]io.Writer, len(horcruxFiles))
		for i := range writers {
			writers[i] = horcruxFiles[i]
		}
		writer = io.MultiWriter(writers...)
	}
	_, err = io.Copy(writer, reader)
	if err != nil {
		return errors.New("Problem copying the horcruxes")
	}

	fmt.Println("Done!")
	return nil
}

func header(name string, index int, n int, m int, headerBytes []byte) string {
	return fmt.Sprintf(`/* This is a 'horcrux', an encrypted fragment of '%s'. It is number %d of %d horcruxes that contain parts of the original file. They can be merged when at least %d fragments are present with the program found here: https://github.com/pepa65/horcrux */
-- HEADER --
%s
-- BODY --
`, name, index, n, m, headerBytes)
}

func generateKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	return key, err
}
