package main

import (
	"os"

	"github.com/jwiklund/tools/parsel/cmd"
	"gopkg.in/alecthomas/kingpin.v2"
)

var ()

func main() {
	var args cmd.Args
	app := kingpin.New("filter", "Filter logs")
	app.Flag("from", "Only include items from this time").Short('F').StringVar(&args.From)
	app.Flag("to", "Only include items until this time").Short('T').StringVar(&args.To)
	app.Flag("delimiter", "Field delimiter").Default("\t").Short('d').StringVar(&args.Delimiter)
	app.Flag("fields", "Only return fields (eg 1,2,3-4)").Short('f').StringVar(&args.Fields)
	app.Flag("filter", "Filtering to perform").StringsVar(&args.Filters)
	app.Flag("cpuprofile", "Write cpuprofile to file").StringVar(&args.Cpuprofile)
	app.Flag("preview", "Preview the result, only return 10 rows").Short('p').BoolVar(&args.Preview)
	app.Flag("verbose", "Be verbose").Short('v').BoolVar(&args.Verbose)
	app.Arg("files", "Files to read (stdin for stdin)").Required().StringsVar(&args.Args)
	app.HelpFlag.Short('h')

	kingpin.MustParse(app.Parse(os.Args[1:]))
	cmd.Parsel(&args)
}
