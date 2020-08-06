package main

import (
	"sync"

	"github.com/hashicorp/go-hclog"
	"go.mongodb.org/mongo-driver/mongo"
)

func aggregate(log hclog.Logger, in ...chan *mongo.InsertOneModel) <-chan *mongo.InsertOneModel {
	aggregated := make(chan *mongo.InsertOneModel)

	go func() {
		log.Debug("Starting aggregator", "pumps", len(in))
		var wg sync.WaitGroup
		for n, c := range in {
			log.Debug("Starting aggregator", "pump", n+1)
			wg.Add(1)
			go func(ch chan *mongo.InsertOneModel, number int) {
				log.Debug("Started aggregator", "pump", number+1)
				for op := range ch {
					log.Debug("Sending downstream", "pump", number+1)
					aggregated <- op
					log.Debug("Sent downstream", "pump", number+1)
				}
				log.Debug("Done aggregating", "pump", number+1)
				wg.Done()
			}(c, n)
		}
		wg.Wait()
		close(aggregated)
	}()
	return aggregated
}
