package domain

import (
	"testing"
)

func TestServiceStatusTransitions(t *testing.T) {
	tests := []struct {
		name      string
		from      ServiceStatus
		to        ServiceStatus
		wantValid bool
	}{
		// Valid transitions
		{"draft -> creating", ServiceDraft, ServiceCreating, true},
		{"draft -> archived", ServiceDraft, ServiceArchived, true},
		{"creating -> ready", ServiceCreating, ServiceReady, true},
		{"creating -> degraded", ServiceCreating, ServiceDegraded, true},
		{"creating -> archived", ServiceCreating, ServiceArchived, true},
		{"ready -> degraded", ServiceReady, ServiceDegraded, true},
		{"ready -> archived", ServiceReady, ServiceArchived, true},
		{"degraded -> ready", ServiceDegraded, ServiceReady, true},
		{"degraded -> archived", ServiceDegraded, ServiceArchived, true},

		// Invalid transitions
		{"draft -> ready (skip creating)", ServiceDraft, ServiceReady, false},
		{"draft -> degraded", ServiceDraft, ServiceDegraded, false},
		{"ready -> draft", ServiceReady, ServiceDraft, false},
		{"ready -> creating", ServiceReady, ServiceCreating, false},
		{"archived -> draft", ServiceArchived, ServiceDraft, false},
		{"archived -> ready", ServiceArchived, ServiceReady, false},
		{"degraded -> creating", ServiceDegraded, ServiceCreating, false},
		{"degraded -> draft", ServiceDegraded, ServiceDraft, false},
		{"creating -> draft", ServiceCreating, ServiceDraft, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			if got != tt.wantValid {
				t.Errorf("%s -> %s: got %v, want %v", tt.from, tt.to, got, tt.wantValid)
			}
		})
	}
}

func TestReleaseStatusTransitions(t *testing.T) {
	tests := []struct {
		name      string
		from      ReleaseStatus
		to        ReleaseStatus
		wantValid bool
	}{
		// Valid transitions — full happy path
		{"created -> reviewing", ReleaseCreated, ReleaseReviewing, true},
		{"created -> deploying", ReleaseCreated, ReleaseDeploying, true},
		{"reviewing -> approved", ReleaseReviewing, ReleaseApproved, true},
		{"reviewing -> failed", ReleaseReviewing, ReleaseFailed, true},
		{"approved -> deploying", ReleaseApproved, ReleaseDeploying, true},
		{"approved -> failed", ReleaseApproved, ReleaseFailed, true},
		{"deploying -> verifying", ReleaseDeploying, ReleaseVerifying, true},
		{"deploying -> failed", ReleaseDeploying, ReleaseFailed, true},
		{"deploying -> rollback_pending", ReleaseDeploying, ReleaseRollbackPending, true},
		{"verifying -> observing", ReleaseVerifying, ReleaseObserving, true},
		{"verifying -> failed", ReleaseVerifying, ReleaseFailed, true},
		{"verifying -> rollback_pending", ReleaseVerifying, ReleaseRollbackPending, true},
		{"observing -> succeeded", ReleaseObserving, ReleaseSucceeded, true},
		{"observing -> rollback_pending", ReleaseObserving, ReleaseRollbackPending, true},
		{"rollback_pending -> rolling_back", ReleaseRollbackPending, ReleaseRollingBack, true},
		{"rolling_back -> rolled_back", ReleaseRollingBack, ReleaseRolledBack, true},
		{"rolling_back -> failed", ReleaseRollingBack, ReleaseFailed, true},

		// Invalid transitions
		{"created -> succeeded (skip steps)", ReleaseCreated, ReleaseSucceeded, false},
		{"created -> approved", ReleaseCreated, ReleaseApproved, false},
		{"created -> failed", ReleaseCreated, ReleaseFailed, false},
		{"succeeded -> anything", ReleaseSucceeded, ReleaseCreated, false},
		{"rolled_back -> anything", ReleaseRolledBack, ReleaseCreated, false},
		{"failed -> anything", ReleaseFailed, ReleaseCreated, false},
		{"reviewing -> deploying (must go through approved)", ReleaseReviewing, ReleaseDeploying, false},
		{"deploying -> succeeded (skip verifying/observing)", ReleaseDeploying, ReleaseSucceeded, false},
		{"observing -> rolled_back", ReleaseObserving, ReleaseRolledBack, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.CanTransitionTo(tt.to)
			if got != tt.wantValid {
				t.Errorf("%s -> %s: got %v, want %v", tt.from, tt.to, got, tt.wantValid)
			}
		})
	}
}

func TestPaginationNormalize(t *testing.T) {
	tests := []struct {
		name       string
		input      Pagination
		wantLimit  int
		wantOffset int
	}{
		{"zero limit", Pagination{Offset: 0, Limit: 0}, 20, 0},
		{"negative limit", Pagination{Offset: 0, Limit: -5}, 20, 0},
		{"negative offset", Pagination{Offset: -10, Limit: 10}, 10, 0},
		{"over 100 limit", Pagination{Offset: 0, Limit: 200}, 20, 0},
		{"limit 101", Pagination{Offset: 0, Limit: 101}, 20, 0},
		{"exactly 100 limit (allowed)", Pagination{Offset: 0, Limit: 100}, 100, 0},
		{"valid values", Pagination{Offset: 10, Limit: 50}, 50, 10},
		{"limit 1", Pagination{Offset: 0, Limit: 1}, 1, 0},
		{"limit 99", Pagination{Offset: 5, Limit: 99}, 99, 5},
		{"all negative", Pagination{Offset: -1, Limit: -1}, 20, 0},
		{"large offset valid", Pagination{Offset: 1000, Limit: 10}, 10, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.input
			p.Normalize()
			if p.Limit != tt.wantLimit {
				t.Errorf("Limit: got %d, want %d", p.Limit, tt.wantLimit)
			}
			if p.Offset != tt.wantOffset {
				t.Errorf("Offset: got %d, want %d", p.Offset, tt.wantOffset)
			}
		})
	}
}

func TestServiceTransitionsMap(t *testing.T) {
	// Verify all statuses that have outgoing transitions are in the map
	expectedStatuses := []ServiceStatus{
		ServiceDraft, ServiceCreating, ServiceReady, ServiceDegraded,
	}
	for _, s := range expectedStatuses {
		if _, ok := ServiceTransitions[s]; !ok {
			t.Errorf("ServiceTransitions missing entry for %q", s)
		}
	}
	// Archived should NOT have outgoing transitions
	if targets, ok := ServiceTransitions[ServiceArchived]; ok && len(targets) > 0 {
		t.Errorf("ServiceArchived should not have outgoing transitions, got %v", targets)
	}
}

func TestReleaseTransitionsMap(t *testing.T) {
	// Terminal states should not have outgoing transitions
	terminalStates := []ReleaseStatus{
		ReleaseSucceeded, ReleaseFailed, ReleaseRolledBack,
	}
	for _, s := range terminalStates {
		if targets, ok := ReleaseTransitions[s]; ok && len(targets) > 0 {
			t.Errorf("ReleaseTransitions[%q] should be terminal, got %v", s, targets)
		}
	}
}
