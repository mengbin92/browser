package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
)

func Fullname(pbFile string) string {
	return fmt.Sprintf("./pb/%s.pb", pbFile)
}

func CalcPassword(pass, salt string) string {
	newpass, _ := CalcSha384Hash([]byte(pass + salt))
	return hex.EncodeToString(newpass)
}

func CalcSha384Hash(in []byte) ([]byte, error) {
	var sha = sha512.New384()
	_, err := sha.Write(in)
	if err != nil {
		return []byte{}, err
	}

	return sha.Sum([]byte(nil)), nil
}
