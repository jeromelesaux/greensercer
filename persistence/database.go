package persistence

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jeromelesaux/greenserver/persistence/amazon"
	"github.com/jmoiron/sqlx"
)

var dbx *sqlx.DB
var (
	_dbEndpoint string
	_awsRegion  string
	_dbUser     string
	_dbName     string
	_dbPassword string
)

func Initialise(dbEndpoint, awsRegion, dbUser, dbName, dbpassword string) error {
	_dbEndpoint = dbEndpoint
	_awsRegion = awsRegion
	_dbUser = dbUser
	_dbName = dbName
	_dbPassword = dbpassword
	if err := connect(); err != nil {
		return err
	}
	if err := createSchema(); err != nil {
		return err
	}
	return nil
}

func connect() error {
	dbx = amazon.ConnectRds(_dbEndpoint, _awsRegion, _dbUser, _dbName, _dbPassword)
	return nil
}

func createSchema() error {
	schema := "create table if not exists devices (" +
		"uid serial primary key ," +
		"devicetoken text, " +
		"lastnotificationdate date);"

	if err := connect(); err != nil {
		return err
	}
	var stmt *sql.Stmt
	var err error
	stmt, err = dbx.Prepare(schema)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	return err
}

func InsertNewDevice(o *DeviceTable) error {
	if err := connect(); err != nil {
		return err
	}
	var err error
	var tx *sql.Tx

	tx, err = dbx.Begin()

	query := "insert into devices(devicetoken) values('" +
		o.DeviceToken +
		"');"
	fmt.Fprintf(os.Stdout, "%s\n", query)

	insert, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	_, err = insert.Exec()
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func UpdateDevice(o *DeviceTable) error {
	if err := connect(); err != nil {
		return err
	}
	var err error
	var tx *sql.Tx

	tx, err = dbx.Begin()

	query := "update  devices set lastnotificationdate = '" +
		time.Now().Add(time.Hour*2).Format(time.RFC3339) + "'" +
		" where uid = '" + o.Uid + "';"
	fmt.Fprintf(os.Stdout, "%s\n", query)

	insert, err := tx.Prepare(query)

	if err != nil {
		return err
	}

	_, err = insert.Exec()
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func GetAllDevices() (devices []*DeviceTable, err error) {
	devices = make([]*DeviceTable, 0)
	if err := connect(); err != nil {
		return devices, err
	}
	query := "select uid, devicetoken,lastnotificationdate from devices;"

	var res *sql.Rows

	res, err = dbx.Query(query)
	if err != nil {
		return devices, err
	}

	for res.Next() {
		var uid, token string
		var created time.Time
		err = res.Scan(&uid, &token, &created)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while getting row from sql (%v)\n", err)
		} else {
			devices = append(devices, DeviceTableRaw(uid, token, created))
		}

	}
	res.Close()

	return devices, nil
}

func GetDeviceByToken(token string) (device *DeviceTable, err error) {
	if err := connect(); err != nil {
		return nil, err
	}
	query := "select uid, devicetoken,lastnotificationdate from devices where devicetoken = " +
		"'" + token + "'" +
		";"

	var res *sql.Rows

	res, err = dbx.Query(query)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var uid, token string
		var created time.Time
		err = res.Scan(&uid, &token, &created)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while getting row from sql (%v)\n", err)
		} else {
			return DeviceTableRaw(uid, token, created), nil
		}

	}
	return nil, fmt.Errorf("empty result")
}

func GetDeviceByUid(uid string) (device *DeviceTable, err error) {
	if err := connect(); err != nil {
		return nil, err
	}
	query := "select uid, devicetoken,lastnotificationdate from devices where uid = " +
		"'" + uid + "'" +
		";"

	var res *sql.Rows

	res, err = dbx.Query(query)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var uid, token string
		var created time.Time
		err = res.Scan(&uid, &token, &created)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error while getting row from sql (%v)\n", err)
		} else {
			return DeviceTableRaw(uid, token, created), nil
		}

	}
	return nil, fmt.Errorf("empty result")
}
