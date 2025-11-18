---

# Практическая работа №7/8  
## «Банковское приложение» (Ошибки, интерфейсы)

---

## Общая структура

Программа реализована в **одном файле `main.go`**, но логически разделена на следующие блоки:

1. **Кастомные ошибки**  
2. **Модели данных** (`Account`, `Transaction`)  
3. **Интерфейсы** (`AccountService`, `Storage`)  
4. **Реализация интерфейсов**  
5. **Вспомогательные функции ввода**  
6. **Главный цикл приложения (меню)**

---

## Объяснение кода

### 1. Импорты и кастомные ошибки

```go
package main

import (
	"bufio"   // для построчного чтения ввода
	"errors"  // для создания кастомных ошибок
	"fmt"
	"os"      // доступ к stdin/stdout
	"strconv" // преобразование строки в число
	"strings" // работа со строками
	"time"    // для отметки времени операций
)
```
- Используются стандартные пакеты Go. Никаких внешних зависимостей.

```go
var (
	ErrInsufficientFunds   = errors.New("недостаточно средств на счете")
	ErrInvalidAmount       = errors.New("некорректная сумма: должна быть положительной")
	ErrAccountNotFound     = errors.New("счёт не найден")
	ErrSameAccountTransfer = errors.New("нельзя перевести деньги на тот же счёт")
)
```
- 4 кастомные ошибки созданы с помощью `errors.New()`.

---

### 2. Модели данных

```go
type Transaction struct {
	Type      string    // тип операции
	Amount    float64   // сумма
	Timestamp time.Time // дата и время
	ToFrom    string    // ID другого счёта (для переводов)
}
```
- Хранит одну запись в истории. Поддерживает 4 типа:  
 - `deposit` — пополнение  
 - `withdraw` — снятие  
 - `transfer_in` — входящий перевод  
 - `transfer_out` — исходящий перевод  

```go
type Account struct {
	ID      string        // уникальный идентификатор (генерируется как ACC0001)
	Owner   string        // имя владельца
	Balance float64       // текущий баланс
	History []Transaction // массив операций
}
```
- Основная сущность — банковский счёт.

---

### 3. Интерфейсы

```go
type AccountService interface {
	Deposit(amount float64) error
	Withdraw(amount float64) error
	Transfer(to *Account, amount float64) error
	GetBalance() float64
	GetStatement() string
}
```
- Интерфейс `AccountService` содержит 5 методов.

```go
type Storage interface {
	SaveAccount(account *Account) error
	LoadAccount(accountID string) (*Account, error)
	GetAllAccounts() ([]*Account, error)
}
```
- Интерфейс `Storage` содержит 3 метода для работы с данными.

---

### 4. Реализация `AccountService`

```go
type AccountServiceImpl struct {
	account *Account // ссылка на счёт, с которым работаем
}

func NewAccountService(account *Account) *AccountServiceImpl {
	return &AccountServiceImpl{account: account}
}
```
- Обёртка над `*Account`, реализующая интерфейс.
- Принцип: не изменять структуру `Account`, а добавлять поведение через сервис.

#### `Deposit(amount float64) error`

```go
if amount <= 0 {
	return ErrInvalidAmount
}
s.account.Balance += amount
s.account.History = append(s.account.History, Transaction{...})
```
- Проверка на корректную сумму  
- Увеличение баланса
- Добавление записи в историю  

#### `Withdraw(amount float64) error`

```go
if amount > s.account.Balance {
	return ErrInsufficientFunds
}
s.account.Balance -= amount
// ... добавление в историю
```
- Аналогично, но с проверкой **недостатка средств**.

#### `Transfer(to *Account, amount float64) error`

```go
if s.account.ID == to.ID {
	return ErrSameAccountTransfer
}
// ... проверка суммы и баланса
s.account.Balance -= amount
to.Balance += amount
// ... логирование в ОБОИХ счетах
``` 
- запрещён перевод на тот же счёт  
- атомарное изменение балансов  
- история обновляется у отправителя и получателя  

#### `GetBalance()` и `GetStatement()`

```go
func (s *AccountServiceImpl) GetBalance() float64 {
	return s.account.Balance
}
```
- Простой getter.

```go
func (s *AccountServiceImpl) GetStatement() string {
	var sb strings.Builder
	// Форматированный вывод: ID, владелец, баланс, история с датами
	// Пример строки: "1. [18.11.2025 14:30] Перевод на счёт ACC0002: -500.00"
}
```
**Выписка включает:**
- заголовок  
- текущий баланс  
- нумерованный список операций с датой и описанием  

---

### 5. Реализация `Storage` (InMemory)

```go
type InMemoryStorage struct {
	accounts map[string]*Account // хранение в памяти (ключ — ID)
	nextID   int                 // счётчик для генерации ACC0001, ACC0002...
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		accounts: make(map[string]*Account),
		nextID:   1,
	}
}
```
- Простое хранилище в оперативной памяти (без файлов/БД — по умолчанию для лабораторной).

#### `SaveAccount`

```go
if account.ID == "" {
	account.ID = fmt.Sprintf("ACC%04d", s.nextID)
	s.nextID++
}
s.accounts[account.ID] = account
```
- Автоматическая генерация ID при первом сохранении.

#### `LoadAccount` и `GetAllAccounts`

```go
// LoadAccount: возвращает счёт по ID или ErrAccountNotFound
// GetAllAccounts: возвращает срез всех счетов (для меню "Список всех счетов")
```

---

### 6. Вспомогательные функции ввода

```go
func readInput(prompt string) string {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}
```
- Безопасное чтение строки из консоли.

```go
func readFloat(prompt string) (float64, error) {
	// Цикл с повторным вводом при ошибке
	value, err := strconv.ParseFloat(input, 64)
	// ...
}
```
- Гарантирует, что пользователь введёт **корректное число**.

---

### 7. Главный цикл (`main`)

#### Инициализация

```go
storage := NewInMemoryStorage()
var currentAccount *Account // nil = не вошли в счёт
```

#### Главное меню (когда `currentAccount == nil`)

| Пункт | Действие |
|------|----------|
| `1` | Создаёт новый счёт → генерирует ID → сохраняет в `storage` → устанавливает как текущий |
| `2` | Загружает счёт по ID → проверяет ошибку → устанавливает как текущий |
| `3` | Выводит список всех счетов (ID, владелец, баланс) |
| `0`/`exit` | Корректный выход из программы |

#### Меню счёта (когда `currentAccount != nil`)

| Пункт | Вызывает метод `AccountService` | Обработка ошибок |
|------|------------------------------|------------------|
| `1` | `Deposit()` | Вывод `ErrInvalidAmount` |
| `2` | `Withdraw()` | Вывод `ErrInvalidAmount`, `ErrInsufficientFunds` |
| `3` | `Transfer()` | Вывод всех 4 ошибок из ТЗ |
| `4` | `GetBalance()` | Просто печать |
| `5` | `GetStatement()` | Форматированный вывод истории |
| `6` | `currentAccount = nil` | Возврат в главное меню |

**Все операции обновляют хранилище после перевода:**
 ```go
 _ = storage.SaveAccount(toAccount)
 _ = storage.SaveAccount(currentAccount)
 ```

---

## Пример работы

```
Добро пожаловать в консольное банковское приложение!

=== Главное меню ===
1. Создать счёт
2. Выбрать существующий счёт
3. Список всех счетов
0. Выйти
Выберите действие: 1
Введите имя владельца: Анастасия
Счёт ACC0001 создан для Анастасия

=== Счёт ACC0001 (Анастасия) ===
1. Пополнить счёт
...
Выберите действие: 1
Введите сумму для пополнения: 1000
Счёт пополнен на 1000.00. Новый баланс: 1000.00

Выберите действие: 5
Выписка по счёту ACC0001 (Анастасия):
Текущий баланс: 1000.00
История операций:
  1. [18.11.2025 14:30] Пополнение: +1000.00
```
