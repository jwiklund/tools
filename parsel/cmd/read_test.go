package cmd

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestReadRange(t *testing.T) {
	str := "2017-02-13T08:00:00Z\n2017-02-13T09:00:00Z\n2017-02-13T10:00:00Z"

	before, _ := time.Parse(time.RFC3339, "2017-02-13T09:00:00Z")
	after, _ := time.Parse(time.RFC3339, "2017-02-13T10:00:00Z")
	ch := make(chan rec, 2)
	readFile(strings.NewReader(str), "\t", before, after, ch)
	close(ch)

	res := fmt.Sprintf("%s", readAll(ch))
	if res != "2017-02-13T09:00:00Z" {
		t.Error("Expected 2017-02-13T09:00:00Z but got ", res)
	}
}

func TestReadRangeInfiniteBegining(t *testing.T) {
	str := "2017-02-13T09:00:00Z"

	before := time.Time{}
	after, _ := time.Parse(time.RFC3339, "2017-02-13T10:00:00Z")
	ch := make(chan rec, 2)
	readFile(strings.NewReader(str), "\t", before, after, ch)
	close(ch)

	res := fmt.Sprintf("%s", readAll(ch))
	if res != "2017-02-13T09:00:00Z" {
		t.Error("Expected 2017-02-13T09:00:00Z but got ", res)
	}
}

func TestReadRangeInfiniteEnding(t *testing.T) {
	str := "2017-02-13T09:00:00Z"

	before, _ := time.Parse(time.RFC3339, "2017-02-13T09:00:00Z")
	after := time.Time{}
	ch := make(chan rec, 2)
	readFile(strings.NewReader(str), "\t", before, after, ch)
	close(ch)

	res := fmt.Sprintf("%s", readAll(ch))
	if res != "2017-02-13T09:00:00Z" {
		t.Error("Expected 2017-02-13T09:00:00Z but got ", res)
	}
}

func readAll(ch chan rec) string {
	var dates []string
	for r := range ch {
		dates = append(dates, r.timestamp.Format(time.RFC3339))
	}
	return strings.Join(dates, ", ")
}
