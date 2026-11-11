# Banking System (SDP Final Project)

Web banking app built with Go (Echo) and a single-page frontend showing six GoF design patterns working together.

## Quickstart

Requirements: Go 1.20+ (tested on Go 1.27).

Run the server:
```bash
go run ./cmd/api/main.go
```
Open `http://localhost:8080` in your browser.

Run all tests:
```bash
go test -v ./...
```

---

## Design Patterns Overview

### Core Pattern Code

**Singleton (`internal/db/singleton.go`)** - shared JSON ledger store (`LedgerStore`) guarded by an `RWMutex`:
```go
func GetDBInstance() *LedgerStore {
	initOnce.Do(func() {
		path := os.Getenv("BANK_DB_PATH")
		if path == "" {
			path = defaultLedgerPath
		}
		sharedLedger = openLedger(path)
	})
	return sharedLedger
}
```

**Factory Method (`internal/patterns/factory.go`)** - account types carry per-product credit limits and single-transfer caps:
```go
var accountRegistry = map[string]AccountCreator{
	"checking": CheckingCreator{},
	"savings":  SavingsCreator{},
	"credit":   CreditCreator{},
}
```

**Observer (`internal/patterns/observer.go`)** - `FraudDetector` checks both the `$10,000` AML threshold and transfer velocity (`>3 tx/min`):
```go
if e.Amount > f.Threshold {
	e.Suspicious = true
	e.Reason = fmt.Sprintf("exceeds $%.0f AML limit", f.Threshold)
}
```

**Facade (`internal/service/facade.go`)** - `Transfer` coordinates validation, `TransactionMonitor`, and `LedgerStore.ApplyDebit`:
```go
ev := f.txMonitor.ProcessTransaction(amount, accID)
entry := models.TransactionRecord{
	AccountID:  accID,
	Name:       "Transfer",
	Amount:     amount,
	Suspicious: ev.Suspicious,
	Reason:     ev.Reason,
	Date:       ev.Time,
}
balance, err := ledger.ApplyDebit(entry)
```

### How the Patterns Interact

| Scenario | Flow |
|---|---|
| Open account | `Facade.OpenAccount` -> Factory Method validates ID & builds account with `CreditLimit` -> Singleton saves it |
| Transfer | `Facade.Transfer` -> Singleton checks balance/credit limit & transfer cap -> Observer (`FraudDetector` checks AML + velocity, `AuditLogger` logs, `SecurityTeamAlert` alerts) -> Singleton applies debit under one lock |
| Loan | `Facade.ApplyForLoan` -> `LoanDirector` runs `StandardLoanBuilder` (enforces `$100`–`$500,000` bounds) -> Singleton saves contract |
| Interest | `Facade.CalculateYield` -> Strategy lookup (`checking` 0%, `savings` 5%, `credit`/`loan` 15%) -> `InterestContext` calculates yield |

- `Builder` and `Strategy` share `StandardLoanRatePercent` (`15.0%`) so loan rates stay consistent.
- `BankFacade` registers `FraudDetector` before `AuditLogger` and `SecurityTeamAlert`.

---

## Division of Work

| Student | Patterns | Core Files | Other Work |
|---|---|---|---|
| Shynarbek Nuraziz (Student A) | Singleton, Factory Method, Facade | `internal/db/singleton.go`, `internal/patterns/factory.go`, `internal/service/facade.go` | Models (`internal/models/data.go`), HTTP API (`cmd/api/main.go`), DB & service tests |
| Asset Kamza (Student B) | Builder, Strategy, Observer | `internal/patterns/builder.go`, `internal/patterns/strategy.go`, `internal/patterns/observer.go` | Web UI (`web/index.html`), pattern unit tests, project report |

---

## Project Structure

```text
cmd/api/main.go                  HTTP server and Echo routes
data/database.json               JSON storage file
docs/                            UML diagrams and JSON database schema
internal/db/singleton.go         Thread-safe JSON ledger store (Singleton)
internal/db/db_test.go           Storage, credit limit, and concurrency tests
internal/models/data.go          Account, Loan, and Transaction structs
internal/patterns/builder.go     Loan agreement builder and director
internal/patterns/errors.go      Validation sentinel error
internal/patterns/factory.go     Account factory and account type policies
internal/patterns/observer.go    Transaction monitor, AML + velocity fraud detector
internal/patterns/strategy.go    Interest calculation strategies
internal/patterns/patterns_test.go Unit tests for patterns
internal/service/facade.go       BankFacade service layer
internal/service/facade_test.go  Integration tests through facade
web/index.html                   Single-page frontend
go.mod, go.sum                   Go module files
```

---

## API Endpoints

| Method | Endpoint | Purpose | Pattern |
|---|---|---|---|
| `GET` | `/api/accounts` | List all accounts | Singleton, Facade |
| `POST` | `/api/account` | Open an account (`checking`, `savings`, `credit`) | Factory Method, Singleton |
| `GET` | `/api/transactions` | List recent transactions | Singleton, Facade |
| `POST` | `/api/transaction` | Send a transfer and trigger AML/velocity alerts | Observer, Singleton |
| `POST` | `/api/loan` | Create a loan agreement | Builder, Singleton |
| `GET` | `/api/interest` | Calculate annual interest | Strategy |

---

## Testing

All 10 test functions pass across `internal/db`, `internal/patterns`, and `internal/service`:

```text
?   	ESHKEREE/cmd/api	[no test files]
=== RUN   TestSeedAndPersistence
--- PASS: TestSeedAndPersistence (0.16s)
=== RUN   TestDuplicateAccount
--- PASS: TestDuplicateAccount (0.00s)
=== RUN   TestDebitRulesAndCreditLimit
--- PASS: TestDebitRulesAndCreditLimit (0.01s)
=== RUN   TestConcurrentDebits
--- PASS: TestConcurrentDebits (0.09s)
PASS
ok  	ESHKEREE/internal/db	1.166s
?   	ESHKEREE/internal/models	[no test files]
=== RUN   TestStrategySelection
--- PASS: TestStrategySelection (0.00s)
=== RUN   TestBuilderValidationAndPayments
--- PASS: TestBuilderValidationAndPayments (0.00s)
=== RUN   TestFactoryMethod
--- PASS: TestFactoryMethod (0.00s)
=== RUN   TestObserverNotifiesAndDetectsVelocity
--- PASS: TestObserverNotifiesAndDetectsVelocity (0.00s)
PASS
ok  	ESHKEREE/internal/patterns	0.790s
=== RUN   TestTransferFlow
[AUDIT] 23:17:04 | account ACC-1 | $500.00 | OK
[AUDIT] 23:17:04 | account ACC-1 | $15000.00 | FLAGGED
[ALERT] Suspicious transaction $15000.00 from account ACC-1 (exceeds $10000 AML limit)! Security team notified.
--- PASS: TestTransferFlow (0.15s)
=== RUN   TestAccountsAndLoansAndInterest
[AUDIT] 23:17:04 | account CR-9 | $300.00 | OK
--- PASS: TestAccountsAndLoansAndInterest (0.00s)
PASS
ok  	ESHKEREE/internal/service	1.016s
```
