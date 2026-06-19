package audit

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Entry represents a single audit log record.
type Entry struct {
	ActorID      int64       `json:"actor_id"`
	ActorRole    string      `json:"actor_role"` // user, system, ai
	Action       string      `json:"action"`     // create, update, delete, approve, rollback
	ResourceType string      `json:"resource_type"`
	ResourceID   int64       `json:"resource_id"`
	RequestID    string      `json:"request_id"`
	Detail       interface{} `json:"detail,omitempty"`
	Result       string      `json:"result"` // success, failure (default: success)
}

// Writer is a non-blocking audit log writer.
type Writer struct {
	db     *pgxpool.Pool
	log    *zap.Logger
	ch     chan Entry
	done   chan struct{}
}

// NewWriter creates a buffered audit writer with a background goroutine.
func NewWriter(db *pgxpool.Pool, log *zap.Logger, bufferSize int) *Writer {
	if bufferSize <= 0 {
		bufferSize = 4096
	}
	w := &Writer{
		db:   db,
		log:  log.Named("audit"),
		ch:   make(chan Entry, bufferSize),
		done: make(chan struct{}),
	}
	go w.run()
	return w
}

func (w *Writer) run() {
	defer close(w.done)
	batch := make([]Entry, 0, 64)
	timer := time.NewTicker(1 * time.Second)
	defer timer.Stop()

	for {
		select {
		case e, ok := <-w.ch:
			if !ok {
				if len(batch) > 0 {
					w.flush(batch)
				}
				return
			}
			batch = append(batch, e)
			if len(batch) >= 64 {
				w.flush(batch)
				batch = batch[:0]
			}
		case <-timer.C:
			if len(batch) > 0 {
				w.flush(batch)
				batch = batch[:0]
			}
		}
	}
}

func (w *Writer) flush(batch []Entry) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, e := range batch {
		result := e.Result
		if result == "" {
			result = "success"
		}
		_, err := w.db.Exec(ctx,
			`INSERT INTO audit_log (actor_id, actor_role, action, resource_type, resource_id, request_id, detail, result, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			e.ActorID, e.ActorRole, e.Action, e.ResourceType, e.ResourceID,
			e.RequestID, e.Detail, result, time.Now().UTC(),
		)
		if err != nil {
			w.log.Error("failed to write audit log", zap.Error(err), zap.String("action", e.Action))
		}
	}
}

// Log enqueues a single audit entry (non-blocking).
func (w *Writer) Log(e Entry) {
	select {
	case w.ch <- e:
	default:
		w.log.Warn("audit buffer full, dropping entry", zap.String("action", e.Action))
	}
}

// Close flushes remaining entries and stops the background goroutine.
func (w *Writer) Close() {
	close(w.ch)
	<-w.done
}
