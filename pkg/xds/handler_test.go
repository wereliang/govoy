package xds

import (
	"fmt"
	"testing"
)

func TestHandleXDS(t *testing.T) {
	handler := &handler{}
	// olds := []string{"a", "b", "c", "e", "m", "x", "y", "z"}
	olds := []string{"a", "b", "c", "e", "m"}
	news := []string{"a", "e", "f", "g", "x", "y"}
	handler.deleteXDS(
		olds, news,
		func(x string) {
			fmt.Println("delete", x)
		})
}
