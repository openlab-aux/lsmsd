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

package webservice

import (
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	db "github.com/openlab-aux/lsmsd/database"
	"github.com/trevex/golem"
	"net/http"
)

type UpdateService struct {
	Router *golem.Router
	u      *golem.Room
	S      *restful.WebService
}

func NewUpdateService() *UpdateService {
	res := new(UpdateService)
	res.Router = golem.NewRouter()
	err := res.Router.OnConnect(res.join)
	if err != nil {
		panic(err)
	}
	err = res.Router.OnClose(res.leave)
	if err != nil {
		panic(err)
	}
	res.u = golem.NewRoom()

	service := new(restful.WebService)
	service.
		Path("/ws").
		Doc("Update channel").
		ApiVersion("0.1")

	service.Route(service.GET("").
		Doc("Just use a GET request to invoke a websocket connection").
		To(res.restfulHandlerWrapper))

	res.S = service
	return res
}

func (u *UpdateService) restfulHandlerWrapper(req *restful.Request, res *restful.Response) {
	h := u.Router.Handler()
	log.Debug("wrap", h, res.ResponseWriter, req)

	h(res.ResponseWriter, req.Request)
}

func (u *UpdateService) leave(conn *golem.Connection) {
	log.Debug("Lost ws connection")
	u.u.Leave(conn)
}

func (u *UpdateService) join(conn *golem.Connection, req *http.Request) {
	log.Debug("Got ws connect")
	u.u.Join(conn)
}

func (u *UpdateService) PushUpdate(obj interface{}) {
	var _type string
	switch obj.(type) {
	case *db.ItemHistory:
		_type = "ItemHistory"

	case *db.PolicyHistory:
		_type = "PolicyHistory"

	default:
		panic("Invalid type; ws")

	}
	log.Debug("ws: pushed obj")
	u.u.Emit(_type, obj)
}
