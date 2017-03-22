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
	testFilter(t, "a line", "1:a", true)
	testFilter(t, "a line", "2:a", false)
	testFilter(t, "a line", "1:line", false)
	testFilter(t, "a line", "2:line", true)
	testFilter(t, "a line", "3:out of bounds", false)
}

func TestFilterFieldLess(t *testing.T) {
	testFilter(t, "10", "1:<9", false)
	testFilter(t, "10", "1:<10", false)
	testFilter(t, "10", "1:<10.1", true)
	testFilter(t, "10", "1:<11", true)
	testFilter(t, "b", "1:<a", false)
	testFilter(t, "b", "1:<b", false)
	testFilter(t, "b", "1:<c", true)
}

func TestFilterFieldMore(t *testing.T) {
	testFilter(t, "10", "1:>9", true)
	testFilter(t, "10.1", "1:>10", true)
	testFilter(t, "10", "1:>10", false)
	testFilter(t, "10", "1:>11", false)
	testFilter(t, "b", "1:>a", true)
	testFilter(t, "b", "1:>b", false)
	testFilter(t, "b", "1:>c", false)
}

func TestFilterLast(t *testing.T) {
	testFilter(t, "0 1", "-1:1", true)
	testFilter(t, "0 1", "-1:0", false)
	testFilter(t, "0 1", "-1:<2", true)
	testFilter(t, "0 1", "-1:<1", false)
	testFilter(t, "0 1", "-1:>0", true)
	testFilter(t, "0 1", "-1:>1", false)
}

func TestFilterContains(t *testing.T) {
	testFilter(t, "a line", "2:in", true)
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
