package db

import (
	"database/sql"
	"log"
)
import _ "github.com/go-sql-driver/mysql"

type MySql struct {
	db *sql.DB
}

func NewMySql() *MySql {
	return &MySql{}
}

func (mysql *MySql) Connect(database string, user string, password string) {

	db, err := sql.Open("mysql", user+":"+password+"@/"+database)
	if err != nil {
		log.Println("Cannot connect to db: ", err)
	}
	mysql.db = db
}

func (mysql *MySql) Close() {
	defer mysql.db.Close()
}

func (mysql *MySql) GetUser(email string) User {
	rows, err := mysql.db.Query(
		"SELECT id, email, password_hash, active, update_token, unsubscribed, timestamp, is_premium "+
			"FROM user "+
			"WHERE email = ?", email)
	if err != nil {
		log.Println("Cannot query a user: ", email, err)
	}
	defer rows.Close()

	user := User{}
	for rows.Next() {
		err := rows.Scan(&user.id, &user.email, &user.passwordHash, &user.active, &user.updateToken,
			&user.unsubscribed, &user.timestamp, &user.isPremium)
		if err != nil {
			log.Println("Cannot scan a user: ", email, err)
		}
	}

	err = rows.Err()
	if err != nil {
		log.Println("Rows error: ", err)
	}

	return user
}
