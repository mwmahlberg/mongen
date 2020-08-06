package main

import (
	"bytes"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func pump(count int, tmpl *template.Template, log hclog.Logger) (out chan *mongo.InsertOneModel) {
	out = make(chan *mongo.InsertOneModel)
	go func() {
		log.Debug("Starting pump")
		buf := bytes.NewBuffer(nil)
		var op *mongo.InsertOneModel
		// var err error
		var v bson.D
		for i := 0; i < count; i++ {
			log.Debug("Sending downstream")
			buf.Reset()
			if err := tmpl.Execute(buf, nil); err != nil {
				panic(err)
			}

			if err := bson.UnmarshalExtJSON(buf.Bytes(), false, &v); err != nil {
				panic(err)
			}
			op = mongo.NewInsertOneModel()
			op.SetDocument(v)
			out <- op
			log.Debug("Sent downstream")
		}
		log.Debug("Finished document generation", "count", count)
		close(out)
	}()
	return out
}
