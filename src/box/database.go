/*                                                                                                                                                                                                                  
 * Copyright 2011 Dustin Frisch<fooker@lab.sh>
 * 
 * This file is part of box.
 * 
 * box is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * 
 * box is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * 
 * You should have received a copy of the GNU General Public License
 * along with box. If not, see <http://www.gnu.org/licenses/>.
 */

package box

import (
	"database/sql"
	"flag"
	"fmt"
	"time"
)
import (
	_ "github.com/bmizerany/pq"
)

type UploadRecord struct {
	// The ID of the upload
	Id string `json:"id"`

	// The filename of the upload
	Filename string `json:"filename"`

	// The name of the upload
	Title string `json:"title"`

	// The description of the upload
	Description string `json:"description"`

	// The user who has done the upload
	User string `json:"user"`

	// The creation timestamp of the upload
	Creation time.Time `json:"creation"`

	// The expiration timestamp of the upload
	Expiration time.Time `json:"expiration"`

	// The size of the upload
	Size uint64 `json:"size"`
}

// Comand line arguments
var (
	flag_db_host string
	flag_db_port string
	flag_db_name string
	flag_db_user string
	flag_db_pass string
)

func init() {
	flag.StringVar(&flag_db_host, "db_host", "127.0.0.1", "host of the database connection")
	flag.StringVar(&flag_db_port, "db_port", "5432", "database port of the database connection")
	flag.StringVar(&flag_db_name, "db_name", "box", "name of the database connection")
	flag.StringVar(&flag_db_user, "db_user", "box", "username of the database connection")
	flag.StringVar(&flag_db_pass, "db_pass", "box", "password of the database connection")
}

type Database struct {
	host string
	port string
	name string
	user string
	pass string

	connection *sql.DB
}

type Transaction struct {
	connection  *sql.DB
	transaction *sql.Tx
}

// Connects to a database
func ConnectDatabase() (db *Database, err error) {
	var connection *sql.DB

	// Create connection string to PSQL database
	var source string = fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable", flag_db_host, flag_db_port, flag_db_name, flag_db_user, flag_db_pass)

	// Connect to PostgreSQL
	connection, err = sql.Open("postgres", source)
	if err != nil {
		return
	}

	// Create database wrapper
	db = &Database{
		host: flag_db_host,
		port: flag_db_port,
		name: flag_db_name,
		user: flag_db_user,
		pass: flag_db_pass,

		connection: connection,
	}

	return
}

// Returns the host of the database
func (db *Database) Host() string {
	return db.host
}

// Returns the port of the database
func (db *Database) Port() string {
	return db.port
}

// Returns the name of the database
func (db *Database) Name() string {
	return db.name
}

// Returns the user of the database
func (db *Database) User() string {
	return db.user
}

// Begin a new transaction
func (db *Database) BeginTransaction() (tx *Transaction, err error) {
	var transaction *sql.Tx

	transaction, err = db.connection.Begin()
	if err != nil {
		return
	}

	tx = &Transaction{
		connection:  db.connection,
		transaction: transaction,
	}

	return
}

// Execute a given function in a transaction.
// The transaction will be commited if the function returned no error, otherwise
// the transaction is rolled back.
func (tx *Transaction) do(f func(tx *sql.Tx) (err error)) (err error) {
	// Call function and rollback on error
	err = f(tx.transaction)
	if err != nil {
		tx.Rollback()
		return
	}

	return
}

// Insert a upload record into the database
func (tx *Transaction) Insert(record UploadRecord) (err error) {
	err = tx.do(func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(`
      INSERT INTO uploads (
        "id",
        "filename",
        "title",
        "description",
        "user",
        "creation",
        "expiration",
        "size"
      ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			record.Id,
			record.Filename,
			record.Title,
			record.Description,
			record.User,
			record.Creation,
			record.Expiration,
			record.Size)

		return
	})

	return
}

// Query the currently used space for the given user
func (tx *Transaction) QuerySpaceConsumptionFor(user string) (result uint64, err error) {
	err = tx.do(func(tx *sql.Tx) (err error) {
		var row *sql.Row

		row = tx.QueryRow(`
      SELECT
        SUM("u"."size") AS "size"
			FROM (
			  SELECT --DISTINCT
			    "u"."user" AS "user",
			    "u"."id" AS "id",
			    "u"."size" AS "size"
			  FROM "uploads" AS "u"
			  WHERE "u"."expiration" >= NOW()
			    AND "u"."user" = $1
			) AS "u"
			GROUP BY "u"."user"`,
			user)

    // Get result - use zero if query returned no result for the given username
		err = row.Scan(&result)
		if err == sql.ErrNoRows {
			result = 0
			err = nil
		}

		return
	})

	return
}

// Query the currently used space for all users whith currently active uploads
func (tx *Transaction) QuerySpaceConsumption() (result map[string]uint64, err error) {
	err = tx.do(func(tx *sql.Tx) (err error) {
		var rows *sql.Rows

		rows, err = tx.Query(`
      SELECT
        "u"."user" AS "user",
        SUM("u"."size") AS "size"
      FROM (
        SELECT DISTINCT
          "u"."user" AS "user",
          "u"."id" AS "id",
          "u"."size" AS "size"
        FROM "uploads" AS "u"
        WHERE "u"."expiration" >= NOW()
      ) AS "u"
      GROUP BY "u"."user"`)
		if err != nil {
			return
		}

		defer rows.Close()

		result = make(map[string]uint64)

		for rows.Next() {
			var (
				user  string
				space uint64
			)

			err = rows.Scan(&user, &space)
			if err != nil {
				return
			}

			result[user] = space
		}

		return
	})

	return
}

// Query all expired uploads - uploads thus not having any records which are currently active
func (tx *Transaction) QueryExpiredUploads() (result []string, err error) {
	err = tx.do(func(tx *sql.Tx) (err error) {
		var rows *sql.Rows

		rows, err = tx.Query(`
      SELECT 
			  "u"."id" AS "id"
			FROM "uploads" AS "u"
			GROUP BY "u"."id"
			HAVING MAX("u"."expiration") < NOW()`)
		if err != nil {
			return
		}

		defer rows.Close()

		for rows.Next() {
			var id string

			err = rows.Scan(&id)
			if err != nil {
				return
			}

			result = append(result, id)
		}

		return
	})

	return
}

// Close the transaction
func (tx *Transaction) Commit() (err error) {
	err = tx.transaction.Commit()
	if err != nil {
		tx.Rollback()
		return
	}

	return
}

// Close the transaction
func (tx *Transaction) Rollback() (err error) {
	err = tx.transaction.Rollback()

	return
}

// Close the database
func (db *Database) Close() (err error) {
	err = db.connection.Close()

	return
}
