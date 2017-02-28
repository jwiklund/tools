package cmd

import (
	"bufio"
	"fmt"
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

func init() {
	RootCmd.Flags().StringP("from", "F", "", "Only include items from this time")
	RootCmd.Flags().StringP("to", "T", "", "Only include items until this time")
	RootCmd.Flags().StringP("delimiter", "d", "\t", "Field Delimiter")
	RootCmd.Flags().StringP("fields", "f", "", "Only return fields (eg 1,2,3-4)")
	RootCmd.Flags().StringP("query", "q", "", "Only return lines matching query")
	RootCmd.Flags().BoolP("verbose", "v", false, "Be verbose")
	RootCmd.Flags().String("cpuprofile", "", "Write cpuprofile to file")
}

type rec struct {
	timestamp time.Time
	records   [][]byte
}

func root(cmd *cobra.Command, args []string) {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		fmt.Println("Couldn't parse verbose, assuming false", err)
		verbose = false
	}
	cpuprofile := cmd.Flag("cpuprofile").Value.String()
	if cpuprofile != "" {
		if verbose {
			fmt.Println("Storing cpu profile in " + cpuprofile)
		}
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
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

	output := bufio.NewWriter(os.Stdout)
	for _, file := range args {
		r, err := newReaderFile(file, delimiter, from, to)
		if err != nil {
			fmt.Println(err)
			continue
		}
		first := true
		var firstTime, lastTime time.Time
		for r.Read() {
			if first {
				first = false
				firstTime = r.rec.timestamp
			}
			lastTime = r.rec.timestamp
			result(r.rec, delimiter, fields, output)
		}
		if verbose {
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
	if len(delimiter) > 1 {
		return nil, fmt.Errorf("delimiter of size != 1 not supported")
	}
	return &reader{
		scanner:   bufio.NewScanner(f),
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
	space := 0
	for space < len(line) && line[space] != ' ' && line[space] != delimiter {
		space = space + 1
	}
	date, err := time.Parse(time.RFC3339, string(line[0:space]))
	if err != nil {
		return err
	}
	rec.timestamp = date

	recs := rec.records[0:0]
	last := space + 1
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

func result(r rec, delimiter string, fields []int, out *bufio.Writer) {
	out.WriteString(r.timestamp.Format(time.RFC3339))
	if len(fields) == 0 {
		for _, record := range r.records {
			out.WriteString(delimiter)
			out.Write(record)
		}
	} else {
		for _, field := range fields {
			if field < len(r.records) {
				out.WriteString(delimiter)
				out.Write(r.records[field])
			}
		}
	}
	out.WriteString("\n")
}