package main

import (
	"fmt"
	"net/http"
	"os"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"
	// "github.com/pivotal-cf/brokerapi/auth"

	"github.com/starkandwayne/kafka-service-broker/broker"
	"github.com/starkandwayne/kafka-service-broker/brokerconfig"
	"github.com/starkandwayne/kafka-service-broker/kafka"
)

// Version set by ci/scripts/shipit; if "" it means "development"
var Version = ""

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-v" || os.Args[1] == "--version" {
			if Version == "" {
				fmt.Printf("kafka-service-broker (development)\n")
			} else {
				fmt.Printf("kafka-service-broker v%s\n", Version)
			}
			os.Exit(0)
		}
	}

	brokerLogger := lager.NewLogger("kafka-service-broker")
	brokerLogger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	brokerLogger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	brokerLogger.Info("Starting Kafka service broker")

	config, err := brokerconfig.LoadConfig()
	if err != nil {
		panic(err)
	}

	topicRepo := kafka.NewTopicRepository(config.KafkaConfiguration, brokerLogger)
	sharedPlanRepo := kafka.NewSharedPlanRepository(config.KafkaConfiguration, brokerLogger)

	serviceBroker := &broker.KafkaServiceBroker{
		InstanceCreators: map[string]broker.InstanceCreator{
			"topic":  topicRepo,
			"shared": sharedPlanRepo,
		},
		InstanceBinders: map[string]broker.InstanceBinder{
			"topic":  topicRepo,
			"shared": sharedPlanRepo,
		},
		Config: config,
	}

	brokerCredentials := brokerapi.BrokerCredentials{
		Username: config.Broker.Username,
		Password: config.Broker.Password,
	}

	brokerAPI := brokerapi.New(serviceBroker, brokerLogger, brokerCredentials)

	brokerLogger.Info("listening :" + config.Broker.ListenPort)
	http.Handle("/", brokerAPI)
	brokerLogger.Fatal("http-listen", http.ListenAndServe("0.0.0.0:"+config.Broker.ListenPort, nil))
}
