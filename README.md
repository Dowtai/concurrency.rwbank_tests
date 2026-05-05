# Read-write Bank

## Задание
В этом задании вам предлагается написать свой многопоточный банк


---

```go
func New(accountsNum int, maxAmount ...int64) *bankImpl {
...
}
```

В конструкторе банка передаётся количество аккаунтов `accountsNum`,
а также может передаваться ограничения на количество денег у пользователя.
Если ограничение не передаётся, необходимо использовать `DefaultMaxAmount`.


---

```go
func (b *bankImpl) NumberOfAccounts() int {
	...
}
```

Метод ```NumberOfAccounts``` должен возвращать количество аккаунтов в банке

---

```go
func (b *bankImpl) GetAmount(index int) (int64, error) {
    ...
}
```

Метод `GetAmount` должен возвращать количество денег на аккаунте под номером `index`. При этом индексация аккаунтов
начинается с 0.

---

```go
func (b *bankImpl) TotalAmount() int64 {
	...
}
```

Метод `TotalAmount` должен возвращать суммарное количество денег на всех аккаунтах банка.

---

```go
func (b *bankImpl) Deposit(index int, amount int64) (int64, error) {
	...
}
```

Метод `Deposit` пополняет счёт аккаунта под номером `index`, возвращая новую сумму на счёте после пополнения.


```go
const MaxAmount int64 = 1_000_000_000_000_000_000
var ErrOverflow = errors.New("overflow")
```

Если сумма на счёте превысила `MaxAmount`, метод должен вернуть ошибку `ErrOverflow`.

```go
var ErrInvalidAmount = errors.New("invalid amount")
```

Если сумма `amount` меньше или равна нулю, метод должен вернуть ошибку `ErrInvalidAmount`.

---

```go
func (b *bankImpl) Withdraw(index int, amount int64) (int64, error) {
	...
}
```

Метод `Withdraw` выводит деньги с аккаунта под номером `index`, возвращая новую сумму на счёте после пополнения.


```go
var ErrUnderflow = errors.New("underflow")
```

Метод должен вернуть ошибку `ErrUnderflow`, если сумма вывода превышает остаток на счёте

---

```go
func (b *bankImpl) Transfer(fromIndex, toIndex int, amount int64) error {
    ...
}
```

Метод `Transfer` переводит `amount` единиц валюты с аккаунта под номером `fromIndex` на аккаунт под номером `toIndex`.

---


```go
func (b *bankImpl) Consolidate(fromIndices []int, toIndex int) error {
	...
}
```

Метод `Consolidate` переводит все деньги с аккаунтов `fromIndices` на аккаунт `toIndex`.

---
Общие советы по выполнению:

Стоит обратить внимание на создание ошибок при возврате, и подумать как лучше это сделать.

## Сдача
* Решение необходимо реализовать в файле [bank.go](./internal/bank/bank.go)
* Открыть pull request из ветки `hw` в ветку `main` **вашего репозитория**
* В описании PR заполнить количество часов, которые вы потратили на это задание

## Особенности реализации
* Используйте тесты и линтер, чтобы заполнить недосказанности, они проверяют все требования условия
* Рекомендуется сначала написать однопоточную реализацию
* В этом задании необходимо использовать RWMutex и тонкие блокировки


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
