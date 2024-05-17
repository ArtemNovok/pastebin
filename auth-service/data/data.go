package data

import (
	"database/sql"
)

var DB *sql.DB

type User struct {
	Id       int64  `json:"id,omitempty"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewDB(db *sql.DB) error {
	DB = db
	err := CreateUserTable()
	if err != nil {
		return err
	}
	return nil
}

func CreateUserTable() error {
	query := `create table if not exists users(
		id serial primary key,
		username varchar(200) unique, 
		email varchar(200) unique, 
		password varchar(500)
	)`
	stmt, err := DB.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec()
	if err != nil {
		return err
	}
	return nil
}

func (u *User) SaveUser() error {
	query := `insert into users (username, email, password) values($1, $2, $3)`
	stmt, err := DB.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(u.UserName, u.Email, u.Password)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) GetUserByEmail() (User, error) {
	var user User
	query := `select * from users where email = $1`
	res := DB.QueryRow(query, u.Email)
	err := res.Err()
	if err != nil {
		return User{}, err
	}
	if err = res.Scan(&user.Id, &user.UserName, &user.Email, &user.Password); err != nil {
		return User{}, err
	}
	return user, nil
}
