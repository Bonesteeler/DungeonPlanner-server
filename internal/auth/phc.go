package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"

	"go.pact.im/x/option"
	"go.pact.im/x/phcformat"
	"go.pact.im/x/phcformat/encode"
	"golang.org/x/crypto/argon2"
)

const algorithm = "argon2id"
const m = 64*1024
const t = 1
const p = 4
const keyLen = 32

func GenerateString(password string) string {
	salt := GenerateSalt()
	key := argon2.IDKey([]byte(password), salt, t, m, p, keyLen)
	out := phcformat.Append(
		nil,
		encode.NewString(algorithm),
		option.Value(encode.NewString(strconv.Itoa(argon2.Version))),
		option.Value(encode.NewString(fmt.Sprintf("m=%d,t=%d,p=%d", m, t, p))),
		option.Value(encode.NewBase64(salt)),
		option.Value(encode.NewBase64(key)),
	)
	return string(out)
}

func ValidateString(password string, expected string) bool {
	hash, errBool := phcformat.Parse(expected)
	if errBool != true {
		return false
	}
	// Salt
	salt, errBool := hash.Salt.Unwrap()
	if errBool != true {
		return false
	}
	saltBytes, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return false
	}
	// Params
	paramsString, errBool := hash.Params.Unwrap()
	if errBool != true {
		return false
	}
	parsedM, parsedT, parsedP, err := ParseParams(paramsString)
	if err != nil {
		return false
	}
	// Output
	expectedOutput, errBool := hash.Output.Unwrap()
	if errBool != true {
		return false
	}
	result := argon2.IDKey(
		[]byte(password),
		saltBytes,
		parsedT,
		parsedM,
		parsedP,
		keyLen)
	return subtle.ConstantTimeCompare([]byte(base64.RawStdEncoding.EncodeToString(result)), []byte(expectedOutput)) == 1
}

func ParseParams(paramsString string) (uint32, uint32, uint8, error) {
	err := error(nil)
	parsedM := uint64(0)
	parsedT := uint64(0)
	parsedP := uint64(0)
	paramItr := phcformat.IterParams(paramsString)
	params := paramItr.Iter()
	for k, v := range params {
		switch k {
		case "m":
			parsedM, err = strconv.ParseUint(v, 10, 32)
		case "t":
			parsedT, err = strconv.ParseUint(v, 10, 32)
		case "p":
			parsedP, err = strconv.ParseUint(v, 10, 8)
		default:
			return 0, 0, 0, fmt.Errorf("unknown parameter: %s", k)
		}
	}
	if parsedM == 0 || parsedT == 0 || parsedP == 0 {
		return 0, 0, 0, fmt.Errorf("missing required parameters")
	}
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse parameter: %w", err)
	}
	return uint32(parsedM), uint32(parsedT), uint8(parsedP), nil
}

func GenerateSalt() []byte {
	salt := make([]byte, 16)
	rand.Read(salt)
	return salt
}