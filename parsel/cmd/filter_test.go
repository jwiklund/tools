package cmd

import (
	"testing"
)

var debug = false

func TestFilterLine(t *testing.T) {
	testFilter(t, "a line", "a line", true)
	testFilter(t, "a line", "a", true)
	testFilter(t, "a line", "b", false)
}

func TestFilterField(t *testing.T) {
	testFilter(t, "a line", "0:a", true)
	testFilter(t, "a line", "1:a", false)
	testFilter(t, "a line", "0:line", false)
	testFilter(t, "a line", "1:line", true)
	testFilter(t, "a line", "2:out of bounds", false)
}

func TestFilterFieldLess(t *testing.T) {
	testFilter(t, "10", "0:<9", false)
	testFilter(t, "10", "0:<10", false)
	testFilter(t, "10", "0:<10.1", true)
	testFilter(t, "10", "0:<11", true)
	testFilter(t, "b", "0:<a", false)
	testFilter(t, "b", "0:<b", false)
	testFilter(t, "b", "0:<c", true)
}

func TestFilterFieldMore(t *testing.T) {
	testFilter(t, "10", "0:>9", true)
	testFilter(t, "10.1", "0:>10", true)
	testFilter(t, "10", "0:>10", false)
	testFilter(t, "10", "0:>11", false)
	testFilter(t, "b", "0:>a", true)
	testFilter(t, "b", "0:>b", false)
	testFilter(t, "b", "0:>c", false)
}

func TestFilterLast(t *testing.T) {
	testFilter(t, "0 1", "-1:1", true)
	testFilter(t, "0 1", "-1:0", false)
	testFilter(t, "0 1", "-1:<2", true)
	testFilter(t, "0 1", "-1:<1", false)
	testFilter(t, "0 1", "-1:>0", true)
	testFilter(t, "0 1", "-1:>1", false)
}

func testFilter(t *testing.T, row string, filter string, expect bool) {
	line := []byte("2017-03-01T16:02:04Z " + row)
	var r rec
	if err := parse(' ', []byte(line), &r); err != nil {
		t.Fatal("could not parse line", err)
	}
	fn, err := parseFilter(debug, filter)
	if err != nil {
		t.Fatal("could not parse filter", err)
	}
	if fn(r) != expect {
		t.Error("expected matching", expect, "but got", fn(r), "for filter", filter, "and line", string(line))
	}
}
