package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	//"github.com/pepa65/horcrux/pkg/multiplexing"
	"github.com/pepa65/horcrux/pkg/shamir"
	"gopkg.in/yaml.v3"
)

func Query(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return errors.New("problem reading file")
	}

	defer file.Close()
	yml, err := getYmlFile(file)
	if err != nil || yml.Filename == "" {
		return errors.New("bad YAML")
	}

	timestamp := time.Unix(yml.Timestamp, 0)
	fmt.Printf("File '%s' was split at %s\n", yml.Filename, timestamp.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Horcrux-file %d of %d (minimum of %d needed to merge)\n", yml.Index, yml.Total, yml.Minimum)
	return nil
}

func Merge(dir string) error {
	dirfiles, err := ioutil.ReadDir(dir)
	if err != nil {
		return errors.New("empty directory")
	}

	filenames := []string{}
	for _, file := range dirfiles {
		if filepath.Ext(file.Name()) == ".yml" {
			filenames = append(filenames, file.Name())
		}
	}
	var ymls = []ymlFile{}
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			return errors.New("problem reading file")
		}

		defer file.Close()
		yml, err := getYmlFile(file)
		if err != nil {
			return errors.New("unparsable YAML")
		}

		if len(ymls) > 0 && (yml.Filename != ymls[0].Filename || yml.Timestamp != ymls[0].Timestamp || yml.Total != ymls[0].Total || yml.Minimum != ymls[0].Minimum || len(yml.Keypart) != len(ymls[0].Keypart)) {
			fmt.Println("All horcrux-files in the directory must have the same atributes (except index, keypart and payload)")
			return errors.New("all horcrux-files in the directory must have the same atributes (except index, keypart and payload)")
		}
		ymls = append(ymls, *yml)
	}
	n := len(ymls)
	if n == 0 {
		return errors.New("no horcrux-files in directory")
	} else if n < ymls[0].Minimum {
		return fmt.Errorf("not enough horcrux-files, %d are needed to reconstruct, only %d here", ymls[0].Minimum, n)
	}

	keyparts := make([][]byte, n)
	for i := range keyparts {
		keyparts[i], err = hex.DecodeString(ymls[i].Keypart)
		if err != nil {
			return err
		}
	}
	key, err := shamir.Combine(keyparts)
	if err != nil {
		return errors.New("problem recombining the keyparts")
	}

	var encfile []byte
	if ymls[0].Total == ymls[0].Minimum {
		// m == n: Recombine sorted by index
		sortedIndex := make([]int, n)
		for i := range ymls {
			sortedIndex[ymls[i].Index-1] = i
		}
		for i := range sortedIndex {
			payload, err := base64.StdEncoding.DecodeString(ymls[sortedIndex[i]].Payload)
			if err != nil {
				return errors.New("error decoding payload")
			}
			encfile = append(encfile, payload...)
		}
	} else {
		// m < n: All files have the same payload
		encfile, err = base64.StdEncoding.DecodeString(ymls[0].Payload)
		if err != nil {
			return errors.New("error decoding payload")
		}
	}
	fileReader := bytes.NewReader(encfile)
	if err != nil {
		return errors.New("error processing base64 code")
	}
	reader := cryptoReader(fileReader, key)
	newFilename := ymls[0].Filename
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
	return err
}

func getYmlFile(file *os.File) (*ymlFile, error) {
	yml := &ymlFile{}
	f, err := io.ReadAll(file)
	err = yaml.Unmarshal(f, yml)
	if err != nil {
		return nil, errors.New("YAML content not found")
	}
	return yml, nil
}
