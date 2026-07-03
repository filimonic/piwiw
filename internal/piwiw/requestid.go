package piwiw

import "crypto/rand"

// Crockford's Base32 alphabet (excludes I, L, O, U to avoid ambiguity).
const crockfordAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

const requestIDLength = 10

func newRequestID() string {
	raw := make([]byte, requestIDLength)
	_, _ = rand.Read(raw)

	id := make([]byte, requestIDLength)
	for i, b := range raw {
		id[i] = crockfordAlphabet[b&0x1F]
	}
	return string(id)
}
