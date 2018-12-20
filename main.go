package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/araddon/dateparse"
	"github.com/mitchellh/mapstructure"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var numDocs int
var app = kingpin.New("mgogenerate", "Go implementation of a document generator for BSON documents")
var tmplFileName string
var tmplFile *os.File

// Generator is an interface defining the types used for Random generation of Values.
type Generator interface {
	Generate() interface{}
}

type Decimal struct {
	Min   float64
	Max   float64
	Fixed int
}

func (d *Decimal) Generate() interface{} {
	r := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	f, _ := strconv.ParseFloat(fmt.Sprintf("%.*f", d.Fixed, r.Float64()*(d.Max-d.Min)+d.Min), 0)
	return f
}

type ISODate struct {
	Min time.Time
	Max time.Time
}

func (i *ISODate) Generate() interface{} {
	r := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	log.Println(i.Min)
	log.Println(i.Max)
	log.Println(i.Min.Location().String())
	min := i.Min.UnixNano()
	max := i.Max.UnixNano()
	s := r.Int63n(max-min+1) + min
	log.Println(min, max, max-min, s)
	return time.Unix(0, s)
}

func init() {
	app.Flag("number", "number of documents to generate").Short('n').Default("1").IntVar(&numDocs)
	app.Arg("file.json", "template file to be used").Default("template.json").StringVar(&tmplFileName)
}

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if _, err := os.Stat(tmplFileName); os.IsNotExist(err) {
		log.Fatalf("File '%s' does not exist.", tmplFileName)
	}
	var err error

	if tmplFile, err = os.Open(tmplFileName); err != nil {

		switch {
		case os.IsPermission(err):
			log.Fatalf("Insufficient permissions to read file: %s", err)
		default:
			log.Fatalf("Error opening '%s': %s", tmplFileName, err)
		}
	}

	defer tmplFile.Close()

	var tmpl map[string]interface{}

	dec := json.NewDecoder(tmplFile)

	if err = dec.Decode(&tmpl); err != nil {
		panic(err)
	}

	doc := make(map[string]interface{})
	log.Printf("Before loop: %v", tmpl)

	p := time.Unix(0, 0)
	dateDec, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      &p,
		DecodeHook: func(
			f reflect.Kind,
			t reflect.Kind,
			data interface{}) (interface{}, error) {
			log.Println(data, "bar", t, f)
			s, ok := data.(map[string]interface{})
			if !ok {
				return nil, errors.New("Invalid Value")
			}
			d := ISODate{}
			if min, ok := s["min"].(string); ok && s["min"] != "" {
				if min == "$now" {
					d.Min = time.Now()
				} else if d.Min, err = dateparse.ParseAny(min); err != nil {
					log.Printf("Could not parse min date '%s' to date: %s", min, err)
					return nil, err
				}
			}

			if max, ok := s["max"].(string); ok && s["max"] != "" {
				if max == "$now" {
					d.Max = time.Now()
				} else if d.Max, err = dateparse.ParseAny(max); err != nil {
					log.Printf("Could not parse max date '%s' to date: %s", max, err)
					return nil, err
				}
			} else {
				d.Max = time.Now()
			}
			log.Println("Date range:", d.Min, d.Max)
			return d, nil
		},
	})

	for kk, v := range tmpl {
		if t, ok := v.(map[string]interface{}); ok {
			for k, vv := range t {
				switch k {
				case "$numberDecimal":
					d := Decimal{}
					mapstructure.Decode(vv, &d)
					log.Println("Generated", d.Generate())
					doc[kk] = &d
				case "$date":
					d := ISODate{}
					dateDec.Decode(vv)
					log.Println("Generated", d.Generate())
					doc[kk] = &d
				}
			}
		}
	}

	log.Printf("After loop: %#v", doc)

	for k, v := range doc {
		log.Println(v)
		if vv, ok := v.(Generator); !ok {
			panic("Should be matched!")
		} else {
			doc[k] = vv.Generate()
		}
	}

	b, _ := bson.MarshalJSON(&doc)
	log.Println("BSON", string(b))
	b2, e := json.MarshalIndent(&doc, "> ", "  ")
	if e != nil {
		panic(e)
	}
	log.Println("JSON", string(b2))

	i := ISODate{}
	i.Max = time.Now()
	i.Min = time.Now().Add(-48 * time.Hour)
	log.Println(i.Generate())
}
