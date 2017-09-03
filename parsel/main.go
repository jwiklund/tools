package main

import (
	"github.com/jwiklund/tools/parsel/cmd"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	fromRaw    = kingpin.Flag("from", "Only include items from this time").Short('F').String()
	toRaw      = kingpin.Flag("to", "Only include items until this time").Short('T').String()
	delimiter  = kingpin.Flag("delimiter", "Field delimiter").Short('d').String()
	fieldsRaw  = kingpin.Flag("fields", "Only return fields (eg 1,2,3-4)").Short('f').String()
	filtersRaw = kingpin.Flag("filter", "Filtering to perform").Strings()
	cpuprofile = kingpin.Flag("cpuprofile", "Write cpuprofile to file").String()
	preview    = kingpin.Flag("preview", "Preview the result, only return 10 rows").Short('p').Bool()
	verbose    = kingpin.Flag("verbose", "Be verbose").Short('v').Bool()
	files      = kingpin.Flag("files", "Files to read").Required().Strings()
)

func main() {
	kingpin.Parse()

	cmd.FromRaw = fromRaw
	cmd.ToRaw = toRaw
	cmd.Delimiter = delimiter
	cmd.FieldsRaw = fieldsRaw
	cmd.FiltersRaw = filtersRaw
	cmd.Cpuprofile = cpuprofile
	cmd.Preview = preview
	cmd.Verbose = verbose

	cmd.Parsel(*files)
}
