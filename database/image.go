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

package database

import (
	"bytes"
	//	log "github.com/Sirupsen/logrus"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io"
)

type ImageDBProvider struct {
	c *mgo.GridFS
}

func NewImageDBProvider(s *mgo.Session, dbname string) *ImageDBProvider {
	res := new(ImageDBProvider)
	res.c = s.DB(dbname).GridFS("images")
	return res
}

type ImageMetadata struct {
	ItmRef uint64
	User   string
}

func (p *ImageDBProvider) Create(data io.Reader, user, contentType string, obj uint64) (bson.ObjectId, error) {
	f, err := p.c.Create("")
	if err != nil {
		return "", err
	}
	f.SetContentType(contentType)
	meta := new(ImageMetadata)
	meta.ItmRef = obj
	meta.User = user
	f.SetMeta(meta)

	_, err = io.Copy(f, data)
	if err != nil {
		err2 := f.Close()
		if err != nil {
			panic(err2)
		}
		return "", err
	}

	err = f.Close()
	return f.Id().(bson.ObjectId), err
}

func (p *ImageDBProvider) Remove(obj bson.ObjectId) error {
	return p.c.RemoveId(obj)
}

func (p *ImageDBProvider) GetImageById(obj bson.ObjectId) (*bytes.Buffer, string, error) {
	f, err := p.c.OpenId(obj)
	if err != nil {
		return nil, "", err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	if err != nil {
		err2 := f.Close()
		if err2 != nil {
			panic(err2)
		}
		return nil, "", err
	}
	ct := f.ContentType()
	err = f.Close()
	return buf, ct, err
}

func (p *ImageDBProvider) GetImageMetadataById(obj bson.ObjectId) (*ImageMetadata, error) {
	f, err := p.c.OpenId(obj)
	if err != nil {
		return nil, err
	}
	meta := new(ImageMetadata)
	err = f.GetMeta(meta)
	if err != nil {
		err2 := f.Close()
		if err2 != nil {
			panic(err2)
		}
		return nil, err
	}
	err = f.Close()
	return meta, err
}

func (p *ImageDBProvider) Delete(obj bson.ObjectId) error {
	return p.c.RemoveId(obj)
}
