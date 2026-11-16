package grant

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
const codeGroups = 3
const codeGroupSize = 4

func NewCode() (string, error) {
	groups := make([]string, codeGroups)
	for g := range groups {
		letters := make([]byte, codeGroupSize)
		for i := range letters {
			at, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
			if err != nil {
				return "", err
			}
			letters[i] = codeAlphabet[at.Int64()]
		}
		groups[g] = string(letters)
	}
	return strings.Join(groups, "-"), nil
}
