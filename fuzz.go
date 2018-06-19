// +build gofuzz

package expand

import "os"

func Fuzz(b []byte) int {
	Expand(string(b), os.LookupEnv)
	return 0
}
