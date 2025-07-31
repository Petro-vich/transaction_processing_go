package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Petro-vich/transaction_processing_go/internal/models/transaction"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(filepath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	defer func() {
		if err != nil {
			db.Close()
		}
	}()

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS wallet (
        id INTEGER PRIMARY KEY AUTOINCREMENT, 
        address TEXT NOT NULL UNIQUE CHECK(LENGTH(address) == 64),
        balance DECIMAL (15,6) DEFAULT 0.00
    )
`)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS transactions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        from_address TEXT NOT NULL, 
        to_address TEXT NOT NULL CHECK(LENGTH(to_address) == 64),
        amount DECIMAL (15, 2) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (from_address) REFERENCES wallet(address),
		    FOREIGN KEY (to_address) REFERENCES wallet(address) 
)
`)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (st *Storage) CreateWallet(adr string, amount float64) error {
	const op = "storage.sqlite.CreateWallet"

	if len(adr) != 64 {
		return fmt.Errorf("%s: invalid address length (expected 64, got %d)", op, len(adr))
	}
	stmt, err := st.db.Prepare(`
	INSERT INTO wallet (address, balance)
	VALUES (?, ?)
	`)
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	_, err = stmt.Exec(adr, amount)
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}
	return nil
}

func (st *Storage) GetBalance(address string) (float64, error) {
	//TODO: conver float64 to Decimal

	const op = "storage.sqlite.GetBalance"

	stmt, err := st.db.Prepare(`
	SELECT balance
	FROM wallet
	WHERE address = ?
	`)
	if err != nil {
		return 0, fmt.Errorf("%s, %w", op, err)
	}

	var balance float64
	err = stmt.QueryRow(address).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("%s, %w", op, err)
	}

	return balance, nil
}

func (st *Storage) SendMoney(from string, to string, amount float64) error {
	const op = "storage.sqlite.SendMoney"

	tx, err := st.db.Begin()
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}
	defer tx.Rollback()

	balance, err := st.GetBalance(from)
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	if balance-amount < 0 {
		return fmt.Errorf("%s", op, "Недостаточно средств")
	}

	_, err = tx.Exec(`
	UPDATE wallet SET balance = balance - ?
	WHERE address = ?
	`, amount, from)
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	_, err = tx.Exec(`
	UPDATE wallet SET balance = balance + ?
	WHERE address = ?
	`, amount, to)
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	_, err = tx.Exec(`
	INSERT INTO transactions (to_address, from_address, amount, created_at)
	VALUES (?, ?, ?, ?)
	`, from, to, amount, time.Now())
	if err != nil {
		return fmt.Errorf("%s, %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return nil
}

func (st *Storage) GetLast(count int) ([]string, error) {
	const op = "storage.sqlite.GetLast"

	rows, err := st.db.Query(`
	SELECT * 
	FROM transactions
	ORDER BY created_at DESC
	LIMIT ?
	`, count)
	if err != nil {
		return nil, fmt.Errorf("%s, %w", op, err)
	}

	trLast := []transaction.Request{}

	for rows.Next() {
		tr := transaction.Request{}
		err := rows.Scan(&tr.Id, &tr.From, &tr.To, &tr.Amount, &tr.Created_at)
		if err != nil {
			fmt.Println(err)
			continue
		}
		trLast = append(trLast, tr)
	}
	fmt.Println(trLast)

	return nil, nil
}
