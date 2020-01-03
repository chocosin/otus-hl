package storage

import (
	"database/sql"
	"fmt"
	"github.com/chocosin/otus-hl/social/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"os"
)

const dbName = "social"

type MysqlStorage struct {
	db *sql.DB

	insertUserSt       *sql.Stmt
	findByUsernameSt   *sql.Stmt
	getUserSt          *sql.Stmt
	deleteTokenSt      *sql.Stmt
	getTokenSt         *sql.Stmt
	insertTokenSt      *sql.Stmt
	getLatestUsernames *sql.Stmt
}

func (m *MysqlStorage) Close() error {
	return m.db.Close()
}

func (m *MysqlStorage) prepareStatements() error {
	var err error
	m.insertUserSt, err = m.db.Prepare(`
	insert into users(id, username, password, firstName, lastName, age, gender, interests, city) 
	values (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	m.findByUsernameSt, err = m.db.Prepare(`
	select id, username, password, firstName, lastName, age, gender, interests, city from users where username=?
	`)
	if err != nil {
		return err
	}
	m.getUserSt, err = m.db.Prepare(`
	select id, username, password, firstName, lastName, age, gender, interests, city from users where id=?
	`)
	if err != nil {
		return err
	}
	m.getLatestUsernames, err = m.db.Prepare(`
	select username from users order by id desc limit ?
	`)
	if err != nil {
		return err
	}
	if m.insertTokenSt, err = m.db.Prepare(`
	insert into auth_tokens(token, userID) values (?, ?)
	`); err != nil {
		return err
	}
	if m.deleteTokenSt, err = m.db.Prepare(`
	delete from auth_tokens where token=?
	`); err != nil {
		return err
	}
	if m.getTokenSt, err = m.db.Prepare(`
	select token, userID from auth_tokens where token=?
	`); err != nil {
		return err
	}
	return nil
}

func (m *MysqlStorage) InsertToken(token uuid.UUID, userId uuid.UUID) error {
	_, err := m.insertTokenSt.Exec(token.String(), userId.String())
	if err != nil {
		return errors.Wrap(err, "failed to insert token")
	}
	return nil
}

func (m *MysqlStorage) DeleteToken(id uuid.UUID) error {
	_, err := m.deleteTokenSt.Exec(id.String())
	if err != nil {
		return errors.Wrap(err, "failed to delete token")
	}
	return nil
}

func (m *MysqlStorage) GetUserByToken(token uuid.UUID) (*model.User, error) {
	row := m.getTokenSt.QueryRow(token.String())
	var dbToken, dbUserID string
	err := row.Scan(&dbToken, &dbUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to GetUserByToken")
	}
	user, err := m.getUser(dbUserID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to GetUserByToken")
	}
	return user, nil
}

func (m *MysqlStorage) InsertUser(user *model.User) error {
	// for now storing UUID as string
	_, err := m.insertUserSt.Exec(user.ID.String(), user.Username, user.PasswordHash,
		user.FirstName, user.LastName, user.Age, user.Gender, user.JoinInterests(), user.City)
	if err != nil {
		return errors.Wrap(err, "failed to insert user")
	}
	return nil
}

func (m *MysqlStorage) getUser(userID string) (*model.User, error) {
	row := m.getUserSt.QueryRow(userID)
	user, err := m.scanUser(row)
	if err != nil {
		return nil, errors.Wrap(err, "getUser")
	}
	return user, nil
}

func (m *MysqlStorage) FindUserByUsername(username string) (*model.User, error) {
	row := m.findByUsernameSt.QueryRow(username)
	user, err := m.scanUser(row)
	if err != nil {
		return nil, errors.Wrap(err, "FindUserByUsername")
	}
	return user, nil
}

const lastUsernamesLimit = 10

func (m *MysqlStorage) LastUsernames() ([]string, error) {
	rows, err := m.getLatestUsernames.Query(lastUsernamesLimit)
	if err != nil {
		return nil, errors.Wrap(err, "LastRegistered")
	}
	defer rows.Close()

	usernames := make([]string, 0, lastUsernamesLimit)
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, errors.Wrap(err, "LastRegistered")
		}
		usernames = append(usernames, username)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "LastRegistered")
	}
	return usernames, nil
}

func (m *MysqlStorage) scanUser(row *sql.Row) (*model.User, error) {
	var u model.User
	var idStr, interestsJoined string
	err := row.Scan(&idStr, &u.Username, &u.PasswordHash, &u.FirstName, &u.LastName,
		&u.Age, &u.Gender, &interestsJoined, &u.City)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if u.ID, err = uuid.FromString(idStr); err != nil {
		return nil, errors.Wrap(err, "failed to parse user id")
	}
	u.SetInterests(interestsJoined)

	return &u, nil
}

func NewMysqlStorage(config *MysqlConfig) (*MysqlStorage, error) {
	db, err := sql.Open("mysql", config.dsn())
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	storage := &MysqlStorage{db: db}
	err = storage.prepareStatements()
	return storage, err
}

type MysqlConfig struct {
	host     string
	username string
	password string
	dbName   string
}

func (mc *MysqlConfig) dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true&timeout=6s",
		mc.username, mc.password, mc.host, mc.dbName)
}

func NewMysqlConfig() *MysqlConfig {
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = "localhost"
	}
	username := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASS")
	return &MysqlConfig{
		host:     host,
		username: username,
		password: password,
		dbName:   dbName,
	}
}
