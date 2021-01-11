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
		"bundleid text," +
		"notificationtype text," +
		"aps text," +
		"lastnotificationdate date);"

	if err := connect(); err != nil {
		return err
	}
	var stmt *sql.Stmt
	var err error
	fmt.Fprintf(os.Stdout, "[SQL QUERY]:%s\n", schema)
	stmt, err = dbx.Prepare(schema)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	return err
}

func insertQuotes(v string) string {
	return "'" + v + "'"
}

func InsertNewDevice(o *DeviceTable) error {
	if err := connect(); err != nil {
		return err
	}
	defer dbx.Close()
	var err error
	var tx *sql.Tx

	tx, err = dbx.Begin()

	query := "insert into devices(devicetoken, bundleid, notificationtype, aps) values(" +
		insertQuotes(o.DeviceToken) + "," +
		insertQuotes(o.BundleId) + "," +
		insertQuotes(o.Type) + "," +
		insertQuotes(o.Aps) +
		");"
	fmt.Fprintf(os.Stdout, "[SQL QUERY]:%s\n", query)

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
	defer dbx.Close()
	var err error
	var tx *sql.Tx

	tx, err = dbx.Begin()

	query := "update  devices set lastnotificationdate = '" +
		time.Now().Add(time.Hour*2).Format(time.RFC3339) + "'" +
		" where uid = '" + o.Uid + "';"
	fmt.Fprintf(os.Stdout, "[SQL QUERY]:%s\n", query)

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
	defer dbx.Close()
	query := "select uid, devicetoken, bundleid, notificationtype, aps, lastnotificationdate from devices;"

	var res *sql.Rows
	fmt.Fprintf(os.Stdout, "[SQL QUERY]:%s\n", query)
	res, err = dbx.Query(query)
	if err != nil {
		return devices, err
	}

	for res.Next() {
		d := scanRow(res)
		if d != nil {
			devices = append(devices, d)
		}
	}
	res.Close()

	return devices, nil
}

func GetDeviceByToken(token string) (device *DeviceTable, err error) {
	if err := connect(); err != nil {
		return nil, err
	}
	defer dbx.Close()
	query := "select uid, devicetoken, bundleid, notificationtype, aps, lastnotificationdate from devices where devicetoken = " +
		insertQuotes(token) +
		";"

	var res *sql.Rows
	fmt.Fprintf(os.Stdout, "[SQL QUERY]:%s\n", query)
	res, err = dbx.Query(query)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		d := scanRow(res)
		if d != nil {
			return d, nil
		}
	}
	return nil, fmt.Errorf("empty result")
}

func GetDeviceByUid(uid string) (device *DeviceTable, err error) {
	if err := connect(); err != nil {
		return nil, err
	}
	defer dbx.Close()
	query := "uid, devicetoken, bundleid, notificationtype, aps, lastnotificationdate where uid = " +
		insertQuotes(uid) +
		";"

	var res *sql.Rows
	fmt.Fprintf(os.Stdout, "[SQL QUERY]:%s\n", query)
	res, err = dbx.Query(query)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		d := scanRow(res)
		if d != nil {
			return d, nil
		}
	}
	return nil, fmt.Errorf("empty result")
}

func scanRow(res *sql.Rows) *DeviceTable {
	var uid, token, bundleid, notificationType, aps sql.NullString
	var created sql.NullTime
	err := res.Scan(&uid, &token, &bundleid, &notificationType, &aps, &created)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[SQL ERROR]:error while getting row from sql (%v)\n", err)
		return nil
	}
	return DeviceTableRaw(uid.String, token.String, bundleid.String, notificationType.String, aps.String, created.Time)
}
