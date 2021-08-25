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
	"encoding/hex"
	log "github.com/sirupsen/logrus"
	"github.com/emicklei/go-restful"
	db "github.com/openlab-aux/lsmsd/database"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
	"net/http"
)

type ImageWebService struct {
	d *db.ImageDBProvider
	S *restful.WebService
}

func NewImageService(d *db.ImageDBProvider) *ImageWebService {
	res := new(ImageWebService)
	res.d = d

	s := new(restful.WebService)

	s.Path("/images").
		Doc("Picture related services").
		ApiVersion("0.1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	s.Route(s.GET("/{id}").
		Param(restful.PathParameter("id", "Unique image identifier")).
		Doc("Returns a image").
		To(res.GetImageById).
		//Writes(db.Image{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	s.Route(s.GET("/{id}/meta").
		Param(restful.PathParameter("id", "Unique image identifier")).
		Doc("Returns metadata of an image").
		To(res.GetImageMetadataById).
		Writes(db.ImageMetadata{}).
		Do(returnsInternalServerError, returnsNotFound, returnsBadRequest))

	res.S = s

	return res
}

func (p *ImageWebService) GetImageById(req *restful.Request, res *restful.Response) {
	log.Debug(p)
	hid := req.PathParameter("id")
	id, err := hex.DecodeString(hid)
	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusBadRequest, "")
		return
	}
	buf, ct, err := p.d.GetImageById(bson.ObjectId(id))
	if err != nil {
		log.Debug(err)
		if err == mgo.ErrNotFound {
			res.WriteErrorString(http.StatusNotFound, err.Error())
			return
		}
		res.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
		return
	}
	res.AddHeader("Content-Type", ct)
	io.Copy(res, buf)
}

func (p *ImageWebService) GetImageMetadataById(req *restful.Request, res *restful.Response) {
	hid := req.PathParameter("id")
	id, err := hex.DecodeString(hid)
	if err != nil {
		log.Debug(err)
		res.WriteErrorString(http.StatusBadRequest, "")
		return
	}
	meta, err := p.d.GetImageMetadataById(bson.ObjectId(id))
	if err != nil {
		log.Debug(err)
		if err == mgo.ErrNotFound {
			res.WriteErrorString(http.StatusNotFound, err.Error())
			return
		}
		res.WriteErrorString(http.StatusInternalServerError, ERROR_INTERNAL)
	}
	res.WriteAsJson(*meta)
}
