package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/hashicorp/go-hclog"
	"gopkg.in/cheggaaa/pb.v2"

	"github.com/mwmahlberg/mgogenerate/generate"
	"go.mongodb.org/mongo-driver/mongo"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const customProgress pb.ProgressBarTemplate = `{{string . "prefix"}}{{counters . }} {{bar . }} {{percent . }} {{speed . "%s docs/sec"}} {{rtime . "ETA %s"}}{{string . "suffix"}}`

var (
	app     = kingpin.New("mgogenerate", "Go implementation of a document generator for BSON documents")
	numDocs int
	numOps  int
	runners int
	host    string
	port    uint16

	db         string
	collection string

	tmplFileName string

	logger hclog.Logger
	debug  bool
)

func init() {
	app.Version("1.0.1")
	app.Flag("runners", "number of concurrent generators").Short('r').Default("2").IntVar(&runners)
	app.Flag("number", "number of documents to generate").Short('n').Default("1").IntVar(&numDocs)
	app.Flag("ops", "number of inserts per bulk operation").Default("1000").Short('o').IntVar(&numOps)
	app.Flag("host", "host to connect to").Default("127.0.0.1").Short('h').StringVar(&host)
	app.Flag("port", "port to connect to").Short('p').Default("27017").Uint16Var(&port)
	app.Flag("db", "database to use").Short('d').Default("test").StringVar(&db)
	app.Flag("collection", "collection to use").Short('c').Default("mgogenerate").StringVar(&collection)
	app.Flag("debug", "activate debug logging").Default("false").BoolVar(&debug)
	app.Arg("file.json", "template file to be used").Default("template.json").StringVar(&tmplFileName)

	logger = hclog.Default()
}

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if debug {
		logger.SetLevel(hclog.Debug)
	}

	var (
		err      error
		tmplFile *os.File
		tmpl     *template.Template
	)

	if tmplFile, err = setupFile(tmplFileName); err != nil {
		logger.Error("opening template file", "path", tmplFileName, "error", err)
		os.Exit(1)
	}
	defer tmplFile.Close()

	if tmpl, err = setupTemplate(tmplFile); err != nil {
		logger.Error("setting up template", "file", tmplFile, "error", err)
		os.Exit(3)
	}

	/* Sanitize the input */
	if numDocs < numOps {
		logger.Warn("Adjusting number of operations for bulk commit", "documents", numDocs, "operations", numOps)
		numOps = numDocs
		logger.Warn("Adjusted the number of operations for bulk commit", "numOps", numOps)
	}
	logger.Debug("Bulk operations commit interval", "operations", numOps)

	mongoURL := fmt.Sprintf("mongodb://%s:%d", host, port)

	outs := make([]chan *mongo.InsertOneModel, 0)

	chunkSize := numDocs / runners
	remainder := numDocs % chunkSize
	for r := 1; r <= runners; r++ {
		if r == runners {
			chunkSize = chunkSize + remainder
		}
		logger.Debug("Spinning up pump", "number", r, "docs", chunkSize)
		outs = append(outs, pump(chunkSize, tmpl, logger.Named("pump").With("pump#", r)))
	}

	a := aggregate(logger.Named("aggregate"), outs...)

	progress := setUpProgressBar()
	d, err := mongoSink(a, mongoURL, int(numOps), progress, logger.Named("sink"))

	logger.Debug("Finished generating document",
		"time elapsed", d,
		"documents generated", numDocs,
		"average/document", time.Duration((*d).Nanoseconds()/int64(numDocs)))
}

func setUpProgressBar() *pb.ProgressBar {
	progress := pb.New(int(numDocs))
	progress.SetTemplate(customProgress)
	progress.Set("prefix", "Documents written: ")
	return progress
}

func setupFile(tmplFileName string) (tmplFile *os.File, err error) {
	/*
	 * Check whether the input file exists
	 */
	if _, err = os.Stat(tmplFileName); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file does not exist: %s", err)
	} else if err != nil {
		return nil, fmt.Errorf("accessing template file: %s", err)
	}

	if tmplFile, err = os.Open(tmplFileName); err != nil {

		switch {
		case os.IsPermission(err):
			return nil, fmt.Errorf("opening template file '%s': insufficient persmissions: %s", tmplFileName, err)
		default:
			return nil, fmt.Errorf("opening template file '%s': %s", tmplFileName, err)
		}
	}

	return
}

func setupTemplate(tmplData io.Reader) (*template.Template, error) {

	var err error

	buf := bytes.NewBuffer(nil)

	if _, err := io.Copy(buf, tmplData); err != nil {
		return nil, fmt.Errorf("Error reading template data: %s", err)
	}

	tmpl := template.New("input")
	fm := template.FuncMap{
		"randDecimal": generate.Decimal,
		"randInteger": generate.Integer,
		"isoDate":     generate.ISODate,
		"objectId":    generate.ObjectId,
		"oneOf":       generate.OneOf,
		"N":           generate.N,
	}

	for k, v := range sprig.FuncMap() {
		fm[k] = v
	}

	tmpl.Funcs(fm)

	if tmpl, err = tmpl.Parse(buf.String()); err != nil {
		return nil, fmt.Errorf("Error parsing template: %s", err)
	}

	rand.Seed(time.Now().UnixNano())

	return tmpl, nil
}
