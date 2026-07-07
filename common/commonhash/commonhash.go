package commonhash

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
)

type ctxKey int

const RandomIntCtxKey ctxKey = 0

func GetRandomInt() (int, error) {
	upperBound, lowerBound := 999999, -999999
	randomBigInt, err := rand.Int(rand.Reader,
		big.NewInt(int64(upperBound-lowerBound)))
	if err != nil {
		return 0, fmt.Errorf("unable to create random *big.Int number")
	}

	return int(randomBigInt.Int64()) + 1 + lowerBound, nil
}

func HashStringSlice(slc []string, salt string) []string {
	var result []string
	checkUnique := map[string]struct{}{}

	// use plain sha256 because bigquery has no function equivalent
	// to hmac
	hash := sha256.New()

	for _, s := range slc {
		hash.Reset()
		hash.Write([]byte(s + salt))
		digest := base64.StdEncoding.EncodeToString(hash.Sum(nil))

		_, ok := checkUnique[digest]
		if !ok {
			checkUnique[digest] = struct{}{}
			result = append(result, digest)
		}
	}

	return result
}

func HashString(secret []byte, str string, salt string) string {
	hash := hmac.New(sha256.New, secret)
	hash.Write([]byte(str + salt))

	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}
