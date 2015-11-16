package sharder

import (
	"bytes"
	"strings"
	"testing"

	"appengine/aetest"
)

func TestFullCircle(t *testing.T) {
	// TODO: verify 20 shards

	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	data := []byte(strings.Repeat("01234567890123456789", 1e6))

	// Writing
	err = Writer(c, "test", data)
	if err != nil {
		t.Fatal(err)
	}

	// Reading
	read, err := Reader(c, "test")
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(read, data) {
		t.Fail()
	}
}

var test bool

func BenchmarkFullCycle(b *testing.B) {
	c, _ := aetest.NewContext(nil)
	defer c.Close()
	data := []byte(strings.Repeat("1", 1e6))

	for i := 0; i < b.N; i++ {
		Writer(c, "test", data)
		read, _ := Reader(c, "test")
		test = bytes.Equal(read, data)
	}
}
