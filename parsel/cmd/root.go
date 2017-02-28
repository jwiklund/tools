package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"bytes"

	"github.com/spf13/cobra"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "parsel",
	Short: "Parse and search logs",
	Long:  `Parse and search logs`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: root,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.Flags().StringP("from", "F", "", "Only include items from this time")
	RootCmd.Flags().StringP("to", "T", "", "Only include items until this time")
	RootCmd.Flags().StringP("delimiter", "d", "\t", "Field Delimiter")
	RootCmd.Flags().StringP("fields", "f", "", "Only return fields (eg 1,2,3-4)")
	RootCmd.Flags().StringP("query", "q", "", "Only return lines matching query")
	RootCmd.Flags().BoolP("verbose", "v", false, "Be verbose")
}

type rec struct {
	timestamp time.Time
	records   []string
}

func root(cmd *cobra.Command, args []string) {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		fmt.Println("Couldn't parse verbose, assuming false", err)
		verbose = false
	}

	now := time.Now()
	parseDuration := func(what string) time.Time {
		str := cmd.Flag(what).Value.String()
		if str == "" {
			return time.Time{}
		}

		duration, err := time.ParseDuration(str)
		if err == nil {
			return now.Add(-duration)
		}

		t, err := time.Parse(time.RFC3339, str)
		if err == nil {
			return t
		}

		fmt.Printf("Invalid %s %s duration|time (must be of type .*(ns|us|ms|s|m|h) or RFC3339 ie 2017-02-13T09:16:57Z)\n", what, str)
		os.Exit(1)
		return now
	}
	from := parseDuration("from")
	to := parseDuration("to")

	if verbose {
		fmt.Printf("Return records between %s and %s\n", from, to)
	}

	delimiter := cmd.Flag("delimiter").Value.String()
	fields, err := parseFields(cmd.Flag("fields").Value.String())
	if err != nil {
		fmt.Println("Invalid fields")
		os.Exit(1)
	}

	for _, file := range args {
		reader := make(chan rec, 1)
		go read(file, delimiter, from, to, reader)
		first := true
		var firstTime, lastTime time.Time
		for rec := range reader {
			if first {
				first = false
				firstTime = rec.timestamp
			}
			lastTime = rec.timestamp
			fmt.Println(result(rec, delimiter, fields))
		}
		if verbose {
			fmt.Printf("file %s time %s to %s\n", file, firstTime, lastTime)
		}
	}
}

func parseFields(fields string) ([]int, error) {
	if fields == "" {
		return nil, nil
	}
	stringFields := strings.Split(fields, ",")
	res := make([]int, 0, len(stringFields))
	for _, field := range stringFields {
		fieldNr, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("could not parse field %s: %v", field, err)
		}
		res = append(res, fieldNr)
	}
	return res, nil
}

func parse(delimiter []byte, line []byte) (rec, error) {
	space := 0
	for space < len(line) && line[space] != ' ' {
		space = space + 1
	}
	date, err := time.Parse(time.RFC3339, string(line[0:space]))
	if err != nil {
		return rec{}, err
	}

	var recs []string
	if space+1 < len(line) {
		parts := bytes.Split(line[space+1:], delimiter)
		recs = make([]string, len(parts))

		for i, part := range parts {
			recs[i] = string(part)
		}
	}

	return rec{
		timestamp: date,
		records:   recs,
	}, nil
}

func read(file, delimiter string, from, to time.Time, output chan rec) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		fmt.Printf("Could not open %s: %s\n", file, err)
		return
	}

	readFile(f, delimiter, from, to, output)

	f.Close()
	close(output)
}

func readFile(in io.Reader, delimiter string, from, to time.Time, output chan rec) {
	scanner := bufio.NewScanner(in)
	d := []byte(delimiter)
	zero := time.Time{}
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}
		r, err := parse(d, scanner.Bytes())
		if err != nil {
			fmt.Printf("Could not parse line %s: %s\n", scanner.Text(), err)
		} else {
			if !from.After(r.timestamp) || from == zero {
				if r.timestamp.Before(to) || to == zero {
					output <- r
				}
			}
		}
	}
}

func result(r rec, delimiter string, fields []int) string {
	records := r.records
	if len(fields) > 0 {
		res := make([]string, 0, len(records))
		for _, field := range fields {
			if field < len(records) {
				res = append(res, records[field])
			}
		}
		records = res
	}
	return r.timestamp.Format(time.RFC3339) + delimiter + strings.Join(records, delimiter)
}
