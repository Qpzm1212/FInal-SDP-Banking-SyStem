package patterns

import (
	"fmt"
	"sync"
	"time"
)

// AML review kicks in above $10,000 or more than 3 transfers on one account inside a minute.
const (
	FraudThreshold     = 10000.0
	VelocityMaxInMin   = 3
	VelocityWindowSize = time.Minute
)

type TransactionEvent struct {
	AccountID  string
	Amount     float64
	Time       time.Time
	Suspicious bool
	Reason     string
}

type Observer interface {
	Update(e *TransactionEvent)
}

type TransactionMonitor struct {
	mu        sync.RWMutex
	observers []Observer
}

func NewTransactionMonitor() *TransactionMonitor {
	return &TransactionMonitor{}
}

func (t *TransactionMonitor) AddObserver(o Observer) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.observers = append(t.observers, o)
}

func (t *TransactionMonitor) RemoveObserver(o Observer) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, obs := range t.observers {
		if obs == o {
			t.observers = append(t.observers[:i], t.observers[i+1:]...)
			return
		}
	}
}

func (t *TransactionMonitor) Notify(e *TransactionEvent) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, obs := range t.observers {
		obs.Update(e)
	}
}

func (t *TransactionMonitor) ProcessTransaction(amount float64, accID string) *TransactionEvent {
	event := &TransactionEvent{
		AccountID: accID,
		Amount:    amount,
		Time:      time.Now(),
	}
	t.Notify(event)
	return event
}

type FraudDetector struct {
	Threshold  float64
	mu         sync.Mutex
	recentHits map[string][]time.Time
}

func (f *FraudDetector) Update(e *TransactionEvent) {
	if e.Amount > f.Threshold {
		e.Suspicious = true
		e.Reason = fmt.Sprintf("exceeds $%.0f AML limit", f.Threshold)
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.recentHits == nil {
		f.recentHits = map[string][]time.Time{}
	}
	cutoff := e.Time.Add(-VelocityWindowSize)
	kept := f.recentHits[e.AccountID][:0]
	for _, ts := range f.recentHits[e.AccountID] {
		if ts.After(cutoff) {
			kept = append(kept, ts)
		}
	}
	kept = append(kept, e.Time)
	f.recentHits[e.AccountID] = kept

	if !e.Suspicious && len(kept) > VelocityMaxInMin {
		e.Suspicious = true
		e.Reason = "rapid transfer velocity (>3 tx/min)"
	}
}

type AuditLogger struct{}

func (a *AuditLogger) Update(e *TransactionEvent) {
	tag := "OK"
	if e.Suspicious {
		tag = "FLAGGED"
	}
	fmt.Printf("[AUDIT] %s | account %s | $%.2f | %s\n", e.Time.Format("15:04:05"), e.AccountID, e.Amount, tag)
}

// Needs to be added after FraudDetector so e.Suspicious is already populated.
type SecurityTeamAlert struct{}

func (s *SecurityTeamAlert) Update(e *TransactionEvent) {
	if e.Suspicious {
		reason := e.Reason
		if reason == "" {
			reason = "rule violation"
		}
		fmt.Printf("[ALERT] Suspicious transaction $%.2f from account %s (%s)! Security team notified.\n", e.Amount, e.AccountID, reason)
	}
}
