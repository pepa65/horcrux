// After https://github.com/hashicorp/vault/blob/master/shamir/shamir.go

package shamir

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	mrand "math/rand"
	"time"
)

const (
	// ShareOverhead is the byte size overhead of each share
	// when using Split on a key. This is caused by appending
	// a one byte tag to the share.
	ShareOverhead = 1
)

// polynomial represents a polynomial of arbitrary degree
type polynomial struct {
	coefficients []uint8
}

// makePolynomial constructs a random polynomial of the given
// degree but with the provided intercept value.
func makePolynomial(intercept, degree uint8) (polynomial, error) {
	// Create a wrapper
	p := polynomial{
		coefficients: make([]byte, degree+1),
	}

	// Ensure the intercept is set
	p.coefficients[0] = intercept

	// Assign random co-efficients to the polynomial
	if _, err := rand.Read(p.coefficients[1:]); err != nil {
		return p, err
	}

	return p, nil
}

// evaluate returns the value of the polynomial for the given x
func (p *polynomial) evaluate(x uint8) uint8 {
	// Special case the origin
	if x == 0 {
		return p.coefficients[0]
	}

	// Compute the polynomial value using Horner's method.
	degree := len(p.coefficients) - 1
	out := p.coefficients[degree]
	for i := degree - 1; i >= 0; i-- {
		coeff := p.coefficients[i]
		out = add(mult(out, x), coeff)
	}
	return out
}

// interpolatePolynomial takes N sample points and returns
// the value at a given x using a lagrange interpolation.
func interpolatePolynomial(x_samples, y_samples []uint8, x uint8) uint8 {
	limit := len(x_samples)
	var result, basis uint8
	for i := 0; i < limit; i++ {
		basis = 1
		for j := 0; j < limit; j++ {
			if i == j {
				continue
			}
			num := add(x, x_samples[j])
			denom := add(x_samples[i], x_samples[j])
			term := div(num, denom)
			basis = mult(basis, term)
		}
		group := mult(y_samples[i], basis)
		result = add(result, group)
	}
	return result
}

// div divides two numbers in GF(2^8)
func div(a, b uint8) uint8 {
	if b == 0 {
		// leaks some timing information but we don't care anyways as this
		// should never happen, hence the panic
		panic("divide by zero")
	}

	var goodVal, zero uint8
	log_a := logTable[a]
	log_b := logTable[b]
	diff := (int(log_a) - int(log_b)) % 255
	if diff < 0 {
		diff += 255
	}

	ret := expTable[diff]

	// Ensure we return zero if a is zero but aren't subject to timing attacks
	goodVal = ret

	if subtle.ConstantTimeByteEq(a, 0) == 1 {
		ret = zero
	} else {
		ret = goodVal
	}

	return ret
}

// mult multiplies two numbers in GF(2^8)
func mult(a, b uint8) (out uint8) {
	var goodVal, zero uint8
	log_a := logTable[a]
	log_b := logTable[b]
	sum := (int(log_a) + int(log_b)) % 255

	ret := expTable[sum]

	// Ensure we return zero if either a or b are zero but aren't subject to
	// timing attacks
	goodVal = ret

	if subtle.ConstantTimeByteEq(a, 0) == 1 {
		ret = zero
	} else {
		ret = goodVal
	}

	if subtle.ConstantTimeByteEq(b, 0) == 1 {
		ret = zero
	}

	return ret
}

// add combines two numbers in GF(2^8)
// This can also be used for subtraction since it is symmetric.
func add(a, b uint8) uint8 {
	return a ^ b
}

// Split takes an arbitrarily long key and generates a `number`
// number of shares, `minimum` of which are required to reconstruct
// the key. The number and minimum must be at least 2, and less
// than 256. The returned shares are each one byte longer than the key
// as they attach a tag used to reconstruct the key.
func Split(key []byte, number, minimum int) ([][]byte, error) {
	// Sanity check the input
	if number < minimum {
		return nil, fmt.Errorf("number cannot be less than minimum")
	}

	if number > 255 {
		return nil, fmt.Errorf("number cannot exceed 255")
	}

	//// Not technically required
	//if minimum < 2 {
	//	return nil, fmt.Errorf("minimum must be at least 2")
	//}

	if len(key) == 0 {
		return nil, fmt.Errorf("cannot split an empty key")
	}

	// Generate random list of x coordinates
	mrand.Seed(time.Now().UnixNano())
	xCoordinates := mrand.Perm(255)

	// Allocate the output array, initialize the final byte
	// of the output with the offset. The representation of each
	// output is {y1, y2, .., yN, x}.
	out := make([][]byte, number)
	for idx := range out {
		out[idx] = make([]byte, len(key)+1)
		out[idx][len(key)] = uint8(xCoordinates[idx]) + 1
	}

	// Construct a random polynomial for each byte of the key.
	// Because we are using a field of size 256, we can only represent
	// a single byte as the intercept of the polynomial, so we must
	// use a new polynomial for each byte.
	for idx, val := range key {
		p, err := makePolynomial(val, uint8(minimum-1))
		if err != nil {
			return nil, err
		}

		// Generate a `number` number of (x,y) pairs
		// We cheat by encoding the x value once as the final index,
		// so that it only needs to be stored once.
		for i := 0; i < number; i++ {
			x := uint8(xCoordinates[i]) + 1
			y := p.evaluate(x)
			out[i][idx] = y
		}
	}

	// Return the encoded keys
	return out, nil
}

// Combine is used to reverse a Split and reconstruct a key
// once a `minimum` number of keyparts are available.
func Combine(keyparts [][]byte) ([]byte, error) {
	//// Not technically required
	// Verify enough keyparts provided
	//if len(keyparts) < 2 {
	//	return nil, fmt.Errorf("less than two keyparts cannot be used to reconstruct the key")
	//}

	// Verify the keyparts are all the same length
	firstPartLen := len(keyparts[0])
	if firstPartLen < 2 {
		return nil, fmt.Errorf("keyparts must be at least two bytes")
	}
	for i := 1; i < len(keyparts); i++ {
		if len(keyparts[i]) != firstPartLen {
			return nil, fmt.Errorf("all keyparts must be the same length")
		}
	}

	// Create a buffer to store the reconstructed key
	key := make([]byte, firstPartLen-1)

	// Buffer to store the samples
	x_samples := make([]uint8, len(keyparts))
	y_samples := make([]uint8, len(keyparts))

	// Set the x value for each sample and ensure no x_sample values are the same,
	// otherwise div() can be unhappy
	checkMap := map[byte]bool{}
	for i, keypart := range keyparts {
		samp := keypart[firstPartLen-1]
		if exists := checkMap[samp]; exists {
			return nil, fmt.Errorf("duplicate keypart detected")
		}
		checkMap[samp] = true
		x_samples[i] = samp
	}

	// Reconstruct each byte
	for idx := range key {
		// Set the y value for each sample
		for i, keypart := range keyparts {
			y_samples[i] = keypart[idx]
		}

		// Interpolate the polynomial and compute the value at 0
		val := interpolatePolynomial(x_samples, y_samples, 0)

		// Evaluate the 0th value to get the intercept
		key[idx] = val
	}
	return key, nil
}
