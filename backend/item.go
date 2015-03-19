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
	//	"database/sql"
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type Item struct {
	ID          uint64
	Name        string
	Description string
	Contains    []uint64
	Owner       uint64
	Maintainer  uint64
	Usage       uint64
	Discard     uint64
}

func NewItemService() *restful.WebService {
	service := new(restful.WebService)
	service.
		Path("/item").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	service.Route(service.GET("/{id}").To(GetItemById))
	service.Route(service.GET("").To(ListItem))
	service.Route(service.PUT("").To(UpdateItem))
	service.Route(service.PUT("/{id}").To(UpdateItem))
	service.Route(service.POST("").To(CreateItem))
	service.Route(service.DELETE("/{id}").To(DeleteItem))
	return service
}

func GetItemById(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	log.WithFields(log.Fields{"Requested ID": sid}).Debug("Got Request")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	itm, err := getItemById(id)
	if err != nil || itm.ID == 0 {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		log.WithFields(log.Fields{"Error Msg": err}).
			Info(ERROR_INVALID_ID)
		return
	}
	response.WriteEntity(itm)
}

func getItemById(id uint64) (Item, error) {
	stmt, err := db.Prepare("SELECT * FROM item WHERE id=?;")
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).
			Fatal(ERROR_STMT_PREPARE)
	}
	defer stmt.Close()
	row := stmt.QueryRow(id)
	itm := new(Item)
	unpacked_string := ""
	err = row.Scan(&itm.ID, &itm.Name, &itm.Description, &unpacked_string, &itm.Owner, &itm.Maintainer, &itm.Usage, &itm.Discard)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).
			Warn("Scan Error")
	}
	itm.Contains = unpack_contains(unpacked_string)
	return *itm, nil
}

func unpack_contains(input string) []uint64 {
	ids := strings.Split(input, ";")

	if len(ids) <= 1 {
		return make([]uint64, 0)
	}
	res := make([]uint64, len(ids))
	for i := 0; i != len(ids); i++ {
		log.WithFields(log.Fields{"String": ids[i], "length": len(ids), "i": i}).Debug("Trying to parse string")
		temp, err := strconv.ParseUint(ids[i], 10, 64)
		if err != nil {
			log.WithFields(log.Fields{"Error Msg": err}).
				Warn("Failed to unpack 'Contains' string")
		}
		res[i] = temp
	}
	return res
}

func pack_contains(input []uint64) string {
	if len(input) == 0 {
		return ""
	}
	if len(input) == 1 {
		return strconv.FormatUint(input[0], 10)
	}
	str := ""
	for i := 0; i != len(input); i++ {
		str = str + strconv.FormatUint(input[i], 10) + ";"
	}
	return str
}

func CreateItem(request *restful.Request, response *restful.Response) {
	itm := new(Item)
	err := request.ReadEntity(itm)
	if err != nil {
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_INPUT)
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_INPUT)
		return
	}

	id, err := createItem(itm)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity("/item/" + strconv.FormatUint(id, 10))
}

func createItem(itm *Item) (uint64, error) {
	stmt, err := db.Prepare(`
		INSERT INTO item ('name', 'description', 'contains', 'owner', 'maintainer', 'usage', 'discard') VALUES (?, ?, ?, ?, ?, ?, ?);`)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Fatal(ERROR_STMT_PREPARE)
		return 0, err
	}
	defer stmt.Close()
	res, err := stmt.Exec(template.HTMLEscapeString(itm.Name), template.HTMLEscapeString(itm.Description), pack_contains(itm.Contains), itm.Owner, itm.Maintainer, itm.Usage, itm.Discard)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INSERT)
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_QUERY)
		return 0, err
	}

	return uint64(id), nil
}

func ListItem(request *restful.Request, response *restful.Response) {
	itm, err := listItem()
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	response.WriteEntity(itm)
}

func listItem() ([]Item, error) {
	stmt, err := db.Prepare(`
		SELECT * FROM item;`)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Fatal(ERROR_STMT_PREPARE)
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query()
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return nil, err
	}
	result := make([]Item, 0)
	for rows.Next() {
		temp := new(Item) // Item{0, "", make([]uint64, 0), 0, 0, 0, 0)
		str := ""
		rows.Scan(&temp.ID, &temp.Name, &temp.Description, &str, &temp.Owner, &temp.Maintainer, &temp.Usage, &temp.Discard)
		temp.Contains = unpack_contains(str)
		result = append(result, *temp)
	}

	return result, nil
}

func UpdateItem(request *restful.Request, response *restful.Response) {
	itm := new(Item)
	err := request.ReadEntity(itm)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	stmt, err := db.Prepare(`
		SELECT * FROM item ORDER BY id DESC LIMIT 0, 1;`)
	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}
	defer stmt.Close()
	row := stmt.QueryRow()
	var id uint64
	temp := new(Item)
	temp2 := ""
	err = row.Scan(&id, &temp.Name, &temp.Description, &temp2, &temp.Owner, &temp.Maintainer, &temp.Usage, &temp.Discard)

	if err != nil {
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return
	}

	log.WithFields(log.Fields{"Id": id}).Debug("Highest ID")

	if itm.ID <= id {
		err := updateItem(itm)
		if err != nil {
			response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
			return
		} else {
			response.WriteEntity(true)
		}
	} else {
		CreateItem(request, response)
	}
}

func updateItem(itm *Item) error {
	stmt, err := db.Prepare(`
		UPDATE item SET name=?, description=?, contains=?, owner=?, maintainer=?, usage=?, discard=? WHERE id=?;`)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(template.HTMLEscapeString(itm.Name), template.HTMLEscapeString(itm.Description), pack_contains(itm.Contains), itm.Owner, itm.Maintainer, itm.Usage, itm.Discard, itm.ID)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Warn(ERROR_INTERNAL)
		return err
	}
	return nil
}

func DeleteItem(request *restful.Request, response *restful.Response) {
	sid := request.PathParameter("id")
	id, err := strconv.ParseUint(sid, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INVALID_ID)
		response.WriteErrorString(http.StatusBadRequest, ERROR_INVALID_ID)
		return
	}

	err = deleteItem(id)
	if err != nil {
		log.WithFields(log.Fields{"Error Msg": err}).Info(ERROR_INTERNAL)
		response.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	response.WriteEntity(true)
}

func deleteItem(id uint64) error {
	stmt, err := db.Prepare(`
		DELETE FROM item WHERE id=?;
		`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}
