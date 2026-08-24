package workers

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/softivite/puxbay/internal/config"
	"gorm.io/gorm"
)

const (
	TypeWebhookDelivery = "webhook:delivery"
	TypePriceSync       = "product:price_sync"
	TypePrintJob        = "print:job"
	TypeAccountingSync  = "accounting:sync"
)

// WorkerServer wraps the Asynq server.
type WorkerServer struct {
	server *asynq.Server
	mux    *asynq.ServeMux
	db     *gorm.DB
	client *asynq.Client
}

// NewWorkerServer initializes the Asynq server and registers task handlers.
func NewWorkerServer(cfg *config.Config, db *gorm.DB) *WorkerServer {
	redisOpt, err := asynq.ParseRedisURI(cfg.Redis.URL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URI: %v", err)
	}

	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally configure queues with different priorities
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	mux := asynq.NewServeMux()
	ws := &WorkerServer{
		server: server,
		mux:    mux,
		db:     db,
		client: asynq.NewClient(redisOpt),
	}

	// Register task handlers
	mux.HandleFunc(TypeWebhookDelivery, ws.HandleWebhookDelivery)
	mux.HandleFunc(TypePriceSync, ws.HandlePriceSync)
	mux.HandleFunc(TypePrintJob, ws.HandlePrintJob)
	mux.HandleFunc(TypeAccountingSync, ws.HandleAccountingSync)

	return ws
}

// Client returns the Asynq client for enqueuing tasks from handlers.
func (ws *WorkerServer) Client() *asynq.Client {
	return ws.client
}

// Start begins processing tasks.
func (ws *WorkerServer) Start() error {
	log.Println("👷 Starting Asynq worker server...")
	return ws.server.Start(ws.mux)
}

// Stop gracefully shuts down the server.
func (ws *WorkerServer) Stop() {
	ws.server.Stop()
	ws.client.Close()
}
