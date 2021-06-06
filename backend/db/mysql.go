package db

import (
	"database/sql"
	"fmt"
	"github.com/syncloud/redirect/model"
	"log"
	"time"
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
	db.SetConnMaxIdleTime(time.Hour)
	mysql.db = db
}

func (mysql *MySql) Close() {
	defer mysql.db.Close()
}

func (mysql *MySql) GetUser(id int64) (*model.User, error) {
	return mysql.selectUserByField("id", id)
}

func (mysql *MySql) GetUserByEmail(email string) (*model.User, error) {
	return mysql.selectUserByField("email", email)
}

func (mysql *MySql) GetUserByUpdateToken(updateToken string) (*model.User, error) {
	return mysql.selectUserByField("update_token", updateToken)
}

func (mysql *MySql) selectUserByField(field string, value interface{}) (*model.User, error) {
	row := mysql.db.QueryRow(
		"SELECT "+
			"id, "+
			"email, "+
			"password_hash, "+
			"active, "+
			"update_token, "+
			"unsubscribed, "+
			"timestamp, "+
			"premium_status_id "+
			"FROM user "+
			"WHERE "+field+" = ?", value)

	user := &model.User{}
	err := row.Scan(&user.Id, &user.Email, &user.PasswordHash, &user.Active, &user.UpdateToken,
		&user.Unsubscribed, &user.Timestamp, &user.PremiumStatusId)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("no user found: %s=%s\n", field, value)
			return nil, nil
		} else {
			log.Printf("cannot scan user: %s=%s, error: %s\n", field, value, err)
			return nil, err
		}
	}

	return user, nil
}

func (mysql *MySql) InsertUser(user *model.User) (int64, error) {
	stmt, err := mysql.db.Prepare(
		"INSERT into user (" +
			"email, " +
			"password_hash, " +
			"active, " +
			"update_token, " +
			"unsubscribed, " +
			"timestamp, " +
			"premium_status_id " +
			") values (?,?,?,?,?,?,?)")
	if err != nil {
		log.Println("unable to insert user (prepare): ", err)
		return 0, err
	}
	res, err := stmt.Exec(
		user.Email,
		user.PasswordHash,
		user.Active,
		user.UpdateToken,
		user.Unsubscribed,
		user.Timestamp,
		user.PremiumStatusId,
	)
	if err != nil {
		log.Println("unable to insert user (exec): ", err)
		return 0, err
	}
	defer stmt.Close()
	return res.LastInsertId()
}

func (mysql *MySql) UpdateUser(user *model.User) error {
	stmt, err := mysql.db.Prepare(
		"UPDATE user SET " +
			"email = ?, " +
			"password_hash = ?, " +
			"active = ?, " +
			"update_token = ?, " +
			"unsubscribed = ?, " +
			"timestamp = ?, " +
			"premium_status_id = ? " +
			"WHERE id = ?")
	if err != nil {
		log.Println("sql error: ", err)
		return err
	}
	now := time.Now()
	_, err = stmt.Exec(
		user.Email,
		user.PasswordHash,
		user.Active,
		user.UpdateToken,
		user.Unsubscribed,
		&now,
		user.PremiumStatusId,
		user.Id,
	)
	if err != nil {
		log.Println("sql error: ", err)
		return err
	}
	defer stmt.Close()
	return nil
}

func (mysql *MySql) DeleteUser(userId int64) error {

	stmt, err := mysql.db.Prepare("DELETE FROM user WHERE id = ?")
	if err != nil {
		log.Println("Cannot delete user: ", userId, err)
		return fmt.Errorf("DB error")
	}
	defer stmt.Close()
	_, err = stmt.Exec(userId)
	if err != nil {
		log.Println("Cannot delete user: ", userId, err)
		return fmt.Errorf("DB error")
	}
	return nil
}

func (mysql *MySql) GetDomainByToken(token string) (*model.Domain, error) {
	return mysql.getDomainByField("update_token", token)
}

func (mysql *MySql) GetDomainByName(name string) (*model.Domain, error) {
	return mysql.getDomainByField("name", name)
}

func (mysql *MySql) getDomainByField(field string, value string) (*model.Domain, error) {
	row := mysql.db.QueryRow(
		"SELECT "+
			"id, "+
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
			"last_update, "+
			"name "+
			"FROM domain "+
			"WHERE "+field+" = ?", value)

	var mapLocalAddress *bool
	domain := &model.Domain{}
	err := row.Scan(
		&domain.Id,
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
		&domain.Name,
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

func (mysql *MySql) DeleteDomain(domainId uint64) error {
	stmt, err := mysql.db.Prepare("DELETE FROM domain WHERE id = ?")
	if err != nil {
		log.Println("Cannot delete domain (prepare): ", domainId, err)
		return fmt.Errorf("DB error")
	}
	defer stmt.Close()
	_, err = stmt.Exec(domainId)
	if err != nil {
		log.Println("Cannot delete domain (exec): ", domainId, err)
		return fmt.Errorf("DB error")
	}
	return nil
}

func (mysql *MySql) DeleteAllDomains(userId int64) error {

	stmt, err := mysql.db.Prepare("DELETE FROM domain WHERE user_id = ?")
	if err != nil {
		log.Println("Cannot delete domains for user: ", userId, err)
		return fmt.Errorf("DB error")
	}
	defer stmt.Close()
	_, err = stmt.Exec(userId)
	if err != nil {
		log.Println("Cannot delete domains for user: ", userId, err)
		return fmt.Errorf("DB error")
	}
	return nil
}

func (mysql *MySql) GetUserDomains(userId int64) ([]*model.Domain, error) {
	domains := make([]*model.Domain, 0)
	rows, err := mysql.db.Query(
		"SELECT "+
			"id, "+
			"name, "+
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
			"WHERE user_id = ?", userId)
	if err != nil {
		log.Println("Cannot select domains for user: ", userId, err)
		return nil, fmt.Errorf("DB error")
	}
	defer rows.Close()

	for rows.Next() {
		var mapLocalAddress *bool
		domain := &model.Domain{}
		err := rows.Scan(
			&domain.Id,
			&domain.Name,
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
			log.Println("Cannot scan domains for user: ", userId, err)
			return nil, fmt.Errorf("DB error")
		}
		if mapLocalAddress != nil {
			domain.MapLocalAddress = *mapLocalAddress
		} else {
			domain.MapLocalAddress = false
		}
		domains = append(domains, domain)
	}
	if err := rows.Err(); err != nil {
		log.Println("Cannot process domains for user: ", userId, err)
		return nil, fmt.Errorf("DB error")
	}
	return domains, nil
}

func (mysql *MySql) UpdateDomain(domain *model.Domain) error {
	stmt, err := mysql.db.Prepare(
		"UPDATE domain SET " +
			"name = ?, " +
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
		domain.Name,
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
			"name, " +
			"update_token, " +
			"user_id, " +
			"device_mac_address, " +
			"device_name, " +
			"device_title, " +
			"last_update" +
			") values (?,?,?,?,?,?,?)")
	if err != nil {
		log.Println("unable to insert domain (prepare): ", err)
		return err
	}
	_, err = stmt.Exec(
		domain.Name,
		domain.UpdateToken,
		domain.UserId,
		domain.DeviceMacAddress,
		domain.DeviceName,
		domain.DeviceTitle,
		domain.LastUpdate,
	)
	if err != nil {
		log.Println("unable to insert domain (exec): ", err)
		return err
	}
	defer stmt.Close()
	return nil
}

func (mysql *MySql) GetAction(userId int64, actionTypeId uint64) (*model.Action, error) {
	row := mysql.db.QueryRow(
		"SELECT "+
			"id, "+
			"action_type_id, "+
			"user_id, "+
			"token, "+
			"timestamp "+
			"FROM action "+
			"WHERE user_id = ? and action_type_id = ?", userId, actionTypeId)
	action := &model.Action{}
	err := row.Scan(&action.Id, &action.ActionTypeId, &action.UserId, &action.Token, &action.Timestamp)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		log.Println("Cannot scan an action: ", userId, actionTypeId, err)
		return nil, err
	default:
		return action, nil
	}

}

func (mysql *MySql) GetActionByToken(token string, actionTypeId uint64) (*model.Action, error) {
	row := mysql.db.QueryRow(
		"SELECT "+
			"id, "+
			"action_type_id, "+
			"user_id, "+
			"token, "+
			"timestamp "+
			"FROM action "+
			"WHERE token = ? and action_type_id = ?", token, actionTypeId)
	action := &model.Action{}
	err := row.Scan(&action.Id, &action.ActionTypeId, &action.UserId, &action.Token, &action.Timestamp)
	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		log.Println("Cannot scan an action: ", token, actionTypeId, err)
		return nil, err
	default:
		return action, nil
	}

}

func (mysql *MySql) InsertAction(action *model.Action) error {
	stmt, err := mysql.db.Prepare(
		"INSERT into action (" +
			"action_type_id, " +
			"user_id, " +
			"token, " +
			"timestamp" +
			") values (?,?,?,?)")
	if err != nil {
		log.Println("unable to insert action (prepare): ", err)
		return err
	}
	_, err = stmt.Exec(
		action.ActionTypeId,
		action.UserId,
		action.Token,
		action.Timestamp,
	)
	if err != nil {
		log.Println("unable to insert action (exec): ", err)
		return err
	}
	defer stmt.Close()
	return nil

}

func (mysql *MySql) UpdateAction(action *model.Action) error {
	stmt, err := mysql.db.Prepare(
		"UPDATE action SET " +
			"action_type_id = ?, " +
			"user_id = ?, " +
			"token = ?, " +
			"timestamp = ? " +
			"WHERE id = ?")
	if err != nil {
		log.Println("unable to update action (prepare): ", err)
		return err
	}
	_, err = stmt.Exec(
		action.ActionTypeId,
		action.UserId,
		action.Token,
		action.Timestamp,
		action.Id,
	)
	if err != nil {
		log.Println("unable to update action (exec): ", err)
		return err
	}
	defer stmt.Close()
	return nil

}

func (mysql *MySql) DeleteActions(userId int64) error {

	stmt, err := mysql.db.Prepare("DELETE FROM action WHERE user_id = ?")
	if err != nil {
		log.Println("Cannot delete actions for user (prepare): ", userId, err)
		return fmt.Errorf("DB error")
	}
	defer stmt.Close()
	_, err = stmt.Exec(userId)
	if err != nil {
		log.Println("Cannot delete actions for user (exec): ", userId, err)
		return fmt.Errorf("DB error")
	}
	return nil
}

func (mysql *MySql) DeleteAction(actionId uint64) error {

	stmt, err := mysql.db.Prepare("DELETE FROM action WHERE id = ?")
	if err != nil {
		log.Println("Cannot delete action (prepare): ", actionId, err)
		return fmt.Errorf("DB error")
	}
	defer stmt.Close()
	_, err = stmt.Exec(actionId)
	if err != nil {
		log.Println("Cannot delete action (exec): ", actionId, err)
		return fmt.Errorf("DB error")
	}
	return nil
}

func (mysql *MySql) getDomainsLastUpdatedBefore() error {
	//return self.session.query(Domain).filter(Domain.last_update < date).filter(Domain.ip != None).order_by(Domain.last_update).limit(limit)
	return fmt.Errorf("not implemented")
}
