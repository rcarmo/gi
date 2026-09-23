package web

import (
	"context"
	"fmt"

	"github.com/rcarmo/gi/internal/search/indexer"
	searchstore "github.com/rcarmo/gi/internal/search/store"
)

// StartWorkspaceIndex binds explicit refresh work to application lifetime.
// New deliberately does not start goroutines; embeddings must start and defer
// CloseWorkspaceIndex before closing the database. Construction never scans.
// Invalid settings disable indexing without preventing chat/server startup.
func (s *Server) StartWorkspaceIndex(ctx context.Context) error {
	s.indexMu.Lock()
	defer s.indexMu.Unlock()
	if s.indexClosed {
		return indexer.ErrSchedulerClosed
	}
	if s.indexScheduler != nil {
		return nil
	}
	if ctx == nil {
		return fmt.Errorf("workspace index requires application context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	configs := make([]searchstore.ScopeConfig, 0, 3)
	for _, scope := range []string{"all", "notes", "skills"} {
		config, err := s.configuredWorkspaceScope(scope)
		if err != nil {
			return err
		}
		configs = append(configs, config)
	}
	scheduler, err := indexer.NewScheduler(ctx, searchstore.NewRefreshStore(s.store.DB()), configs)
	if err != nil {
		return err
	}
	s.indexConfigs = make(map[string]searchstore.ScopeConfig, len(configs))
	for _, config := range configs {
		s.indexConfigs[config.Scope()] = config
	}
	s.indexScheduler = scheduler
	return nil
}

// CloseWorkspaceIndex rejects new POST work and joins scans, renewals and
// failure cleanup. Concurrent/repeated close calls all wait for the same drain.
func (s *Server) CloseWorkspaceIndex() {
	s.indexMu.Lock()
	s.indexClosed = true
	scheduler := s.indexScheduler
	s.indexMu.Unlock()
	if scheduler != nil {
		scheduler.Close()
	}
}

func (s *Server) requestWorkspaceRefresh(scope string) (*indexer.RefreshTicket, error) {
	s.indexMu.Lock()
	defer s.indexMu.Unlock()
	if s.indexClosed || s.indexScheduler == nil {
		return nil, indexer.ErrSchedulerClosed
	}
	return s.indexScheduler.Request(scope)
}
