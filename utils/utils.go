package utils

import "fmt"

func Fullname(pbFile string) string {
	return fmt.Sprintf("./pb/%s.pb", pbFile)
}
