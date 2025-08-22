package sqlite

import (
	"context"
	"fmt"
	"time"
)

type SessionCleanupService struct {
	db       *DB
	stopChan chan struct{}
	doneChan chan struct{}
}

func NewSessionCleanupService(db *DB) *SessionCleanupService {
	s := &SessionCleanupService{
		db:       db,
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}
	go s.cleanupExpiredSessions()
	return s
}

// Close gracefully shuts down the background cleanup process
// This method blocks until the cleanup process has completely stopped
func (svc *SessionCleanupService) Close() {
	close(svc.stopChan)
	<-svc.doneChan
}

// cleanupExpiredSessions runs every 6 hours and deletes expired sessions
// until there are none remaining
func (svc *SessionCleanupService) cleanupExpiredSessions() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()
	defer close(svc.doneChan)

	for {
		select {
		case <-ticker.C:
			svc.runCleanup()
		case <-svc.stopChan:
			return
		}
	}
}

// runCleanup deletes expired sessions in batches until none remain
func (svc *SessionCleanupService) runCleanup() error {
	ctx := context.Background()
	for {
		err := svc.db.Queries.DeleteExpiredSessions(ctx)
		if err != nil {
			return fmt.Errorf("error deleting expired sessions: %w", err)
		}

		count, err := svc.db.Queries.DeletedExpiredSessionsCount(ctx)
		if err != nil {
			return fmt.Errorf("error counting deleted expired sessions: %w", err)
		}

		// If no sessions were deleted, we're done
		if count == 0 {
			return nil
		}
	}
}
