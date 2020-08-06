package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"gopkg.in/cheggaaa/pb.v2"
)

func mongoSink(in <-chan *mongo.InsertOneModel, mongoURL string, commitInterval int, bar *pb.ProgressBar, log hclog.Logger) (*time.Duration, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Debug("connecting to mongodb", "url", mongoURL)
	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(mongoURL).
		SetWriteConcern(writeconcern.New(writeconcern.J(false), writeconcern.W(0))))

	if err != nil {
		return nil, fmt.Errorf("connecting to mongodb: %s", err)
	}
	log.Debug("connected to mongodb", "url", mongoURL)
	collect := client.Database("test").Collection(collection)
	// TODO: check for nil
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			log.Error("disconnecting from MongoDB", "url", mongoURL, "error", err)
		}
	}()
	ops := make([]mongo.WriteModel, 0, numOps)
	i := 0

	var once sync.Once
	var start time.Time
	for op := range in {

		once.Do(func() {
			start = time.Now()
			logger.Debug("Started", "start", start.Format(time.Kitchen))
			if !log.IsDebug() {
				bar.Start()
			}
		})

		log.Debug("Received document from upstream")
		ops = append(ops, op)
		i++
		if i%commitInterval == 0 {
			log.Debug("Committing documents", "docs", commitInterval)
			_, err := collect.BulkWrite(context.Background(), ops, options.BulkWrite().SetOrdered(false))
			if err != nil {
				return nil, fmt.Errorf("bulk operation: %s", err)
			}
			log.Debug("Committed documents", "docs", commitInterval)
			// Empty the slice, but keep the allocated memory
			ops = make([]mongo.WriteModel, 0, numOps)
		}
		bar.Increment()
	}

	if len(ops) > 0 {
		log.Debug("Commiting remaining documents", "docs", len(ops))
		collect.BulkWrite(context.Background(), ops, options.BulkWrite().SetOrdered(false))
		log.Debug("Committed remaining documents", "docs", len(ops))
	}
	if !log.IsDebug() {
		bar.SetCurrent(int64(numDocs))
		bar.Write()
		bar.Finish()
		d := bar.StartTime().Sub(time.Now())
		return &d, nil
	}
	d := time.Now().Sub(start)
	return &d, nil
}
