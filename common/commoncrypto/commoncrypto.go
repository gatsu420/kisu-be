package commoncrypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
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

type HashStringArgs struct {
	Secret string
	Str    string
	Salt   string
}

type HashStringResult struct {
	Digest string
}

func HashString(args HashStringArgs) HashStringResult {
	hash := hmac.New(sha256.New, []byte(args.Secret))
	hash.Write([]byte(args.Str + args.Salt))

	return HashStringResult{
		Digest: base64.URLEncoding.EncodeToString(hash.Sum(nil)),
	}
}

type VerifyHashedStringArgs struct {
	Secret string
	Str    string
	Salt   string
	Digest string
}

type VerifyHashedStringResult struct {
	IsSameHash bool
}

func VerifyHashedString(args VerifyHashedStringArgs) (VerifyHashedStringResult, error) {
	originalHash := hmac.New(sha256.New, []byte(args.Secret))
	originalHash.Write([]byte(args.Str + args.Salt))

	comparisonHash, err := base64.URLEncoding.DecodeString(args.Digest)
	if err != nil {
		return VerifyHashedStringResult{}, fmt.Errorf("unable to decode digest: %w", err)
	}

	return VerifyHashedStringResult{
		IsSameHash: hmac.Equal(originalHash.Sum(nil), comparisonHash),
	}, nil
}

func GetRandomTableName() string {
	return strings.ReplaceAll(uuid.New().String(),
		"-",
		"")
}
