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
	"time"

	"github.com/pepa65/horcrux/pkg/multiplexing"
	"github.com/pepa65/horcrux/pkg/shamir"
)

func Query(filename string) error {
	header := &horcruxHeader{}
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return errors.New("Problem reading file")
	}
	header, err = getHeaderFromHorcruxFile(file)
	if err != nil || header.OriginalFilename == "" {
		return errors.New("Bad header")
	}
	stamp := time.Unix(header.Timestamp, 0)
	fmt.Printf("Original file '%s' split at %s\n", header.OriginalFilename,
		stamp)
	fmt.Printf("Horcrux %d of %d (minimum of %d needed to merge)\n",
		header.Index, header.Total, header.Threshold)
	return nil
}

func Merge(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return errors.New("Empty directory")
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
		defer file.Close()
		if err != nil {
			return errors.New("Problem reading file")
		}

		currentHeader, err := getHeaderFromHorcruxFile(file)
		if err != nil {
			return errors.New("Bad header")
		}

		if len(headers) > 0 && (currentHeader.OriginalFilename != headers[0].OriginalFilename || currentHeader.Timestamp != headers[0].Timestamp) {
			fmt.Println("All horcruxes in the directory must have the same timestamp and filename")
			return errors.New("All horcruxes in the directory must have the same timestamp and filename")
		}

		headers = append(headers, *currentHeader)
		horcruxFiles = append(horcruxFiles, file)
	}

	if len(headers) == 0 {
		return errors.New("No horcruxes in directory")
	} else if len(headers) < headers[0].Threshold {
		return errors.New(fmt.Sprintf("Not enough horcruxes, %d are needed to resurrect the original, only %d here", headers[0].Threshold, len(headers)))
	}

	keyFragments := make([][]byte, len(headers))
	for i := range keyFragments {
		keyFragments[i] = headers[i].KeyFragment
	}

	key, err := shamir.Combine(keyFragments)
	if err != nil {
		return errors.New("Problem recombining the horcruxes")
	}

	var fileReader io.Reader
	if headers[0].Total == headers[0].Threshold {
		// sort by index
		orderedHorcruxFiles := make([]*os.File, len(horcruxFiles))
		for i, h := range horcruxFiles {
			orderedHorcruxFiles[headers[i].Index-1] = h
		}

		fileReader = &multiplexing.Multiplexer{Readers: orderedHorcruxFiles}
	} else {
		fileReader = horcruxFiles[0] // arbitrarily read from the first horcrux: they all contain the same contents
	}

	reader := cryptoReader(fileReader, key)

	newFilename := headers[0].OriginalFilename
	if fileExists(newFilename) {
		newFilename = prompt("File '%s' already exists here, give new file name: ", newFilename)
	}

	_ = os.Truncate(newFilename, 0)

	newFile, err := os.OpenFile(newFilename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errors.New("Problem writing to file " + newFilename)
	}
	defer newFile.Close()

	_, err = io.Copy(newFile, reader)
	if err != nil {
		return err
	}

	return err
}

// Get header from horcrux file and leave its read pointer at the encrypted
// content for later reading
func getHeaderFromHorcruxFile(file *os.File) (*horcruxHeader, error) {
	currentHeader := &horcruxHeader{}
	scanner := bufio.NewScanner(file)
	bytesBeforeBody := 0
	for scanner.Scan() {
		line := scanner.Text()
		bytesBeforeBody += len(scanner.Bytes()) + 1
		if line == "-- HEADER --" {
			scanner.Scan()
			bytesBeforeBody += len(scanner.Bytes()) + 1
			headerLine := scanner.Bytes()
			json.Unmarshal(headerLine, currentHeader)
			scanner.Scan() // one more to get past the body line
			bytesBeforeBody += len(scanner.Bytes()) + 1
			break
		}
	}
	if _, err := file.Seek(int64(bytesBeforeBody), io.SeekStart); err != nil {
		return nil, errors.New("Problem accessing the horcrux")
	}

	if currentHeader == nil {
		return nil, errors.New("No header found in horcrux file")
	}
	return currentHeader, nil
}
