package commoncrypto

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
)

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

func GetRandomTableName() string {
	return strings.ReplaceAll(uuid.New().String(),
		"-",
		"")
}
