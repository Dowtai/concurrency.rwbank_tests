//go:build model_test

package bank

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand"
	"path/filepath"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEmptyBank(t *testing.T) {
	t.Parallel()

	bank := New(10)
	require.Equal(t, 10, bank.NumberOfAccounts())

	for i := 0; i < 10; i++ {
		require.Zero(t, bank.GetAmount(i))
	}

	require.Zero(t, bank.TotalAmount())
}

func TestDeposit(t *testing.T) {
	t.Parallel()

	bank := New(10)
	amount := int64(1234)

	result, err := bank.Deposit(1, amount)
	require.NoError(t, err)

	require.Equal(t, amount, result)
	require.Equal(t, amount, bank.GetAmount(1))
	require.Equal(t, amount, bank.TotalAmount())
}

func TestWithdraw(t *testing.T) {
	t.Parallel()

	bank := New(10)
	amount := int64(2345)

	_, err := bank.Deposit(1, amount)
	require.NoError(t, err)

	result, err := bank.Withdraw(1, 1234)
	require.NoError(t, err)

	require.EqualValues(t, 1111, result)
	require.EqualValues(t, 1111, bank.GetAmount(1))
	require.EqualValues(t, 1111, bank.TotalAmount())
}

func TestTransfer(t *testing.T) {
	t.Parallel()

	bank := New(10)
	_, err := bank.Deposit(1, 9876)
	require.NoError(t, err)

	err = bank.Transfer(1, 2, 5432)
	require.NoError(t, err)

	require.EqualValues(t, 9876-5432, bank.GetAmount(1))
	require.EqualValues(t, 5432, bank.GetAmount(2))
	require.EqualValues(t, 9876, bank.TotalAmount())
}

func TestConsolidate(t *testing.T) {
	t.Parallel()

	bank := New(10)
	_, _ = bank.Deposit(1, 4823)
	_, _ = bank.Deposit(2, 4234)
	_, _ = bank.Deposit(3, 3131)
	totalBefore := bank.TotalAmount()

	src := []int{1, 2}
	tmp := slices.Clone(src)

	err := bank.Consolidate(src, 4)
	require.NoError(t, err)
	require.Equal(t, tmp, src)

	require.Zero(t, bank.GetAmount(1))
	require.Zero(t, bank.GetAmount(2))
	require.EqualValues(t, 4823+4234, bank.GetAmount(4))

	err = bank.Consolidate([]int{3}, 4)
	require.NoError(t, err)
	require.Zero(t, bank.GetAmount(3))

	require.Equal(t, totalBefore, bank.GetAmount(4))
	require.Equal(t, totalBefore, bank.TotalAmount())
}

func TestConsolidateRandomAccounts(t *testing.T) {
	t.Parallel()

	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	numAccounts := rnd.Intn(10) + 2
	targetID := numAccounts + 1

	bank := New(100)

	totalExpected := int64(0)
	sourceIDs := make([]int, 0, numAccounts)
	for i := 1; i <= numAccounts; i++ {
		amount := int64(rnd.Intn(10_000) + 1)
		sourceIDs = append(sourceIDs, i)
		_, err := bank.Deposit(i, amount)
		require.NoError(t, err)

		totalExpected += amount
	}

	totalBefore := bank.TotalAmount()

	err := bank.Consolidate(sourceIDs, targetID)
	require.NoError(t, err)

	for _, id := range sourceIDs {
		require.Zero(t, bank.GetAmount(id), "source account %d should be empty", id)
	}

	require.EqualValues(t, totalExpected, bank.GetAmount(targetID))
	require.Equal(t, totalBefore, bank.TotalAmount())
}

func TestOverflowDeposit(t *testing.T) {
	t.Parallel()

	bank := New(1)

	_, err := bank.Deposit(0, MaxAmount)
	require.NoError(t, err)

	_, err = bank.Deposit(0, 1)
	require.ErrorIs(t, err, ErrOverflow)

	require.Equal(t, MaxAmount, bank.GetAmount(0))
}

func TestOverflowTransfer(t *testing.T) {
	t.Parallel()

	bank := New(2)

	_, err := bank.Deposit(0, MaxAmount)
	require.NoError(t, err)

	_, err = bank.Deposit(1, 1)
	require.NoError(t, err)

	err = bank.Transfer(1, 0, 1)
	require.ErrorIs(t, err, ErrOverflow)

	require.Equal(t, MaxAmount, bank.GetAmount(0))
	require.EqualValues(t, 1, bank.GetAmount(1))
}

func TestUnderflowWithdraw(t *testing.T) {
	t.Parallel()

	bank := New(1)

	_, err := bank.Withdraw(0, 1)
	require.ErrorIs(t, err, ErrUnderflow)
	require.EqualValues(t, 0, bank.GetAmount(0))
}

func TestUnderflowTransfer(t *testing.T) {
	t.Parallel()

	bank := New(2)

	err := bank.Transfer(0, 1, 100)
	require.ErrorIs(t, err, ErrUnderflow)

	require.EqualValues(t, 0, bank.GetAmount(0))
	require.EqualValues(t, 0, bank.GetAmount(1))
}

func TestTransferDeadlock(t *testing.T) {
	require.Eventually(t, func() bool {
		const accounts = 2

		bank := New(accounts)
		bank.Deposit(0, 1_000_000)
		bank.Deposit(1, 1_000_000)

		wg := new(sync.WaitGroup)
		wg.Go(func() {
			for range 100_000 {
				bank.Transfer(0, 1, 100)
			}
		})

		wg.Go(func() {
			for range 100_000 {
				bank.Transfer(1, 0, 100)
			}
		})

		wg.Wait()
		return true
	}, 5*time.Second, 100*time.Millisecond, "fatal error: all goroutines are asleep - deadlock!")
}

func TestConsolidateDeadlock(t *testing.T) {
	require.Eventually(t, func() bool {
		const accounts = 3

		bank := New(accounts)
		bank.Deposit(0, 1_000_000)
		bank.Deposit(1, 1_000_000)

		wg := new(sync.WaitGroup)
		wg.Go(func() {
			for range 1_000_000 {
				bank.Deposit(0, 100)
				bank.TotalAmount()
			}
		})

		wg.Go(func() {
			for range 1_000_000 {
				bank.Deposit(1, 100)
				bank.TotalAmount()
			}
		})

		wg.Go(func() {
			for range 100_000 {
				bank.Consolidate([]int{1, 0}, 2)
			}
		})

		wg.Wait()
		return true
	}, 15*time.Second, 100*time.Millisecond, "fatal error: all goroutines are asleep - deadlock!")
}

func TestInvalidArguments(t *testing.T) {
	t.Parallel()

	bank := New(2)

	_, err := bank.Deposit(0, 0)
	require.ErrorIs(t, err, ErrInvalidAmount)

	_, err = bank.Withdraw(0, 0)
	require.ErrorIs(t, err, ErrInvalidAmount)

	err = bank.Transfer(0, 0, 100)
	require.Error(t, err)

	err = bank.Transfer(0, 1, -1)
	require.ErrorIs(t, err, ErrInvalidAmount)

	require.Error(t, bank.Consolidate([]int{}, 1))
	require.Error(t, bank.Consolidate([]int{1, 1}, 0))
	require.Error(t, bank.Consolidate([]int{1}, 1))
}

func TestConsolidateOverflow(t *testing.T) {
	t.Parallel()

	bank := New(3)

	_, err := bank.Deposit(0, MaxAmount)
	require.NoError(t, err)

	_, err = bank.Deposit(1, 1)
	require.NoError(t, err)

	err = bank.Consolidate([]int{0, 1}, 2)
	require.ErrorIs(t, err, ErrOverflow)

	require.EqualValues(t, MaxAmount, bank.GetAmount(0))
	require.EqualValues(t, 1, bank.GetAmount(1))
	require.EqualValues(t, 0, bank.GetAmount(2))
}

func TestStressMultithreaded(t *testing.T) {
	const (
		accounts        = 100
		workers         = 16
		amount          = 1000
		mod       int64 = 100
		sleepTime       = 2 * time.Second
	)

	bank := New(accounts)
	expected := make([]int64, accounts)
	for i := 0; i < accounts; i++ {
		bank.Deposit(i, 1_000_000_000)
		expected[i] = 1_000_000_000
	}

	failed := atomic.Bool{}
	wg := new(sync.WaitGroup)

	for i := 0; i < workers; i++ {
		wg.Go(func() {
			rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
			for !failed.Load() {
				i := rnd.Intn(accounts)
				amt := int64(rnd.Intn(amount)+1) / mod * mod
				op := rnd.Intn(4)
				switch op {
				case 0:
					bank.Deposit(i, amt)
					atomic.AddInt64(&expected[i], amt)
				case 1:
					bank.Withdraw(i, amt)
					atomic.AddInt64(&expected[i], -amt)
				case 2:
					j := rnd.Intn(accounts - 1)
					if j >= i {
						j++
					}
					bank.Transfer(i, j, amt)
					atomic.AddInt64(&expected[i], -amt)
					atomic.AddInt64(&expected[j], amt)
				case 3:
					_ = bank.GetAmount(i)
				}
			}
		})
	}

	time.Sleep(sleepTime)
	failed.Store(true)
	wg.Wait()

	var total int64
	for i := 0; i < accounts; i++ {
		actual := bank.GetAmount(i)
		total += actual
		require.Equal(t, expected[i], actual, "account[%d] mismatch", i)
	}
	require.Equal(t, total, bank.TotalAmount())
}

func TestNoChannels(t *testing.T) {
	filesToCheck := []string{
		"./bank.go",
	}

	for _, relPath := range filesToCheck {
		absPath, err := filepath.Abs(relPath)
		require.NoError(t, err)

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, absPath, nil, parser.AllErrors)
		require.NoError(t, err)

		ast.Inspect(node, func(n ast.Node) bool {
			if makeExpr, ok := n.(*ast.CallExpr); ok {
				if ident, ok := makeExpr.Fun.(*ast.Ident); ok && ident.Name == "make" {
					_, ok = makeExpr.Args[0].(*ast.ChanType)
					require.False(t, ok, "сhannels are prohibited in this assignment.")
				}
			}

			return true
		})
	}
}
