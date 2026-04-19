package chat

import (
	"sync"
	"time"

	"mrchat/internal/modules/catalog"
)

type upstreamCooldownTracker struct {
	mu      sync.Mutex
	now     func() time.Time
	entries map[string]upstreamCooldownState
}

type upstreamCooldownState struct {
	consecutiveFailures int
	blacklistUntil      time.Time
}

func newUpstreamCooldownTracker() *upstreamCooldownTracker {
	return &upstreamCooldownTracker{
		now:     time.Now,
		entries: make(map[string]upstreamCooldownState),
	}
}

func (t *upstreamCooldownTracker) isCoolingDown(upstreamID string) (bool, time.Time, int) {
	if t == nil {
		return false, time.Time{}, 0
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	state, ok := t.entries[upstreamID]
	if !ok {
		return false, time.Time{}, 0
	}

	now := t.now().UTC()
	if state.blacklistUntil.IsZero() || !now.Before(state.blacklistUntil) {
		state.blacklistUntil = time.Time{}
		if state.consecutiveFailures <= 0 {
			delete(t.entries, upstreamID)
			return false, time.Time{}, 0
		}

		t.entries[upstreamID] = state
		return false, time.Time{}, state.consecutiveFailures
	}

	return true, state.blacklistUntil, state.consecutiveFailures
}

func (t *upstreamCooldownTracker) recordFailure(upstream *catalog.Upstream) (int, time.Time) {
	if t == nil || upstream == nil {
		return 0, time.Time{}
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.entries[upstream.ID]
	now := t.now().UTC()
	if !state.blacklistUntil.IsZero() && !now.Before(state.blacklistUntil) {
		state.blacklistUntil = time.Time{}
	}

	state.consecutiveFailures++

	threshold := upstream.FailureThreshold
	if threshold <= 0 {
		threshold = 1
	}

	if state.consecutiveFailures >= threshold && upstream.CooldownSeconds > 0 {
		state.blacklistUntil = now.Add(time.Duration(upstream.CooldownSeconds) * time.Second)
	}

	t.entries[upstream.ID] = state
	return state.consecutiveFailures, state.blacklistUntil
}

func (t *upstreamCooldownTracker) recordSuccess(upstreamID string) {
	if t == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.entries, upstreamID)
}
