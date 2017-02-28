package cmd

import (
	"bufio"
	"fmt"
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
	RootCmd.Flags().StringP("from", "F", "10m", "Only include items from this time")
	RootCmd.Flags().StringP("to", "T", "0m", "Only include items until this time")
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
	fromDuration, err := cmd.Flags().GetDuration("from")
	var from time.Time
	if err == nil {
		from = now.Add(-fromDuration)
	} else {
		from, err = time.Parse(time.RFC3339, cmd.Flag("from").Value.String())
		if err != nil {
			fmt.Println("Invalid from duration|time (must be of type .*(ns|us|ms|s|m|h) or RFC3339 ie 2017-02-13T09:16:57.780+00:00)")
			os.Exit(1)
		}
	}

	toDuration, err := cmd.Flags().GetDuration("to")
	var to time.Time
	if err == nil {
		to = now.Add(-toDuration)
	} else {
		to, err = time.Parse(time.RFC3339, cmd.Flag("to").Value.String())
		if err != nil {
			fmt.Println("Invalid to duration|time (must be of type .*(ns|us|ms|s|m|h) or RFC3339 ie 2017-02-13T09:16:57.780+00:00)")
			os.Exit(1)
		}
	}

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
	res := make([]int, len(stringFields))
	for i, field := range stringFields {
		fieldNr, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("could not parse field %s: %v", field, err)
		}
		res[i] = fieldNr
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

	parts := bytes.Split(line[space+1:], delimiter)
	recs := make([]string, len(parts))

	for i, part := range parts {
		recs[i] = string(part)
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
	defer f.Close()

	scanner := bufio.NewScanner(f)
	d := []byte(delimiter)
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}
		r, err := parse(d, scanner.Bytes())
		if err != nil {
			fmt.Printf("Could not parse line %s", scanner.Text())
		} else {
			if from.Before(r.timestamp) {
				continue
			}
			if r.timestamp.After(to) {
				continue
			}
			output <- r
		}
	}
	close(output)
}

func result(r rec, delimiter string, fields []int) string {
	records := r.records
	if fields != nil {
		res := make([]string, 0, len(records))
		for _, field := range fields {
			if field < len(records) {
				res = append(res, records[field])
			}
		}
		records = res
	}
	return strings.Join(records, delimiter)
}
