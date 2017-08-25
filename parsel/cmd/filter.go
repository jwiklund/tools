package cmd

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type filterFn func(rec) bool

func (f filterFn) and(s filterFn) filterFn {
	return func(r rec) bool {
		return f(r) && s(r)
	}
}

func parseFilters(verbose bool, delimiter string, filters []string) (filterFn, error) {
	fn := filterFn(func(_ rec) bool {
		return true
	})
	for _, filter := range filters {
		nfn, err := parseFilter(verbose, delimiter, filter)
		if err != nil {
			return nil, err
		}
		fn = fn.and(nfn)
	}

	return fn, nil
}

func parseFilter(verbose bool, delimiter, filter string) (filterFn, error) {
	colon := strings.Index(filter, ":")
	if colon < 0 {
		return maybeNot(verbose, delimiter, filter, filterContains), nil
	}
	if colon == len(filter)-1 {
		return nil, errors.Errorf("missing filter for field %s", filter)
	}
	field, err := strconv.Atoi(filter[0:colon])
	if err != nil {
		return nil, fmt.Errorf("could not parse field index %s: %v", filter[0:colon], err)
	}
	if field == 0 {
		return nil, fmt.Errorf("Invalid index, 0 is for date and is non filterable")
	} else if field > 0 {
		field = field - 1
	}
	fieldFilter := filter[colon+1:]
	return maybeNot(verbose, delimiter, fieldFilter, func(v bool, d string, f string) filterFn {
		if f[0] == '<' {
			return filterLess(verbose, field, f[1:])
		} else if f[0] == '>' {
			return filterMore(verbose, field, f[1:])
		}
		return filterField(verbose, field, f)
	}), nil
}

func maybeNot(verbose bool, delimiter string, filter string, filterCreator func(bool, string, string) filterFn) filterFn {
	not := false
	if filter[0] == '!' {
		not = true
		filter = filter[1:]
	}
	fn := filterCreator(verbose, delimiter, filter)
	if not {
		return func(r rec) bool {
			res := !fn(r)
			if verbose {
				fmt.Println("filter.not", filter, ":", res)
			}
			return res
		}
	}
	return fn
}

func filterContains(verbose bool, delimiter string, filter string) filterFn {
	if filter[0] == '^' {
		filter = delimiter + filter[1:]
	}
	var original []byte
	if filter[len(filter)-1] == '$' {
		original = []byte(filter[0 : len(filter)-1])
		filter = string(original) + delimiter
	}
	find := []byte(filter)
	return func(r rec) bool {
		res := bytes.Contains(r.line, find)
		if !res && original != nil {
			if len(r.line) >= len(original) {
				res = bytes.Equal(r.line[len(r.line)-len(original):], original)
			}
		}
		if verbose {
			fmt.Println("filter: ", string(filter), res)
		}
		return res
	}
}

func filterMore(verbose bool, field int, filter string) filterFn {
	filterNumber, err := strconv.ParseFloat(filter, 32)
	if err == nil {
		return func(r rec) bool {
			fieldIndex := field
			if fieldIndex < 0 {
				fieldIndex = len(r.records) + field
			}
			if (fieldIndex < 0) || (len(r.records) <= fieldIndex) {
				if verbose {
					fmt.Println("filter.moreField:", field, filter, "too few records")
				}
				return false
			}
			value, err := strconv.ParseFloat(string(r.records[fieldIndex]), 32)
			if err != nil {
				if verbose {
					fmt.Println("filter.moreField:", field, filter, "not a number")
				}
				return false
			}
			res := value > filterNumber
			if verbose {
				fmt.Println("filter.moreField:", field, value, "<", filter, res)
			}
			return res
		}
	} else if verbose {
		fmt.Println("filter.moreField", filter, "not a number")
	}
	filterBytes := []byte(filter)
	return func(r rec) bool {
		fieldIndex := field
		if fieldIndex < 0 {
			fieldIndex = len(r.records) + field
		}
		if (fieldIndex < 0) || (len(r.records) < fieldIndex) {
			if verbose {
				fmt.Println("filter.moreField:", field, filter, "too few records")
			}
			return false
		}
		res := bytes.Compare(r.records[fieldIndex], filterBytes) > 0
		if verbose {
			fmt.Println("filter.moreField:", field, string(r.records[field]), "<", filter, res)
		}
		return res
	}
}

func filterLess(verbose bool, field int, filter string) filterFn {
	filterNumber, err := strconv.ParseFloat(filter, 32)
	if err == nil {
		return func(r rec) bool {
			fieldIndex := field
			if fieldIndex < 0 {
				fieldIndex = len(r.records) + field
			}
			if (fieldIndex < 0) || (len(r.records) <= fieldIndex) {
				if verbose {
					fmt.Println("filter.lessField:", field, filter, "too few records")
				}
				return false
			}
			value, err := strconv.ParseFloat(string(r.records[fieldIndex]), 32)
			if err != nil {
				if verbose {
					fmt.Println("filter.lessField:", field, filter, "not a number")
				}
				return false
			}
			res := value < filterNumber
			if verbose {
				fmt.Println("filter.lessField:", field, value, "<", filter, res)
			}
			return res
		}
	}
	filterBytes := []byte(filter)
	return func(r rec) bool {
		fieldIndex := field
		if fieldIndex < 0 {
			fieldIndex = len(r.records) + field
		}
		if (fieldIndex < 0) || (len(r.records) < fieldIndex) {
			if verbose {
				fmt.Println("filter.lessField:", field, filter, "too few records")
			}
			return false
		}
		res := bytes.Compare(r.records[fieldIndex], filterBytes) < 0
		if verbose {
			fmt.Println("filter.lessField:", field, string(r.records[fieldIndex]), "<", filter, res)
		}
		return res
	}
}

func filterField(verbose bool, field int, filter string) filterFn {
	var compareFn func([]byte) bool

	if filter[0] == '^' {
		filterBytes := []byte(filter[1:])
		filterLen := len(filterBytes)
		compareFn = func(bs []byte) bool {
			if len(bs) < filterLen {
				return false
			}
			return bytes.Equal(bs[0:filterLen], filterBytes)
		}
	} else if filter[len(filter)-1] == '$' {
		filterBytes := []byte(filter[0 : len(filter)-1])
		filterLen := len(filterBytes)
		compareFn = func(bs []byte) bool {
			if len(bs) < filterLen {
				return false
			}
			return bytes.Equal(bs[len(bs)-filterLen:], filterBytes)
		}
	} else {
		filterBytes := []byte(filter)
		compareFn = func(bs []byte) bool {
			return bytes.Contains(bs, filterBytes)
		}
	}

	return func(r rec) bool {
		fieldIndex := field
		if fieldIndex < 0 {
			fieldIndex = len(r.records) + field
		}
		if (fieldIndex < 0) || (len(r.records) <= fieldIndex) {
			if verbose {
				fmt.Println("filter.field:", field, filter, "too few records")
			}
			return false
		}
		res := compareFn(r.records[fieldIndex])
		if verbose {
			fmt.Println("filter.field:", field, filter, "contains", string(r.records[fieldIndex]), res)
		}
		return res
	}
}
