package storage

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/lionslon/go-keepass/internal/models"
)

const (
	checkUserExist = `SELECT COUNT(*) FROM users WHERE login = $1`
	createUser     = `INSERT INTO users (login, password) VALUES($1,$2)`
	getUser        = `SELECT id, password FROM users WHERE login = $1`
	addData        = `INSERT INTO data (user_id, data_id, data) VALUES($1,$2, $3)`
	getData        = `SELECT data FROM data WHERE user_id = $1 AND data_id = $2`
	deleteData     = `DELETE FROM data WHERE user_id = $1 AND data_id = $2`
)

type KeeperStorage struct {
	conn *sql.DB
}

func NewKeeperStorage(dns string) (*KeeperStorage, error) {
	conn, err := sql.Open("pgx", dns)
	if err != nil {
		return nil, fmt.Errorf("cannot create connection db: %w", err)
	}

	storage := &KeeperStorage{conn: conn}
	if err := storage.applyDBMigrations(context.Background()); err != nil {
		return nil, fmt.Errorf("cannot apply migrations: %w", err)
	}
	return storage, nil
}

func (m *KeeperStorage) applyDBMigrations(ctx context.Context) error {
	tx, err := m.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}

	defer tx.Rollback()
	// это для возможности генерации uuid
	_, err = tx.ExecContext(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	if err != nil {
		return fmt.Errorf("cannot create uuid extension: %w", err)
	}

	// создаём таблицу для хранения пользователей
	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS users (
			id uuid DEFAULT uuid_generate_v4 (),
			login VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255),
			PRIMARY KEY (id)
        )
    `)
	if err != nil {
		return fmt.Errorf("cannot create users table: %w", err)
	}

	// создаём таблицу для хранения данных пользователя
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS data (
			id uuid DEFAULT uuid_generate_v4 (),
			user_id uuid,
			data_id VARCHAR(255) NOT NULL,
			data BYTEA,
			PRIMARY KEY (id),
			FOREIGN KEY (user_id) REFERENCES users(id),
			UNIQUE (user_id, data_id)
			)
    `)
	if err != nil {
		return fmt.Errorf("cannot create orders table: %w", err)
	}

	// коммитим транзакцию
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("cannot comit transaction: %w", err)
	}
	return nil
}

func (m *KeeperStorage) Close() {
	m.conn.Close()
}

func (m *KeeperStorage) IsUserExist(ctx context.Context, login string) bool {
	var count int
	row := m.conn.QueryRowContext(ctx, checkUserExist, login)
	row.Scan(&count)

	return count > 0
}

func (m *KeeperStorage) CreateUser(ctx context.Context, dto models.AuthDTO) (string, error) {

	if err := dto.GeneratePasswordHash(); err != nil {
		return ``, fmt.Errorf("cannot generate password hash: %w", err)
	}

	_, err := m.conn.ExecContext(ctx, createUser, dto.Login, dto.Password)
	if err != nil {
		return ``, fmt.Errorf("cannot execute create request: %w", err)
	}

	var uuid, password string
	row := m.conn.QueryRowContext(ctx, getUser, dto.Login)
	err = row.Scan(&uuid, &password)
	if err != nil {
		return ``, fmt.Errorf("cannot get created user id: %w", err)
	}

	return uuid, nil
}

func (m *KeeperStorage) Login(ctx context.Context, dto models.AuthDTO) (string, error) {
	var passwordHash string
	var uuid string

	row := m.conn.QueryRowContext(ctx, getUser, dto.Login)
	err := row.Scan(&uuid, &passwordHash)
	if err != nil {
		return ``, fmt.Errorf("cannot get user: %w", err)
	}

	if !dto.CheckPassword(passwordHash) {
		return ``, fmt.Errorf("bad password")
	}

	return uuid, nil
}

func (m *KeeperStorage) AddData(ctx context.Context, userId string, dataId string, data []byte) error {

	_, err := m.conn.ExecContext(ctx, addData, userId, dataId, data)
	if err != nil {
		return fmt.Errorf("cannot execute add new data: %w", err)
	}

	return nil
}

func (m *KeeperStorage) GetData(ctx context.Context, userId string, dataId string) ([]byte, error) {

	var data []byte

	row := m.conn.QueryRowContext(ctx, getData, userId, dataId)
	err := row.Scan(&data)
	if err != nil {
		return nil, fmt.Errorf("cannot scan data: %w", err)
	}

	return data, nil
}

func (m *KeeperStorage) DeleteData(ctx context.Context, userId string, dataId string) error {

	_, err := m.conn.ExecContext(ctx, deleteData, userId, dataId)
	if err != nil {
		return fmt.Errorf("cannot execute delete user data: %w", err)
	}

	return nil
}
