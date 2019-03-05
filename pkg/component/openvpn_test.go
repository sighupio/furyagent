package component

import (
	"fmt"
	"testing"
)

func TestGenerateTaKey(t *testing.T) {
	data, err := getTaKey()
	if err != nil {
		t.Fail()
	}
	fmt.Println(string(data))
}
