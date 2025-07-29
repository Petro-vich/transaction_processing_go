package sqlite

import (
	"database/sql"
	"fmt"

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
