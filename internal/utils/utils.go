package utils

import (
	"crypto/sha512"
	"encoding/hex"
)

func CalcPassword(pass, salt string) string {
	newpass, _ := CalcSha512Hash([]byte(pass + salt))
	return hex.EncodeToString(newpass)
}

func CalcSha512Hash(in []byte) ([]byte, error) {
	var sha = sha512.New()
	_, err := sha.Write(in)
	if err != nil {
		return []byte{}, err
	}

	return sha.Sum([]byte(nil)), nil
}
