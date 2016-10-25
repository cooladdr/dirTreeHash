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
		{"../main.go", "89345168873d82091a01384ac5d6789ec4df778e"},
		{"../readme.txt", "831f45b138ef60c3524fd98194584c8836446b8e"}}

	for _, test := range testCases {
		h := fmt.Sprintf("%x", getHash(test.file))

		if h != test.hash {
			t.Errorf("getHash(%s)=%s;\nout %s", test.file, h, test.hash)
		}
	}
}
