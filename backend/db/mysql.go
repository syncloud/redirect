package db

import (
	"database/sql"
	"fmt"
	"github.com/syncloud/redirect/model"
	"log"
	"strings"
	"time"
)
import _ "github.com/go-sql-driver/mysql"

type MySql struct {
	host     string
	database string
	user     string
	password string
	db       *sql.DB
}

func NewMySql(host string, database string, user string, password string) *MySql {
	return &MySql{
		host:     host,
		database: database,
		user:     user,
		password: password,
	}
}

func (m *MySql) Start() error {
	db, err := sql.Open(
		"mysql",
		fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", m.user, m.password, m.host, m.database),
	)
	if err != nil {
		return fmt.Errorf("cannot connect to db: %v", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	m.db = db
	return nil
}

func (m *MySql) Close() {
	defer m.db.Close()
}

func (m *MySql) GetUser(id int64) (*model.User, error) {
	return m.selectUserByField("id", id)
}

func (m *MySql) GetUserByEmail(email string) (*model.User, error) {
	return m.selectUserByField("email", email)
}

func (m *MySql) GetUserByUpdateToken(updateToken string) (*model.User, error) {
	return m.selectUserByField("update_token", updateToken)
}

func (m *MySql) selectUserByField(field string, value interface{}) (*model.User, error) {
	row := m.db.QueryRow(
		"SELECT "+
			"id, "+
			"email, "+
			"password_hash, "+
			"active, "+
			"update_token, "+
			"notification_enabled, "+
			"timestamp, "+
			"subscription_id "+
			"FROM user "+
			"WHERE "+field+" = ?", value)

	user := &model.User{}
	err := row.Scan(&user.Id, &user.Email, &user.PasswordHash, &user.Active, &user.UpdateToken,
		&user.NotificationEnabled, &user.Timestamp, &user.SubscriptionId)

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

func (m *MySql) InsertUser(user *model.User) (int64, error) {
	stmt, err := m.db.Prepare(
		"INSERT into user (" +
			"email, " +
			"password_hash, " +
			"active, " +
			"update_token, " +
			"notification_enabled, " +
			"timestamp " +
			") values (?,?,?,?,?,?)")
	if err != nil {
		log.Println("unable to insert user (prepare): ", err)
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(
		user.Email,
		user.PasswordHash,
		user.Active,
		user.UpdateToken,
		user.NotificationEnabled,
		user.Timestamp,
	)
	if err != nil {
		log.Println("unable to insert user (exec): ", err)
		return 0, err
	}
	return res.LastInsertId()
}

func (m *MySql) UpdateUser(user *model.User) error {
	stmt, err := m.db.Prepare(
		"UPDATE user SET " +
			"email = ?, " +
			"password_hash = ?, " +
			"active = ?, " +
			"update_token = ?, " +
			"notification_enabled = ?, " +
			"timestamp = ?, " +
			"subscription_id = ? " +
			"WHERE id = ?")
	if err != nil {
		log.Println("sql error: ", err)
		return err
	}
	now := time.Now()
	defer stmt.Close()
	_, err = stmt.Exec(
		user.Email,
		user.PasswordHash,
		user.Active,
		user.UpdateToken,
		user.NotificationEnabled,
		&now,
		user.SubscriptionId,
		user.Id,
	)
	if err != nil {
		log.Println("sql error: ", err)
		return err
	}
	return nil
}

func (m *MySql) GetUsersByField(field string, value string) ([]*model.User, error) {
	users := make([]*model.User, 0)
	rows, err := m.db.Query(
		"SELECT "+
			"id, "+
			"email, "+
			"password_hash, "+
			"active, "+
			"update_token, "+
			"notification_enabled, "+
			"timestamp, "+
			"subscription_id "+
			"FROM user "+
			"WHERE "+field+" like ?", value)
	if err != nil {
		log.Printf("cannot select users by field: %s, value: %s, error: %v\n", field, value, err)
		return nil, fmt.Errorf("DB error")
	}
	defer rows.Close()

	for rows.Next() {
		user := &model.User{}
		err := rows.Scan(
			&user.Id, &user.Email, &user.PasswordHash, &user.Active, &user.UpdateToken,
			&user.NotificationEnabled, &user.Timestamp, &user.SubscriptionId,
		)
		if err != nil {
			log.Printf("cannot scan users by field: %s, value: %s, error: %v\n", field, value, err)
			return nil, fmt.Errorf("DB error")
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		log.Printf("cannot processes users by field: %s, value: %s, error: %v\n", field, value, err)
		return nil, fmt.Errorf("DB error")
	}
	return users, nil
}

func (m *MySql) DeleteUser(userId int64) error {

	stmt, err := m.db.Prepare("DELETE FROM user WHERE id = ?")
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

func (m *MySql) GetDomainByToken(token string) (*model.Domain, error) {
	return m.getDomainByField("update_token", token)
}

func (m *MySql) GetDomainByName(name string) (*model.Domain, error) {
	return m.getDomainByField("name", name)
}

func (m *MySql) GetDomainTokenUpdatedBefore(before time.Time) (string, error) {
	row := m.db.QueryRow(`
SELECT update_token
FROM domain
WHERE last_update < ? or last_update is null
order by last_update limit 1`, before)
	var token string
	err := row.Scan(&token)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		} else {
			log.Println("Cannot scan a domain: ", err)
			return "", fmt.Errorf("DB error")
		}
	}
	return token, nil
}
func (m *MySql) getDomainByField(field string, value string) (*model.Domain, error) {
	row := m.db.QueryRow(
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
			"lower(name), "+
			"hosted_zone_id "+
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
		&domain.HostedZoneId,
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

func (m *MySql) DeleteDomain(domainId uint64) error {
	stmt, err := m.db.Prepare("DELETE FROM domain WHERE id = ?")
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

func (m *MySql) DeleteAllDomains(userId int64) error {

	stmt, err := m.db.Prepare("DELETE FROM domain WHERE user_id = ?")
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

func (m *MySql) GetUserDomains(userId int64) ([]*model.Domain, error) {
	domains := make([]*model.Domain, 0)
	rows, err := m.db.Query(
		"SELECT "+
			"id, "+
			"lower(name), "+
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
			"hosted_zone_id "+
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
			&domain.HostedZoneId,
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

func (m *MySql) UpdateDomain(domain *model.Domain) error {
	stmt, err := m.db.Prepare(
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
	defer stmt.Close()
	_, err = stmt.Exec(
		strings.ToLower(domain.Name),
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
	return nil
}

func (m *MySql) InsertDomain(domain *model.Domain) error {
	stmt, err := m.db.Prepare(
		"INSERT into domain (" +
			"name, " +
			"update_token, " +
			"user_id, " +
			"device_mac_address, " +
			"device_name, " +
			"device_title, " +
			"last_update," +
			"hosted_zone_id" +
			") values (?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Println("unable to insert domain (prepare): ", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		strings.ToLower(domain.Name),
		domain.UpdateToken,
		domain.UserId,
		domain.DeviceMacAddress,
		domain.DeviceName,
		domain.DeviceTitle,
		domain.LastUpdate,
		domain.HostedZoneId,
	)
	if err != nil {
		log.Println("unable to insert domain (exec): ", err)
		return err
	}
	return nil
}

func (m *MySql) GetAction(userId int64, actionTypeId uint64) (*model.Action, error) {
	row := m.db.QueryRow(
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

func (m *MySql) GetActionByToken(token string, actionTypeId uint64) (*model.Action, error) {
	row := m.db.QueryRow(
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

func (m *MySql) InsertAction(action *model.Action) error {
	stmt, err := m.db.Prepare(
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
	defer stmt.Close()
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
	return nil

}

func (m *MySql) UpdateAction(action *model.Action) error {
	stmt, err := m.db.Prepare(
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
	defer stmt.Close()
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
	return nil

}

func (m *MySql) DeleteActions(userId int64) error {

	stmt, err := m.db.Prepare("DELETE FROM action WHERE user_id = ?")
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

func (m *MySql) DeleteAction(actionId uint64) error {

	stmt, err := m.db.Prepare("DELETE FROM action WHERE id = ?")
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

func (m *MySql) GetCount(query string) (int64, error) {
	row := m.db.QueryRow(query)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		} else {
			return 0, err
		}
	}
	return count, nil
}

func (m *MySql) GetOnlineDevicesCount() (int64, error) {
	return m.GetCount(`
select count(*)  
from domain join user on domain.user_id = user.id 
where timestampdiff(minute, last_update, now()) < 600
`)
}

func (m *MySql) GetDomainCount() (int64, error) {
	return m.GetCount(`select count(*) from domain`)
}

func (m *MySql) GetAllUsersCount() (int64, error) {
	return m.GetCount("select count(*) from user")
}

func (m *MySql) GetActiveUsersCount() (int64, error) {
	return m.GetCount("select count(*) from user where active = true")
}

func (m *MySql) GetSubscribedUsersCount() (int64, error) {
	return m.GetCount("select count(*) from user where subscription_id is not null")
}

func (m *MySql) Get2MonthOldActiveUsersWithoutDomainCount() (int64, error) {
	return m.GetCount(`
select count(*)
from user u
left outer join domain d on u.id = d.user_id
where d.id is null
and u.active = true
and timestampdiff(day, u.timestamp, now()) > 60
`)
}
