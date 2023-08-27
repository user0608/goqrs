package migrate

import (
	"database/sql"
	_ "embed"
	"encoding/base64"
	"goqrs/envs"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

//go:embed database.sql
var databaseScript string

func RunSimpleMigration(conn *gorm.DB) error {
	const query = `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'account')`
	var value sql.NullBool
	if err := conn.Raw(query).Scan(&value).Error; err != nil {
		return err
	}
	if value.Bool {
		return nil
	}
	for _, query := range strings.Split(databaseScript, ";") {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		if err := conn.Exec(query).Error; err != nil {
			return err
		}
	}
	username := envs.FindEnv("ROOT_USER", "admin")
	password := envs.FindEnv("ROOT_USER_PASSWORD", "")
	if password == "" {
		result, err := bcrypt.GenerateFromPassword([]byte("admin"), 12)
		if err != nil {
			return err
		}
		password = string(result)
	} else {
		decoded, err := base64.StdEncoding.DecodeString(password)
		if err != nil {
			return err
		}
		password = string(decoded)
	}
	var inserQuery = `insert into account(username, first_name, last_name, email, password)
	values (?,?,?,?,?);`
	if err := conn.Exec(inserQuery, username, username, username, username, password).Error; err != nil {
		return err
	}
	return nil
}
