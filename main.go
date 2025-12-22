package main

import (
	"context"
	"flag"
	"log"
	"strings"

	"github.com/etcdfinder/etcdfinder/internal/api"
	v1 "github.com/etcdfinder/etcdfinder/internal/api/v1"
	"github.com/etcdfinder/etcdfinder/internal/config"
	"github.com/etcdfinder/etcdfinder/internal/ingestor"
	"github.com/etcdfinder/etcdfinder/internal/lib"
	"github.com/etcdfinder/etcdfinder/internal/service"
	"github.com/etcdfinder/etcdfinder/pkg/etcd"
	"github.com/etcdfinder/etcdfinder/pkg/kvstore"
	"github.com/etcdfinder/etcdfinder/pkg/logger"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.Parse()

	ctx := context.Background()

	conf, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err = logger.NewLogger(conf); err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	logger.Infof("Initializing application...")

	// Initialize etcd client
	logger.Infof("Connecting to etcd at %s", conf.Etcd.Endpoints)
	var etcdClient etcd.BaseClient
	if conf.Etcd.Version == lib.ETCD_V3 {
		etcdClient, err = etcd.NewClientV3(
			strings.Split(conf.Etcd.Endpoints, lib.ETCD_ENDPOINTS_SEPERATOR),
			conf.Etcd.WatchEventChannelSize,
			conf.Etcd.RootPrefixEtcd,
			conf.Etcd.PaginationLimit,
			conf.Etcd.EtcdAuditPeriod,
			conf.Etcd.MaxWatchRetries,
		)
	} else {
		etcdClient, err = etcd.NewClientV2(
			strings.Split(conf.Etcd.Endpoints, lib.ETCD_ENDPOINTS_SEPERATOR),
			conf.Etcd.WatchEventChannelSize,
			conf.Etcd.RootPrefixEtcd,
			conf.Etcd.PaginationLimit,
			conf.Etcd.EtcdAuditPeriod,
			conf.Etcd.MaxWatchRetries,
		)
	}
	if err != nil {
		logger.Fatalf("Failed to create etcd client: %v", err)
	}
	defer etcdClient.Close() //nolint

	// Initialize Elasticsearch KV store
	var kvStore kvstore.KVStore
	if conf.Datastore.Type == "meilisearch" {
		kvStore, err = kvstore.NewMeilisearchStore(
			conf.Datastore.Meilisearch.Host,
			conf.Datastore.Meilisearch.IndexName,
			conf.Datastore.Meilisearch.MatchingStrategy)
		if err != nil {
			logger.Fatalf("Failed to create Meilisearch store: %v", err)
		}
	} else {
		logger.Fatalf("Unsupported datastore type: %s", conf.Datastore.Type)
	}
	defer kvStore.Close(ctx) //nolint

	// Initialize ingestor
	ing := ingestor.NewIngestor(kvStore, etcdClient)

	// Start watching for etcd changes in background
	go func() {
		if err := ing.ChangeUpdater(ctx); err != nil {
			logger.Fatalf("ChangeUpdater failed: %v", err)
		}
		logger.Fatalf("ChangeUpdater stopped unexpectedly")
	}()

	// Start etcd connection auditor in background
	go func() {
		auditorErrCh := etcdClient.StartAuditor(ctx)
		if err := <-auditorErrCh; err != nil {
			logger.Fatalf("Etcd connection auditor failed: %v", err)
		}
		logger.Fatalf("Etcd connection auditor stopped unexpectedly")
	}()

	logger.Debugf("Initializing KV store with existing etcd data...")

	// Initialize KV store with existing etcd data
	if err := ing.InitKVStore(ctx); err != nil {
		logger.Fatalf("Failed to initialize KV store: %v", err)
	}

	// Initialize service layer
	etcdFinderService := service.NewDefaultEtcdfinder(etcdClient, kvStore, ing)

	// Initialize router with handlers
	router, err := api.NewRouter(api.Handlers{
		EtcdFinderHandler: v1.NewEtcdfinderHandler(etcdFinderService),
	})
	if err != nil {
		logger.Fatalf("Failed to create router: %v", err)
	}

	// Start the server
	logger.Infof("Starting server on :%s", conf.Server.Port)
	if err := router.Run(":" + conf.Server.Port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
