package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Petro-vich/transaction_processing_go/internal/models/transaction"
	"github.com/Petro-vich/transaction_processing_go/internal/storage"
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
        balance REAL DEFAULT 0.00
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
        amount REAL NOT NULL,
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

	if amount <= 0 {
		return fmt.Errorf("%s balanc must be positiv", op)
	}

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
	if errors.Is(err, sql.ErrNoRows) {
		return 0, storage.ErrAddressNotExist
	}
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
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

	balanceFrom, err := st.GetBalance(from)
	if err == storage.ErrAddressNotExist {
		return err
	} else if err != nil {
		return fmt.Errorf("%s: failed to get balance for to address: %w", op, err)
	}

	if balanceFrom-amount < 0 {
		return storage.ErrInsufficient
	}

	_, err = st.GetBalance(to)
	if err == storage.ErrAddressNotExist {
		return err
	} else if err != nil {
		return fmt.Errorf("%s: failed to get balance for to address: %w", op, err)
	}

	_, err = tx.Exec(`
	UPDATE wallet SET balance = balance - ?
	WHERE address = ?
	`, amount, from)
	if err != nil {
		return fmt.Errorf("%s: failed to update from balance: %w", op, err)
	}

	_, err = tx.Exec(`
		UPDATE wallet SET balance = balance + ?
		WHERE address = ?
	`, amount, to)
	if err != nil {
		return fmt.Errorf("%s: failed to update to balance: %w", op, err)
	}

	_, err = tx.Exec(`
	INSERT INTO transactions (from_address, to_address, amount, created_at)
	VALUES (?, ?, ?, ?)
	`, from, to, amount, time.Now())
	if err != nil {
		return fmt.Errorf("%s: failed to insert transaction: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: commit transaction: %w", op, err)
	}

	return nil
}

func (st *Storage) GetLast(count int) ([]transaction.Request, error) {
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

	transactions := []transaction.Request{}

	for rows.Next() {
		tr := transaction.Request{}
		err := rows.Scan(&tr.Id, &tr.From, &tr.To, &tr.Amount, &tr.Created_at)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows error: %w", op, err)
	}
	return transactions, nil
}

func (st *Storage) IsEmpty() bool {
	res, _ := st.db.Query(`
	SELECT *
	FROM wallet`)

	return !res.Next()
}
