package commands

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"os"
	"strings"
)

func cryptoReader(reader io.Reader, key []byte) io.Reader {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewCTR(block, iv[:])
	return cipher.StreamReader{S: stream, R: reader}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func prompt(message string, args ...interface{}) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(message, args...)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
