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

func parseFilters(verbose bool, filters []string) (filterFn, error) {
	fn := filterFn(func(_ rec) bool {
		return true
	})
	for _, filter := range filters {
		nfn, err := parseFilter(verbose, filter)
		if err != nil {
			return nil, err
		}
		fn = fn.and(nfn)
	}

	return fn, nil
}

func parseFilter(verbose bool, filter string) (filterFn, error) {
	colon := strings.Index(filter, ":")
	if colon < 0 {
		return filterContains(verbose, []byte(filter)), nil
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
	if fieldFilter[0] == '<' {
		return filterLess(verbose, field, fieldFilter[1:]), nil
	} else if fieldFilter[0] == '>' {
		return filterMore(verbose, field, fieldFilter[1:]), nil
	}
	return filterField(verbose, field, fieldFilter), nil
}

func filterContains(verbose bool, filter []byte) filterFn {
	return func(r rec) bool {
		res := bytes.Contains(r.line, filter)
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
			if (fieldIndex < 0) || (len(r.records) < fieldIndex) {
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
			if (fieldIndex < 0) || (len(r.records) < fieldIndex) {
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
	filterBytes := []byte(filter)
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
		res := bytes.Contains(r.records[fieldIndex], filterBytes)
		if verbose {
			fmt.Println("filter.field:", field, filter, "contains", string(r.records[fieldIndex]), res)
		}
		return res
	}
}
