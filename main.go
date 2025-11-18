package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInsufficientFunds   = errors.New("недостаточно средств на счете")
	ErrInvalidAmount       = errors.New("некорректная сумма: должна быть положительной")
	ErrAccountNotFound     = errors.New("счёт не найден")
	ErrSameAccountTransfer = errors.New("нельзя перевести деньги на тот же счёт")
)

type Transaction struct {
	Type      string
	Amount    float64
	Timestamp time.Time
	ToFrom    string
}

type Account struct {
	ID      string
	Owner   string
	Balance float64
	History []Transaction
}

type AccountService interface {
	Deposit(amount float64) error
	Withdraw(amount float64) error
	Transfer(to *Account, amount float64) error
	GetBalance() float64
	GetStatement() string
}

type Storage interface {
	SaveAccount(account *Account) error
	LoadAccount(accountID string) (*Account, error)
	GetAllAccounts() ([]*Account, error)
}

type AccountServiceImpl struct {
	account *Account
}

func NewAccountService(account *Account) *AccountServiceImpl {
	return &AccountServiceImpl{account: account}
}

func (s *AccountServiceImpl) Deposit(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	s.account.Balance += amount
	s.account.History = append(s.account.History, Transaction{
		Type:      "deposit",
		Amount:    amount,
		Timestamp: time.Now(),
	})
	return nil
}

func (s *AccountServiceImpl) Withdraw(amount float64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > s.account.Balance {
		return ErrInsufficientFunds
	}
	s.account.Balance -= amount
	s.account.History = append(s.account.History, Transaction{
		Type:      "withdraw",
		Amount:    amount,
		Timestamp: time.Now(),
	})
	return nil
}

func (s *AccountServiceImpl) Transfer(to *Account, amount float64) error {
	if s.account.ID == to.ID {
		return ErrSameAccountTransfer
	}
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > s.account.Balance {
		return ErrInsufficientFunds
	}

	s.account.Balance -= amount
	to.Balance += amount

	now := time.Now()
	s.account.History = append(s.account.History, Transaction{
		Type:      "transfer_out",
		Amount:    amount,
		Timestamp: now,
		ToFrom:    to.ID,
	})
	to.History = append(to.History, Transaction{
		Type:      "transfer_in",
		Amount:    amount,
		Timestamp: now,
		ToFrom:    s.account.ID,
	})

	return nil
}

func (s *AccountServiceImpl) GetBalance() float64 {
	return s.account.Balance
}

func (s *AccountServiceImpl) GetStatement() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Выписка по счёту %s (%s):\n", s.account.ID, s.account.Owner))
	sb.WriteString(fmt.Sprintf("Текущий баланс: %.2f\n", s.account.Balance))
	sb.WriteString("История операций:\n")
	if len(s.account.History) == 0 {
		sb.WriteString("  (нет операций)\n")
	} else {
		for i, tx := range s.account.History {
			var desc string
			switch tx.Type {
			case "deposit":
				desc = fmt.Sprintf("Пополнение: +%.2f", tx.Amount)
			case "withdraw":
				desc = fmt.Sprintf("Снятие: -%.2f", tx.Amount)
			case "transfer_out":
				desc = fmt.Sprintf("Перевод на счёт %s: -%.2f", tx.ToFrom, tx.Amount)
			case "transfer_in":
				desc = fmt.Sprintf("Перевод со счёта %s: +%.2f", tx.ToFrom, tx.Amount)
			default:
				desc = fmt.Sprintf("Неизвестная операция (%s)", tx.Type)
			}
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s\n",
				i+1,
				tx.Timestamp.Format("02.01.2006 15:04"),
				desc))
		}
	}
	return sb.String()
}

type InMemoryStorage struct {
	accounts map[string]*Account
	nextID   int
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		accounts: make(map[string]*Account),
		nextID:   1,
	}
}

func (s *InMemoryStorage) SaveAccount(account *Account) error {
	if account.ID == "" {
		account.ID = fmt.Sprintf("ACC%04d", s.nextID)
		s.nextID++
	}
	s.accounts[account.ID] = account
	return nil
}

func (s *InMemoryStorage) LoadAccount(accountID string) (*Account, error) {
	if acc, ok := s.accounts[accountID]; ok {
		return acc, nil
	}
	return nil, ErrAccountNotFound
}

func (s *InMemoryStorage) GetAllAccounts() ([]*Account, error) {
	accounts := make([]*Account, 0, len(s.accounts))
	for _, acc := range s.accounts {
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func readInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func readFloat(prompt string) (float64, error) {
	for {
		input := readInput(prompt)
		if input == "" {
			return 0, errors.New("ввод отменён")
		}
		value, err := strconv.ParseFloat(input, 64)
		if err != nil {
			fmt.Println("Некорректное число. Попробуйте ещё раз.")
			continue
		}
		return value, nil
	}
}

func main() {
	storage := NewInMemoryStorage()

	fmt.Println("Добро пожаловать в консольное банковское приложение!")
	fmt.Println("Для завершения введите 'exit' в любом меню.")

	var currentAccount *Account

	for {
		if currentAccount == nil {
			fmt.Println("\n=== Главное меню ===")
			fmt.Println("1. Создать счёт")
			fmt.Println("2. Выбрать существующий счёт")
			fmt.Println("3. Список всех счетов")
			fmt.Println("0. Выйти")
			choice := readInput("Выберите действие: ")

			switch choice {
			case "1":
				owner := readInput("Введите имя владельца: ")
				if owner == "" {
					fmt.Println("Имя не может быть пустым.")
					continue
				}
				account := &Account{
					Owner:   owner,
					Balance: 0.0,
					History: []Transaction{},
				}
				err := storage.SaveAccount(account)
				if err != nil {
					fmt.Printf("Ошибка создания счёта: %v\n", err)
					continue
				}
				currentAccount = account
				fmt.Printf("Счёт %s создан для %s\n", account.ID, account.Owner)

			case "2":
				id := readInput("Введите ID счёта (например, ACC0001): ")
				if id == "" {
					continue
				}
				acc, err := storage.LoadAccount(id)
				if err != nil {
					fmt.Printf("%v\n", err)
					continue
				}
				currentAccount = acc
				fmt.Printf("Вы вошли в счёт %s (%s)\n", acc.ID, acc.Owner)

			case "3":
				accounts, err := storage.GetAllAccounts()
				if err != nil || len(accounts) == 0 {
					fmt.Println("Нет созданных счетов.")
				} else {
					fmt.Println("\n Все счета:")
					for _, acc := range accounts {
						fmt.Printf("  %s | %s | Баланс: %.2f\n",
							acc.ID, acc.Owner, acc.Balance)
					}
				}

			case "0", "exit":
				fmt.Println("Спасибо за использование! До свидания.")
				return

			default:
				fmt.Println("Неверный выбор. Попробуйте снова.")
			}

		} else {
			service := NewAccountService(currentAccount)
			fmt.Printf("\n=== Счёт %s (%s) ===\n", currentAccount.ID, currentAccount.Owner)
			fmt.Println("1. Пополнить счёт")
			fmt.Println("2. Снять средства")
			fmt.Println("3. Перевести на другой счёт")
			fmt.Println("4. Просмотреть баланс")
			fmt.Println("5. Получить выписку")
			fmt.Println("6. Выйти из счёта")
			choice := readInput("Выберите действие: ")

			switch choice {
			case "1":
				amount, err := readFloat("Введите сумму для пополнения: ")
				if err != nil {
					continue
				}
				err = service.Deposit(amount)
				if err != nil {
					fmt.Printf("%v\n", err)
				} else {
					fmt.Printf("Счёт пополнен на %.2f. Новый баланс: %.2f\n",
						amount, service.GetBalance())
				}

			case "2":
				amount, err := readFloat("Введите сумму для снятия: ")
				if err != nil {
					continue
				}
				err = service.Withdraw(amount)
				if err != nil {
					fmt.Printf("%v\n", err)
				} else {
					fmt.Printf("Снято %.2f. Новый баланс: %.2f\n",
						amount, service.GetBalance())
				}

			case "3":
				toID := readInput("Введите ID счёта-получателя: ")
				if toID == "" {
					continue
				}
				toAccount, err := storage.LoadAccount(toID)
				if err != nil {
					fmt.Printf("%v\n", err)
					continue
				}
				amount, err := readFloat("Введите сумму перевода: ")
				if err != nil {
					continue
				}
				err = service.Transfer(toAccount, amount)
				if err != nil {
					fmt.Printf("%v\n", err)
				} else {
					_ = storage.SaveAccount(toAccount)
					_ = storage.SaveAccount(currentAccount)
					fmt.Printf("Переведено %.2f на счёт %s. Новый баланс: %.2f\n",
						amount, toID, service.GetBalance())
				}

			case "4":
				fmt.Printf("Текущий баланс: %.2f\n", service.GetBalance())

			case "5":
				fmt.Println(service.GetStatement())

			case "6":
				fmt.Printf("Вы вышли из счёта %s\n", currentAccount.ID)
				currentAccount = nil

			case "0", "exit":
				fmt.Println("До свидания!")
				return

			default:
				fmt.Println("Неверный выбор.")
			}
		}
	}
}
