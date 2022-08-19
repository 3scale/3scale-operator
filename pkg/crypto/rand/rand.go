package crypto

import (
	"math/rand"
	"time"
)

const lowercaseAlphabetCharset = "abcdefghijklmnopqrstuvwxyz"
const uppercaseAlphabetCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numericCharset = "0123456789"
const alphanumericCharset = lowercaseAlphabetCharset + uppercaseAlphabetCharset + numericCharset
const hexadecimalCharset = numericCharset + "ABCDEF"

var randomGenerator *rand.Rand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

// String generates random alphanumeric string of size 'size'.
func String(length int) string {
	return StringWithCharset(length, alphanumericCharset)
}

// HexadecimalString generates random hexadecimal strings of size 'size'
func HexadecimalString(length int) string {
	return StringWithCharset(length, hexadecimalCharset)
}

// StringWithCharset generates random string of length 'length' with all of its
// random characters existing in and only in the 'charset' set of
// strings
func StringWithCharset(length int, charset string) string {
	result := make([]byte, length)

	for i := range result {
		result[i] = charset[randomGenerator.Int63()%int64(len(charset))]
	}

	return string(result)
}
