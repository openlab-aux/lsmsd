/*
 *    Copyright (C) 2015 Stefan Luecke
 *
 *    This program is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU Affero General Public License as published
 *    by the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    This program is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU Affero General Public License for more details.
 *
 *    You should have received a copy of the GNU Affero General Public License
 *    along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 *    Authors: Stefan Luecke <glaxx@glaxx.net>
 */

package backend

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
)

const (
	DBVERSION           = 1
	ERROR_INVALID_ID    = "Error: Invalid ID"
	ERROR_STMT_PREPARE  = "Error: Statement prepare failed"
	ERROR_INVALID_INPUT = "ERROR: Invalid Input"
	ERROR_INTERNAL      = "Error: Internal Server Error"
	ERROR_INSERT        = "Error: DB Insert failed"
	ERROR_QUERY         = "Error: DB Query failed"
)

var db *sql.DB

func RegisterDatabase(d *sql.DB) {
	db = d
	chk := checkDatabase(db)
	log.WithFields(log.Fields{"Tables": chk}).Debug("chkDB")
	if chk == 0 {
		log.Warning("Database empty - Ignore if this is the first start")
		err := createTableMeta(db)
		if err != nil {
			log.Fatal(err)
		}
		err = createMisc(db)
		if err != nil {
			log.Fatal(err)
		}
		return
	} else if chk != 4 {
		log.Fatal("Database corrupt")
	}
	log.WithFields(log.Fields{"Version": getDatabaseVersion(db)}).Info("Found Database")
}

func getDatabaseVersion(db *sql.DB) int {
	stmt, err := db.Prepare("SELECT dbversion FROM metainf;")
	if err != nil {
		log.Fatal(err)
	}
	result := 0
	err = stmt.QueryRow().Scan(&result)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

// returns # of tables & dbver
func checkDatabase(db *sql.DB) int {
	names := [...]string{"item", "metainf", "user", "policies"}
	stmt, err := db.Prepare("SELECT name FROM sqlite_master WHERE type='table' AND name=?;")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()
	cnt := 0
	for _, n := range names {
		rows, err := stmt.Query(n)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			cnt++
		}
		log.WithFields(log.Fields{"Error": err, "cnt": cnt, "n": n}).Debug("ckDBQuery")
	}
	return cnt
}

func createTableMeta(db *sql.DB) error {
	createMeta := fmt.Sprintf(`
        CREATE TABLE 'metainf' (
            'dbversion' INTEGER
        );
        INSERT INTO metainf(dbversion)
        VALUES("%v");
        `, DBVERSION)
	_, err := db.Exec(createMeta)
	if err != nil {
		return err
	}
	return nil
}

func createMisc(db *sql.DB) error {
	createMisc := `
        CREATE TABLE 'item' (
            'id' INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            'name' TEXT,
            'description' TEXT,
			'contains' TEXT,
			'owner'	INTEGER,
			'maintainer' INTEGER,
			'usage' INTEGER,
			'discard' INTEGER
        );
        CREATE TABLE 'user' (
            'id' INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
            'name' TEXT
        );
		CREATE TABLE 'policies' (
			'id' INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			'name' TEXT,
			'description' TEXT
		);
        `
	_, err := db.Exec(createMisc)
	if err != nil {
		return err
	}
	return err
}
