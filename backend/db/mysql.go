package db

import (
	"database/sql"
	"fmt"
	"github.com/syncloud/redirect/model"
	"log"
)
import _ "github.com/go-sql-driver/mysql"

type MySql struct {
	db *sql.DB
}

func NewMySql() *MySql {
	return &MySql{}
}

func (mysql *MySql) Connect(host string, database string, user string, password string) {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", user, password, host, database))
	if err != nil {
		log.Println("Cannot connect to db: ", err)
	}
	mysql.db = db
}

func (mysql *MySql) Close() {
	defer mysql.db.Close()
}

func (mysql *MySql) GetUser(id uint64) (*model.User, error) {
	row := mysql.db.QueryRow(
		"SELECT "+
			"id, "+
			"email, "+
			"password_hash, "+
			"active, "+
			"update_token, "+
			"unsubscribed, "+
			"timestamp "+
			"FROM user "+
			"WHERE id = ?", id)
	user := &model.User{}
	err := row.Scan(&user.Id, &user.Email, &user.PasswordHash, &user.Active, &user.UpdateToken,
		&user.Unsubscribed, &user.Timestamp)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		log.Println("Cannot scan a user: ", id, err)
		return nil, err
	default:
		return user, nil
	}

}

func (mysql *MySql) GetUserByEmail(email string) (*model.User, error) {
	row := mysql.db.QueryRow(
		"SELECT "+
			"id, "+
			"email, "+
			"password_hash, "+
			"active, "+
			"update_token, "+
			"unsubscribed, "+
			"timestamp "+
			"FROM user "+
			"WHERE email = ?", email)

	user := &model.User{}
	err := row.Scan(&user.Id, &user.Email, &user.PasswordHash, &user.Active, &user.UpdateToken,
		&user.Unsubscribed, &user.Timestamp)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			log.Println("Cannot scan a user: ", user, err)
			return nil, fmt.Errorf("DB error")
		}
	}

	return user, nil
}

func (mysql *MySql) GetPremiumAccount(userId uint64) (*model.PremiumAccount, error) {
	row := mysql.db.QueryRow(
		"SELECT "+
			"id, "+
			"user_id "+
			"FROM premium_user "+
			"WHERE user_id = ?", userId)

	account := &model.PremiumAccount{}
	err := row.Scan(&account.Id, &account.UserId)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			log.Println("Cannot scan an account: ", account, err)
			return nil, fmt.Errorf("DB error")
		}
	}

	return account, nil
}

func (mysql *MySql) GetDomainByToken(token string) (*model.Domain, error) {
	row := mysql.db.QueryRow(
		"SELECT "+
			"id, "+
			"user_domain, "+
			"ip, "+
			"ipv6, "+
			"dkim_key, "+
			"local_ip, "+
			"map_local_address, "+
			"update_token, "+
			"user_id, "+
			"device_mac_address, "+
			"device_name, "+
			"device_title, "+
			"platform_version, "+
			"web_protocol, "+
			"web_port, "+
			"web_local_port, "+
			"last_update "+
			"FROM domain "+
			"WHERE update_token = ?", token)

	var mapLocalAddress *bool
	domain := &model.Domain{}
	err := row.Scan(
		&domain.Id,
		&domain.UserDomain,
		&domain.Ip,
		&domain.Ipv6,
		&domain.DkimKey,
		&domain.LocalIp,
		&mapLocalAddress,
		&domain.UpdateToken,
		&domain.UserId,
		&domain.DeviceMacAddress,
		&domain.DeviceName,
		&domain.DeviceTitle,
		&domain.PlatformVersion,
		&domain.WebProtocol,
		&domain.WebPort,
		&domain.WebLocalPort,
		&domain.LastUpdate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			log.Println("Cannot scan a domain: ", domain, err)
			return nil, fmt.Errorf("DB error")
		}
	}
	if mapLocalAddress != nil {
		domain.MapLocalAddress = *mapLocalAddress
	} else {
		domain.MapLocalAddress = false
	}

	return domain, nil
}

func (mysql *MySql) GetDomainByUserDomain(userDomain string) (*model.Domain, error) {
	row := mysql.db.QueryRow(
		"SELECT "+
			"id, "+
			"user_domain, "+
			"ip, "+
			"ipv6, "+
			"dkim_key, "+
			"local_ip, "+
			"map_local_address, "+
			"update_token, "+
			"user_id, "+
			"device_mac_address, "+
			"device_name, "+
			"device_title, "+
			"platform_version, "+
			"web_protocol, "+
			"web_port, "+
			"web_local_port, "+
			"last_update "+
			"FROM domain "+
			"WHERE user_domain = ?", userDomain)

	var mapLocalAddress *bool
	domain := &model.Domain{}
	err := row.Scan(
		&domain.Id,
		&domain.UserDomain,
		&domain.Ip,
		&domain.Ipv6,
		&domain.DkimKey,
		&domain.LocalIp,
		&mapLocalAddress,
		&domain.UpdateToken,
		&domain.UserId,
		&domain.DeviceMacAddress,
		&domain.DeviceName,
		&domain.DeviceTitle,
		&domain.PlatformVersion,
		&domain.WebProtocol,
		&domain.WebPort,
		&domain.WebLocalPort,
		&domain.LastUpdate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			log.Println("Cannot scan a domain: ", domain, err)
			return nil, fmt.Errorf("DB error")
		}
	}
	if mapLocalAddress != nil {
		domain.MapLocalAddress = *mapLocalAddress
	} else {
		domain.MapLocalAddress = false
	}

	return domain, nil
}

func (mysql *MySql) UpdateDomain(domain *model.Domain) error {
	stmt, err := mysql.db.Prepare(
		"UPDATE domain SET " +
			"user_domain = ?, " +
			"ip = ?, " +
			"ipv6 = ?, " +
			"dkim_key = ?, " +
			"local_ip = ?, " +
			"map_local_address = ?, " +
			"update_token = ?, " +
			"user_id = ?, " +
			"device_mac_address = ?, " +
			"device_name = ?, " +
			"device_title = ?, " +
			"platform_version = ?, " +
			"web_protocol = ?, " +
			"web_port = ?, " +
			"web_local_port = ?, " +
			"last_update = ? " +
			"WHERE id = ?")
	if err != nil {
		log.Println("sql error: ", err)
		return err
	}
	_, err = stmt.Exec(
		domain.UserDomain,
		domain.Ip,
		domain.Ipv6,
		domain.DkimKey,
		domain.LocalIp,
		domain.MapLocalAddress,
		domain.UpdateToken,
		domain.UserId,
		domain.DeviceMacAddress,
		domain.DeviceName,
		domain.DeviceTitle,
		domain.PlatformVersion,
		domain.WebProtocol,
		domain.WebPort,
		domain.WebLocalPort,
		domain.LastUpdate,
		domain.Id,
	)
	if err != nil {
		log.Println("sql error: ", err)
		return err
	}
	defer stmt.Close()
	return nil
}

func (mysql *MySql) InsertDomain(domain *model.Domain) error {
	stmt, err := mysql.db.Prepare(
		"INSERT into domain (" +
			"user_domain, " +
			"update_token, " +
			"user_id, " +
			"device_mac_address, " +
			"device_name, " +
			"device_title, " +
			"last_update" +
			") values (?,?,?,?,?,?,?)")
	if err != nil {
		log.Println("sql error: ", err)
		return err
	}
	_, err = stmt.Exec(
		domain.UserDomain,
		domain.UpdateToken,
		domain.UserId,
		domain.DeviceMacAddress,
		domain.DeviceName,
		domain.DeviceTitle,
		domain.LastUpdate,
	)
	if err != nil {
		log.Println("sql error: ", err)
		return err
	}
	defer stmt.Close()
	return nil
}
