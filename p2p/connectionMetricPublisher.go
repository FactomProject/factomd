package p2p

import (
	"github.com/FactomProject/factomd/modules/events"
	"github.com/FactomProject/factomd/modules/pubsub"
	"github.com/FactomProject/factomd/modules/worker"
)

type MetricPublisher interface {
	Start(w *worker.Thread)
}
type metricPublisher struct {
	publisher                pubsub.IPublisher
	connectionMetricsChannel chan interface{}
}

func NewMetricPublisher(factomNodeName string, connectionMetricChannel chan interface{}) MetricPublisher {
	publisher := pubsub.PubFactory.Threaded(5).Publish(pubsub.GetPath(factomNodeName, events.Path.ConnectionMetrics))
	go publisher.Start()
	return &metricPublisher{
		publisher:                publisher,
		connectionMetricsChannel: connectionMetricChannel,
	}
}

func (metricPublisher *metricPublisher) Start(w *worker.Thread) {
	w.Spawn("ConnectionMetricsPublisher", metricPublisher.publishMetrics)
}

func (metricPublisher *metricPublisher) publishMetrics(w *worker.Thread) {
	w.OnRun(func() {
		for {
			select {
			case metric := <-metricPublisher.connectionMetricsChannel:
				metricPublisher.publisher.Write(metric)
			}
		}
	})
}
