# RW Bank

## Задание
В этом задании вам предлагается написать свой многопоточный банк, используя тонкие блокировки и RWMutex

```go
// MaxAmount is the maximum allowed amount per account.
const MaxAmount int64 = 1_000_000_000_000_000_000

var (
	// ErrInvalidAmount is returned when amount <= 0.
	ErrInvalidAmount = errors.New("invalid amount")

	// ErrOverflow is returned when an operation would exceed MaxAmount.
	ErrOverflow = errors.New("overflow")

	// ErrUnderflow is returned when an account lacks sufficient funds.
	ErrUnderflow = errors.New("underflow")
)

// Bank defines operations for a thread-safe multi-account bank.
type Bank interface {
	// NumberOfAccounts returns the number of accounts in the bank.
	NumberOfAccounts() int

	// GetAmount returns the current balance of the account at the given index.
	//
	// Panics if index is out of bounds.
	GetAmount(index int) int64

	// TotalAmount returns the total sum of money across all accounts.
	TotalAmount() int64

	// Deposit adds a positive amount to the specified account and returns the new balance.
	//
	// Returns:
	//   - ErrInvalidAmount if amount <= 0
	//   - ErrOverflow if the resulting amount would exceed MaxAmount
	//   - panics if index is out of bounds
	Deposit(index int, amount int64) (int64, error)

	// Withdraw subtracts a positive amount from the specified account and returns the new balance.
	//
	// Returns:
	//   - ErrInvalidAmount if amount <= 0
	//   - ErrUnderflow if the account does not have enough funds
	//   - panics if index is out of bounds
	Withdraw(index int, amount int64) (int64, error)

	// Transfer moves a positive amount from one account to another.
	//
	// Returns:
	//   - ErrInvalidAmount if amount <= 0
	//   - error if fromIndex == toIndex
	//   - ErrUnderflow if source account lacks funds
	//   - ErrOverflow if the destination account would overflow
	//   - panics if either index is out of bounds
	Transfer(fromIndex, toIndex int, amount int64) error

	// Consolidate transfers all funds from a list of source accounts to the target account.
	// All fromIndices must be unique and must not include toIndex
	//
	// Returns:
	//   - error if fromIndices is empty, contains duplicates, or includes toIndex
	//   - ErrOverflow if the target account would exceed MaxAmount
	//   - panics if any index is out of bounds
	Consolidate(fromIndices []int, toIndex int) error
}
```

## Сдача
* Решение необходимо реализовать в файле [bank.go](./internal/bank/bank.go)
* Открыть pull request из ветки `hw` в ветку `main` **вашего репозитория**
* В описании PR заполнить количество часов, которые вы потратили на это задание

## Особенности реализации
* Используйте тесты и линтер, чтобы заполнить недосказанности, они проверяют все требования условия
* Рекомендуется сначала написать однопоточную реализацию


## Скрипты
Для запуска скриптов на курсе необходимо установить [go-task](https://taskfile.dev/docs/installation)

`go install github.com/go-task/task/v3/cmd/task@latest`

Перед выполнением задания не забудьте выполнить:

```bash 
task update
```

Запустить линтер:
```bash 
task lint
```

Запустить тесты:
```bash
task test
``` 

Обновить файлы задания
```bash
task update
```

Принудительно обновить файлы задания
```bash
task force-update
```

Скрипты работают на Windows, однако при разработке на этой операционной системе
рекомендуется использовать [WSL](https://learn.microsoft.com/en-us/windows/wsl/install)
