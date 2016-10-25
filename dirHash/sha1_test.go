package dirHash

import (
	"fmt"
	"testing"
)

func Test_Getsha1(t *testing.T) {
	var testCases = []struct {
		file string
		hash string
	}{
		{"sha1.go", "[38 54 244 102 66 11 34 105 44 234 49 130 194 1 200 1 95 132 191 243]"},
		{"../readme.txt", "[109 53 254 237 121 203 198 252 155 118 239 20 21 12 29 229 198 235 209 13]"}}

	for _, test := range testCases {
		h := fmt.Sprintf("%v", getHash(test.file))

		if h != test.hash {
			t.Errorf("getHash(%s)=%v;\nout %s", test.file, h, test.hash)
		}
	}
}
