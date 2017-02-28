package cmd

import (
	"bufio"
	"strings"
	"testing"
	"time"
)

func TestReadRange(t *testing.T) {
	str := "2017-02-13T08:00:00Z\n2017-02-13T09:00:00Z\n2017-02-13T10:00:00Z"

	from, _ := time.Parse(time.RFC3339, "2017-02-13T09:00:00Z")
	to, _ := time.Parse(time.RFC3339, "2017-02-13T10:00:00Z")
	r := reader{
		scanner:   bufio.NewScanner(strings.NewReader(str)),
		delimiter: '\t',
		from:      from,
		to:        to,
	}

	res := readAllDates(&r)
	if res != "2017-02-13T09:00:00Z" {
		t.Error("Expected 2017-02-13T09:00:00Z but got ", res)
	}
}

func TestReadRangeInfiniteBegining(t *testing.T) {
	str := "2017-02-13T09:00:00Z"

	from := time.Time{}
	to, _ := time.Parse(time.RFC3339, "2017-02-13T10:00:00Z")
	r := reader{
		scanner:   bufio.NewScanner(strings.NewReader(str)),
		delimiter: '\t',
		from:      from,
		to:        to,
	}

	res := readAllDates(&r)
	if res != "2017-02-13T09:00:00Z" {
		t.Error("Expected 2017-02-13T09:00:00Z but got ", res)
	}
}

func TestReadRangeInfiniteEnding(t *testing.T) {
	str := "2017-02-13T09:00:00Z"

	from, _ := time.Parse(time.RFC3339, "2017-02-13T09:00:00Z")
	to := time.Time{}
	r := reader{
		scanner:   bufio.NewScanner(strings.NewReader(str)),
		delimiter: '\t',
		from:      from,
		to:        to,
	}

	res := readAllDates(&r)
	if res != "2017-02-13T09:00:00Z" {
		t.Error("Expected 2017-02-13T09:00:00Z but got ", res)
	}
}

func TestReadRangeDefaults(t *testing.T) {
	str := "2017-02-13T09:00:00Z"

	r := reader{
		scanner:   bufio.NewScanner(strings.NewReader(str)),
		delimiter: '\t',
	}

	res := readAllDates(&r)
	if res != "2017-02-13T09:00:00Z" {
		t.Error("Expected 2017-02-13T09:00:00Z but got ", res)
	}
}

func readAllDates(r *reader) string {
	var dates []string
	for r.Read() {
		dates = append(dates, r.rec.timestamp.Format(time.RFC3339))
	}
	return strings.Join(dates, ", ")
}

func TestReadFields(t *testing.T) {
	str := "2017-02-13T09:00:00Z\tfirst\tsecond"

	r := reader{
		scanner:   bufio.NewScanner(strings.NewReader(str)),
		delimiter: '\t',
	}

	res := readAllFields(&r)
	if res != "first, second" {
		t.Error("Expected first, second but got " + res)
	}
}

func readAllFields(r *reader) string {
	var fields []string
	for r.Read() {
		for _, record := range r.rec.records {
			fields = append(fields, string(record))
		}
	}
	return strings.Join(fields, ", ")
}
