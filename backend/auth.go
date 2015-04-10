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
	//"crypto/sha512"
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	//	"gopkg.in/mgo.v2/bson"
	db "github.com/openlab-aux/lsmsd/database"
	"net/http"
)

type BasicAuthService struct {
	d *db.UserDBProvider
}

func NewBasicAuthService(d *db.UserDBProvider) *BasicAuthService {
	res := new(BasicAuthService)
	res.d = d
	return res
}

func (s *BasicAuthService) Auth(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	u, p, ok := request.Request.BasicAuth()
	if !ok {
	}
	usr, err := s.d.GetUserByName(u)
	if err != nil || !ok {
		log.WithFields(log.Fields{"User": u}).Warn("Failed login attempt")
		response.AddHeader("WWW-Authenticate", "Basic realm=\""+request.SelectedRoutePath()+"\"")
		response.WriteErrorString(http.StatusUnauthorized, "Username / Password incorrect")
		return
	}

	pwcorrect := usr.Secret.VerifyPassword(p)
	if pwcorrect {
		log.Debug("User Authentication successful")
	} else {
		log.WithFields(log.Fields{"User": u}).Warn("Failed login attempt")
		response.AddHeader("WWW-Authenticate", "Basic realm=\""+request.SelectedRoutePath()+"\"")
		response.WriteErrorString(http.StatusUnauthorized, "Username / Password incorrect")
		return

	}
	request.SetAttribute("User", usr.Name)
	chain.ProcessFilter(request, response)
}
