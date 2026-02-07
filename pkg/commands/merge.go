package commands

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pepa65/horcrux/pkg/multiplexing"
	"github.com/pepa65/horcrux/pkg/shamir"
)

func Query(filename string, version string) error {
	header := &horcruxHeader{}
	file, err := os.Open(filename)
	if err != nil {
		return errors.New("problem reading file")
	}
	defer file.Close()
	header, err = getHeaderFromHorcruxFile(file)
	if err != nil || header.OriginalFilename == "" {
		return errors.New("bad header")
	}
	stamp := time.Unix(header.Timestamp, 0)
	fmt.Printf("Original file '%s' split at %s by horcrux version %d\n", header.OriginalFilename, stamp, header.Version)
	fmt.Printf("Horcrux-file %d of %d (minimum of %d needed to merge)\n", header.Index, header.Total, header.Threshold)
	parts := strings.Split(version, ".")
	if header.Version != parts[0][0] {
		fmt.Printf("This version of horcrux (%v) is incompatible with this horcrux-file!\n", version)
	}
	return nil
}

func Merge(dir string, version string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return errors.New("empty directory")
	}

	filenames := []string{}
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".horcrux" {
			filenames = append(filenames, file.Name())
		}
	}

	headers := []horcruxHeader{}
	horcruxFiles := []*os.File{}

	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			return errors.New("problem reading file")
		}
		defer file.Close()

		currentHeader, err := getHeaderFromHorcruxFile(file)
		if err != nil {
			return errors.New("bad header")
		}

		parts := strings.Split(version, ".")
		if currentHeader.Version != parts[0][0] {
			fmt.Printf("This version of horcrux (%v) is incompatible with horcrux-file '%v'!\n", version, filename)
			os.Exit(1)
		}

		if len(headers) > 0 && (currentHeader.OriginalFilename != headers[0].OriginalFilename || currentHeader.Timestamp != headers[0].Timestamp) {
			fmt.Println("All horcrux-files in the directory must have the same timestamp & orig.filename")
			return errors.New("all horcrux-files in the directory must have the same timestamp & orig.filename")
		}

		headers = append(headers, *currentHeader)
		horcruxFiles = append(horcruxFiles, file)
	}

	if len(headers) == 0 {
		return errors.New("no horcrux-files in directory")
	} else if len(headers) < headers[0].Threshold {
		return fmt.Errorf("not enough horcrux-files, %d are needed to reconstruct the original, only %d here", headers[0].Threshold, len(headers))
	}

	keyFragments := make([][]byte, len(headers))
	for i := range keyFragments {
		keyFragments[i] = headers[i].KeyFragment
	}

	key, err := shamir.Combine(keyFragments)
	if err != nil {
		return errors.New("problem recombining the horcrux-files")
	}

	var fileReader io.Reader
	if headers[0].Total == headers[0].Threshold {
		// Sort by index
		orderedHorcruxFiles := make([]*os.File, len(horcruxFiles))
		for i, h := range horcruxFiles {
			orderedHorcruxFiles[headers[i].Index-1] = h
		}

		fileReader = &multiplexing.Multiplexer{Readers: orderedHorcruxFiles}
	} else {
		fileReader = horcruxFiles[0] // Read the first horcrux: all the same
	}

	reader := cryptoReader(fileReader, key)

	newFilename := headers[0].OriginalFilename
	if fileExists(newFilename) {
		newFilename = prompt("File '%s' already exists here, give a new file name: ", newFilename)
	}

	_ = os.Truncate(newFilename, 0)

	newFile, err := os.OpenFile(newFilename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errors.New("problem writing to file " + newFilename)
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, reader)
	if err != nil {
		return err
	}

	return err
}

// Get header from horcrux, leave read pointer at the encrypted content
func getHeaderFromHorcruxFile(file *os.File) (*horcruxHeader, error) {
	currentHeader := &horcruxHeader{}
	scanner := bufio.NewScanner(file)
	bytesBeforeBody := 0
	var headerFound bool
	for scanner.Scan() {
		line := scanner.Text()
		bytesBeforeBody += len(scanner.Bytes()) + 1
		if line == "-- HEADER --" {
			scanner.Scan()
			bytesBeforeBody += len(scanner.Bytes()) + 1
			headerLine := scanner.Bytes()
			json.Unmarshal(headerLine, currentHeader)
			scanner.Scan() // One more to get past the body line
			bytesBeforeBody += len(scanner.Bytes()) + 1
			headerFound = true
			break
		}
	}
	if _, err := file.Seek(int64(bytesBeforeBody), io.SeekStart); err != nil {
		return nil, errors.New("problem accessing the horcrux-file")
	}

	if !headerFound {
		return nil, errors.New("no header found in the horcrux-file")
	}
	return currentHeader, nil
}
