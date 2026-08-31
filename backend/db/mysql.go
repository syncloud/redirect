package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/product"
	"go.uber.org/zap"
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
	logger   *zap.Logger
}

func NewMySql(host string, database string, user string, password string, logger *zap.Logger) *MySql {
	return &MySql{
		host:     host,
		database: database,
		user:     user,
		password: password,
		logger:   logger,
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

func (m *MySql) GetNextUserId(id int64) (int64, error) {
	row := m.db.QueryRow(
		"SELECT id FROM user WHERE id > ? order by id asc limit 1", id)

	var nextId int64
	err := row.Scan(&nextId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		} else {
			m.logger.Error("cannot find next user", zap.Error(err))
			return 0, err
		}
	}

	return nextId, nil
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
			"subscription_id, "+
			"subscription_type, "+
			"plan, "+
			"registered_at, "+
			"status_at, "+
			"status "+
			"FROM user "+
			"WHERE "+field+" = ?", value)

	user := &model.User{}
	err := row.Scan(&user.Id, &user.Email, &user.PasswordHash, &user.Active, &user.UpdateToken,
		&user.NotificationEnabled, &user.Timestamp, &user.SubscriptionId, &user.SubscriptionType, &user.Plan, &user.RegisteredAt,
		&user.StatusAt, &user.Status)

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

func gclidAt(user *model.User) *time.Time {
	if user.Gclid == nil {
		return nil
	}
	return &user.Timestamp
}

func (m *MySql) InsertUser(user *model.User) (int64, error) {
	stmt, err := m.db.Prepare(
		"INSERT into user (" +
			"email, " +
			"password_hash, " +
			"active, " +
			"update_token, " +
			"notification_enabled, " +
			"timestamp, " +
			"gclid, " +
			"gclid_at " +
			") values (?,?,?,?,?,?,?,?)")
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
		user.Gclid,
		gclidAt(user),
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
			"subscription_id = ?, " +
			"subscription_type = ?, " +
			"plan = ?, " +
			"status = ?, " +
			"status_at = ? " +
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
		user.SubscriptionType,
		user.Plan,
		user.Status,
		user.StatusAt,
		user.Id,
	)
	if err != nil {
		log.Println("sql error: ", err)
		return err
	}
	return nil
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
			log.Println("Cannot scan a update_token: ", err)
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
			"relay, "+
			"mail_relay, "+
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
		&domain.Relay,
		&domain.MailRelay,
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
			"relay = ?, " +
			"mail_relay = ?, " +
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
		domain.Relay,
		domain.MailRelay,
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
			"timestamp, " +
			"sent_at, " +
			"attempts" +
			") values (?,?,?,?,?,?)")
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
		action.SentAt,
		action.Attempts,
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
			"timestamp = ?, " +
			"sent_at = NULL, " +
			"attempts = 0 " +
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

func (m *MySql) GetPendingActivations(actionTypeId uint64, maxAttempts int, limit int) ([]*model.PendingActivation, error) {
	rows, err := m.db.Query(
		"SELECT a.id, a.token, u.email, a.attempts "+
			"FROM action a JOIN user u ON u.id = a.user_id "+
			"WHERE a.action_type_id = ? AND a.sent_at IS NULL AND a.attempts < ? "+
			"ORDER BY a.id LIMIT ?", actionTypeId, maxAttempts, limit)
	if err != nil {
		log.Println("unable to query pending activations: ", err)
		return nil, err
	}
	defer rows.Close()
	var pending []*model.PendingActivation
	for rows.Next() {
		p := &model.PendingActivation{}
		if err := rows.Scan(&p.ActionId, &p.Token, &p.Email, &p.Attempts); err != nil {
			log.Println("unable to scan pending activation: ", err)
			return nil, err
		}
		pending = append(pending, p)
	}
	return pending, rows.Err()
}

func (m *MySql) MarkActionSent(actionId uint64, now time.Time) error {
	_, err := m.db.Exec("UPDATE action SET sent_at = ?, attempts = attempts + 1 WHERE id = ?", now, actionId)
	if err != nil {
		log.Println("unable to mark action sent: ", err)
	}
	return err
}

func (m *MySql) IncrementActionAttempts(actionId uint64) error {
	_, err := m.db.Exec("UPDATE action SET attempts = attempts + 1 WHERE id = ?", actionId)
	if err != nil {
		log.Println("unable to increment action attempts: ", err)
	}
	return err
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
and (ip is not null or ipv6 is not null) 
`)
}

func (m *MySql) GetOnlineUsersCount() (int64, error) {
	return m.GetCount(`
select count(*) 
from user 
where exists (
	select user_id 
	from domain 
	where user_id = user.id 
	  and timestampdiff(minute, last_update, now()) < 600 
	  and (ip is not null or ipv6 is not null) 
)
`)
}

func (m *MySql) GetDomainCount() (int64, error) {
	return m.GetCount(`select count(*) from domain`)
}

func (m *MySql) AddRelayTraffic(name string, yearMonth string, bytes int64) error {
	stmt, err := m.db.Prepare(
		"INSERT INTO relay_traffic (name, `year_month`, bytes) VALUES (?, ?, ?) " +
			"ON DUPLICATE KEY UPDATE bytes = bytes + VALUES(bytes)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, yearMonth, bytes)
	return err
}

func (m *MySql) AddMailRelayMessages(name string, yearMonth string, messages int64) error {
	stmt, err := m.db.Prepare(
		"INSERT INTO mail_relay_usage (name, `year_month`, messages) VALUES (?, ?, ?) " +
			"ON DUPLICATE KEY UPDATE messages = messages + VALUES(messages)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, yearMonth, messages)
	return err
}

func (m *MySql) GetMailRelayMessages(name string, yearMonth string) (int64, error) {
	row := m.db.QueryRow(
		"SELECT COALESCE(messages, 0) FROM mail_relay_usage WHERE name = ? AND `year_month` = ?",
		name, yearMonth)
	var messages int64
	if err := row.Scan(&messages); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return messages, nil
}

func (m *MySql) AddMailRelayBounces(name string, yearMonth string, bounces int64) error {
	stmt, err := m.db.Prepare(
		"INSERT INTO mail_relay_usage (name, `year_month`, bounces) VALUES (?, ?, ?) " +
			"ON DUPLICATE KEY UPDATE bounces = bounces + VALUES(bounces)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, yearMonth, bounces)
	return err
}

func (m *MySql) GetMailRelayBounces(name string, yearMonth string) (int64, error) {
	row := m.db.QueryRow(
		"SELECT COALESCE(bounces, 0) FROM mail_relay_usage WHERE name = ? AND `year_month` = ?",
		name, yearMonth)
	var bounces int64
	if err := row.Scan(&bounces); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return bounces, nil
}

func (m *MySql) IsMailRelayEnabledForUser(userId int64) (bool, error) {
	row := m.db.QueryRow("SELECT COUNT(*) FROM domain WHERE user_id = ? AND mail_relay = 1", userId)
	var count int64
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *MySql) IsRelayEnabledForUser(userId int64) (bool, error) {
	row := m.db.QueryRow("SELECT COUNT(*) FROM domain WHERE user_id = ? AND relay = 1", userId)
	var count int64
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *MySql) GetMailRelayUsageForUser(userId int64, yearMonth string) (int64, error) {
	row := m.db.QueryRow(
		"SELECT COALESCE(SUM(u.messages), 0) FROM mail_relay_usage u "+
			"JOIN domain d ON u.name = d.name "+
			"WHERE d.user_id = ? AND u.`year_month` = ?", userId, yearMonth)
	var messages int64
	if err := row.Scan(&messages); err != nil {
		return 0, err
	}
	return messages, nil
}

func (m *MySql) GetMailRelayUsageAll(yearMonth string) ([]model.MailRelayDomainUsage, error) {
	rows, err := m.db.Query(
		"SELECT name, COALESCE(messages, 0), COALESCE(bounces, 0) FROM mail_relay_usage WHERE `year_month` = ?",
		yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var usage []model.MailRelayDomainUsage
	for rows.Next() {
		var entry model.MailRelayDomainUsage
		if err := rows.Scan(&entry.Name, &entry.Messages, &entry.Bounces); err != nil {
			return nil, err
		}
		usage = append(usage, entry)
	}
	return usage, rows.Err()
}

func (m *MySql) BlockMailRelay(name string, reason string) error {
	stmt, err := m.db.Prepare(
		"INSERT INTO mail_relay_blocked (name, reason) VALUES (?, ?) " +
			"ON DUPLICATE KEY UPDATE reason = VALUES(reason)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(name, reason)
	return err
}

func (m *MySql) IsMailRelayBlocked(name string) (bool, error) {
	row := m.db.QueryRow("SELECT COUNT(*) FROM mail_relay_blocked WHERE name = ?", name)
	var count int64
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (m *MySql) GetUserDomainNames(userId int64) ([]string, error) {
	rows, err := m.db.Query("SELECT lower(name) FROM domain WHERE user_id = ?", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func (m *MySql) GetRelayTraffic(names []string, yearMonth string) (int64, error) {
	if len(names) == 0 {
		return 0, nil
	}
	arguments := make([]interface{}, 0, len(names)+1)
	arguments = append(arguments, yearMonth)
	placeholders := make([]string, len(names))
	for i, name := range names {
		placeholders[i] = "?"
		arguments = append(arguments, name)
	}
	query := "SELECT COALESCE(SUM(bytes), 0) FROM relay_traffic " +
		"WHERE `year_month` = ? AND name IN (" + strings.Join(placeholders, ",") + ")"
	var bytes int64
	if err := m.db.QueryRow(query, arguments...).Scan(&bytes); err != nil {
		return 0, err
	}
	return bytes, nil
}

func (m *MySql) GetRelayTrafficMonth(yearMonth string) (map[string]int64, error) {
	rows, err := m.db.Query("SELECT name, bytes FROM relay_traffic WHERE `year_month` = ?", yearMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]int64{}
	for rows.Next() {
		var name string
		var bytes int64
		if err := rows.Scan(&name, &bytes); err != nil {
			return nil, err
		}
		result[name] = bytes
	}
	return result, rows.Err()
}

func (m *MySql) GetActiveDevicesByPlatformVersion(window time.Duration) (map[string]int64, error) {
	rows, err := m.db.Query(`
select coalesce(platform_version, ''), count(*)
from domain
where timestampdiff(minute, last_update, now()) < ?
group by platform_version
`, int(window.Minutes()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int64{}
	for rows.Next() {
		var v string
		var n int64
		if err := rows.Scan(&v, &n); err != nil {
			return nil, err
		}
		out[v] = n
	}
	return out, rows.Err()
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

func (m *MySql) InsertOrder(order *product.Order) (int64, error) {
	res, err := m.db.Exec(
		"INSERT into device_order ("+
			"user_id, device, `option`, total, provider, reference, "+
			"name, address, city, postcode, country"+
			") values (?,?,?,?,?,?,?,?,?,?,?)",
		order.UserId, order.Device, order.Option, order.Total, order.Provider, order.Reference,
		order.Name, order.Address, order.City, order.Postcode, order.Country)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (m *MySql) SetOrderProviderReference(id int64, providerReference string) error {
	_, err := m.db.Exec(
		"UPDATE device_order set provider_reference = ? where id = ?", providerReference, id)
	return err
}

func (m *MySql) GetOrderByReference(reference string) (*product.Order, error) {
	order := &product.Order{}
	var providerReference sql.NullString
	var userId sql.NullInt64
	err := m.db.QueryRow(
		"SELECT id, user_id, device, `option`, total, provider, reference, provider_reference, "+
			"name, address, city, postcode, country, paid "+
			"from device_order where reference = ?", reference).
		Scan(&order.Id, &userId, &order.Device, &order.Option, &order.Total,
			&order.Provider, &order.Reference, &providerReference,
			&order.Name, &order.Address, &order.City, &order.Postcode, &order.Country, &order.Paid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	order.ProviderReference = providerReference.String
	order.UserId = userId.Int64
	return order, nil
}

func (m *MySql) MarkOrderPaid(id int64) error {
	_, err := m.db.Exec("UPDATE device_order set paid = 1 where id = ?", id)
	return err
}

func (m *MySql) GetUnpaidOrders(before time.Time) ([]*product.Order, error) {
	rows, err := m.db.Query(
		"SELECT id, user_id, device, `option`, total, provider, reference, provider_reference, "+
			"name, address, city, postcode, country, paid "+
			"from device_order where paid = 0 and provider_reference is not null "+
			"and created_at < ? order by id", before)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	orders := []*product.Order{}
	for rows.Next() {
		order := &product.Order{}
		var providerReference sql.NullString
		var userId sql.NullInt64
		err := rows.Scan(&order.Id, &userId, &order.Device, &order.Option, &order.Total,
			&order.Provider, &order.Reference, &providerReference,
			&order.Name, &order.Address, &order.City, &order.Postcode, &order.Country, &order.Paid)
		if err != nil {
			return nil, err
		}
		order.ProviderReference = providerReference.String
		order.UserId = userId.Int64
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (m *MySql) RedactOrders(userId int64) error {
	_, err := m.db.Exec(
		"UPDATE device_order set name = '', address = '', city = '', postcode = '' where user_id = ?",
		userId)
	return err
}
