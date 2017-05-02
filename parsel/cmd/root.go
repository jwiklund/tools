package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

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

var fromRaw *string
var toRaw *string
var delimiter *string
var fieldsRaw *string
var cpuprofile *string
var filtersRaw *[]string
var preview *bool
var verbose *bool

func init() {
	fromRaw = RootCmd.Flags().StringP("from", "F", "", "Only include items from this time")
	toRaw = RootCmd.Flags().StringP("to", "T", "", "Only include items until this time")
	delimiter = RootCmd.Flags().StringP("delimiter", "d", "\t", "Field Delimiter")
	fieldsRaw = RootCmd.Flags().StringP("fields", "f", "", "Only return fields (eg 1,2,3-4)")
	filtersRaw = RootCmd.Flags().StringArray("filter", []string{}, "Filtering to perform")
	cpuprofile = RootCmd.Flags().String("cpuprofile", "", "Write cpuprofile to file")
	preview = RootCmd.Flags().Bool("preview", false, "Preview the result, only return 10 rows")
	verbose = RootCmd.Flags().BoolP("verbose", "v", false, "Be verbose")
}

type rec struct {
	timestamp time.Time
	line      []byte
	records   [][]byte
}

func root(cmd *cobra.Command, args []string) {
	if *cpuprofile != "" {
		if *verbose {
			fmt.Println("Storing cpu profile in " + *cpuprofile)
		}
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	now := time.Now()
	parseDuration := func(what, value string) time.Time {
		if value == "" {
			return time.Time{}
		}

		duration, err := time.ParseDuration(value)
		if err == nil {
			return now.Add(-duration)
		}

		t, err := time.Parse(time.RFC3339, value)
		if err == nil {
			return t
		}

		fmt.Printf("Invalid %s %s duration|time (must be of type .*(ns|us|ms|s|m|h) or RFC3339 ie 2017-02-13T09:16:57Z)\n", what, value)
		os.Exit(1)
		return now
	}
	from := parseDuration("from", *fromRaw)
	to := parseDuration("to", *toRaw)

	if *verbose {
		fmt.Printf("Return records between %s and %s\n", from, to)
	}

	fields, err := parseFields(*fieldsRaw)
	if err != nil {
		fmt.Println("Invalid fields")
		os.Exit(1)
	}
	filter, err := parseFilters(*verbose, *filtersRaw)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	output := bufio.NewWriter(os.Stdout)
	for _, file := range args {
		var r *reader
		var err error
		if file == "-" {
			r, err = newReader(os.Stdin, *delimiter, from, to)
		} else {
			r, err = newReaderFile(file, *delimiter, from, to)
		}
		if err != nil {
			fmt.Println(err)
			continue
		}
		first := true
		var firstTime, lastTime time.Time
		var count int
		for r.Read() {
			if !filter(r.rec) {
				continue
			}
			if first {
				first = false
				firstTime = r.rec.timestamp
			}
			if *preview {
				if count == 0 {
					printFieldIndexes(r.rec, *delimiter, fields, output)
				}
				if count > 10 {
					break
				}
				count = count + 1
			}
			lastTime = r.rec.timestamp
			result(r.rec, *delimiter, fields, output)
		}
		if *verbose {
			fmt.Printf("file %s time %s to %s\n", file, firstTime, lastTime)
		}
	}
	output.Flush()
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

func newReaderFile(file, delimiter string, from, to time.Time) (*reader, error) {
	f, err := os.OpenFile(file, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("coult not open %s: %s", file, err)
	}
	return newReader(f, delimiter, from, to)
}

func newReader(r io.Reader, delimiter string, from, to time.Time) (*reader, error) {
	if len(delimiter) > 1 {
		return nil, fmt.Errorf("delimiter of size != 1 not supported")
	}
	return &reader{
		scanner:   bufio.NewScanner(r),
		delimiter: delimiter[0],
		from:      from,
		to:        to,
	}, nil
}

type reader struct {
	scanner   *bufio.Scanner
	rec       rec
	delimiter byte
	from      time.Time
	to        time.Time
}

func (r *reader) Read() bool {
	more, found := r.readInternal()
	for more && !found {
		more, found = r.readInternal()
	}
	return found
}

var zero time.Time

func (r *reader) readInternal() (bool, bool) {
	if !r.scanner.Scan() {
		return false, false
	}
	if len(r.scanner.Bytes()) == 0 {
		return true, false
	}
	err := parse(r.delimiter, r.scanner.Bytes(), &r.rec)
	if err != nil {
		fmt.Printf("Could not parse line %s: %s\n", r.scanner.Text(), err)
		return true, false
	}
	if !r.from.After(r.rec.timestamp) || r.from == zero {
		if r.rec.timestamp.Before(r.to) || r.to == zero {
			return true, true
		}
	}
	return true, false
}

func parse(delimiter byte, line []byte, rec *rec) error {
	var last int
	var err error
	last, rec.timestamp, err = parseTime(delimiter, line)
	if err != nil {
		return err
	}
	rec.line = line
	recs := rec.records[0:0]
	current := 1
	for last+current < len(line) {
		if line[last+current] == delimiter {
			recs = append(recs, line[last:last+current])
			last = last + current + 1
			current = 0
		}
		current = current + 1
	}
	if current > 0 {
		recs = append(recs, line[last:last+current])
	}
	rec.records = recs
	return nil
}

func parseTime(delimiter byte, line []byte) (int, time.Time, error) {
	space := 0
	for space < len(line) && line[space] != ' ' && line[space] != '\t' && line[space] != delimiter {
		space = space + 1
	}
	date, err := time.Parse(time.RFC3339, string(line[0:space]))
	if err != nil {
		return 0, time.Time{}, err
	}
	return space + 1, date, nil
}

func result(r rec, delimiter string, fields []int, out *bufio.Writer) {
	if len(fields) == 0 {
		out.WriteString(r.timestamp.Format(time.RFC3339))
		for _, record := range r.records {
			out.WriteString(delimiter)
			out.Write(record)
		}
	} else {
		first := true
		for _, field := range fields {
			// field 0 == timestamp
			fieldIndex := field - 1
			if fieldIndex < len(r.records) {
				if first {
					first = false
				} else {
					out.WriteString(delimiter)
				}
				if fieldIndex == -1 {
					out.WriteString(r.timestamp.Format(time.RFC3339))
				}
				if fieldIndex < 0 {
					negativeRewrite := len(r.records) + fieldIndex + 1
					if negativeRewrite >= 0 && negativeRewrite < len(r.records) {
						out.Write(r.records[negativeRewrite])
					}
				} else {
					out.Write(r.records[fieldIndex])
				}
			}
		}
	}
	out.WriteString("\n")
}

func printFieldIndexes(r rec, delimiter string, fields []int, out *bufio.Writer) {
	if len(fields) == 0 {
		out.WriteString(fmt.Sprintf("%3d\t%s\n", 0, r.timestamp.Format(time.RFC3339)))
		for i, rec := range r.records {
			out.WriteString(fmt.Sprintf("%3d\t%s\n", i+1, rec))
		}
	} else {
		for _, field := range fields {
			fieldIndex := field - 1
			if fieldIndex < len(r.records) {
				if fieldIndex == -1 {
					out.WriteString(fmt.Sprintf("%3d\t%s\n", field, r.timestamp.Format(time.RFC3339)))
				} else if fieldIndex < 0 {
					negativeRewrite := len(r.records) + fieldIndex + 1
					if negativeRewrite >= 0 && negativeRewrite < len(r.records) {
						out.WriteString(fmt.Sprintf("%3d\t%s\n", field, r.records[negativeRewrite]))
					} else {
						out.WriteString(fmt.Sprintf("%3d\tOut of range\n", field))
					}
				} else {
					out.WriteString(fmt.Sprintf("%3d\t%s\n", field, r.records[fieldIndex]))
				}
			} else {
				out.WriteString(fmt.Sprintf("%3d\tOut of range\n", field))
			}
		}
	}
	out.WriteString("\n")
}
