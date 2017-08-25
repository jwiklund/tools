package cmd

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestFilterResultIndex(t *testing.T) {
	fields, err := parseFields("5")
	if err != nil {
		t.Fatal("Invalid field", err)
	}
	filter, err := parseFilter(false, "  ", "5:5")
	if err != nil {
		t.Fatal("Invalid filter", err)
	}
	r, err := newReader(strings.NewReader("2006-01-02T15:04:05Z 1 2 3 4 5"), " ", time.Time{}, time.Time{})
	if err != nil {
		t.Fatal("Invalid reader", err)
	}
	if !r.Read() {
		t.Fatal("Could not read line")
	}
	if !filter(r.rec) {
		t.Error("Filter didn't match")
	}
	b := bytes.Buffer{}
	buf := bufio.NewWriter(&b)
	result(r.rec, " ", fields, buf)
	buf.Flush()

	if strings.TrimSpace(b.String()) != "5" {
		t.Error("Expected 5 but got", b.String())
	}
}
