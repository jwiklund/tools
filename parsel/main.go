package main

import (
	"os"

	"github.com/jwiklund/tools/parsel/cmd"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app        = kingpin.New("filter", "Filter logs")
	fromRaw    = app.Flag("from", "Only include items from this time").Short('F').String()
	toRaw      = app.Flag("to", "Only include items until this time").Short('T').String()
	delimiter  = app.Flag("delimiter", "Field delimiter").Default("\t").Short('d').String()
	fieldsRaw  = app.Flag("fields", "Only return fields (eg 1,2,3-4)").Short('f').String()
	filtersRaw = app.Flag("filter", "Filtering to perform").Strings()
	cpuprofile = app.Flag("cpuprofile", "Write cpuprofile to file").String()
	preview    = app.Flag("preview", "Preview the result, only return 10 rows").Short('p').Bool()
	verbose    = app.Flag("verbose", "Be verbose").Short('v').Bool()
	files      = app.Arg("files", "Files to read").Required().Strings()
)

func main() {
	app.HelpFlag.Short('h')

	kingpin.MustParse(app.Parse(os.Args[1:]))

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
