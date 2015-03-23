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

package main

import (
	"flag"
	//	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/openlab-aux/lsmsd/backend"
	"net/http"
	//	"time"
	"code.google.com/p/gcfg"
	"gopkg.in/mgo.v2"
)

type Config struct {
	Network struct {
		ListenTo string
	}
	Crypto struct {
		Enabled     bool
		Certificate string
		KeyFile     string
		Pepperfile  string
	}
	Database struct {
		Server string
		DB     string
	}
	Logging struct {
		Level string
	}
}

const (
	DEFAULT_CONFIG_PATH     = "./config.gcfg"
	DEFAULT_NETWORK_ADDRESS = ":8080"
	DEFAULT_LOGLEVEL        = "INFO"
	DEFAULT_CRYPTO          = false
	DEFAULT_CERTIFICATE     = "./cert.pem"
	DEFAULT_KEYFILE         = "keyfile.key"
	DEFAULT_DATABASE_SERVER = "localhost"
	DEFAULT_DATABASE_DB     = "lsmsd"
	DEFAULT_PEPPERFILE      = "./.pepper"
)

func main() {
	log.WithFields(log.Fields{"Version": "0.1"}).Info("lsmsd starting")
	var cfg Config
	var configpath = flag.String("cfgpath", DEFAULT_CONFIG_PATH, "path to your config file")
	var listento = flag.String("listento", DEFAULT_NETWORK_ADDRESS, "listen to address and port")
	var loglevel = flag.String("loglevel", DEFAULT_LOGLEVEL, "verbosity")
	var enable_crypto = flag.Bool("enablecrypto", DEFAULT_CRYPTO, "Use TLS instead of plain text")
	var certificate = flag.String("certificate", DEFAULT_CERTIFICATE, "certificate path")
	var pepper = flag.String("pepper", DEFAULT_PEPPERFILE, "path to your pepperfile")
	var keyfile = flag.String("keyfile", DEFAULT_KEYFILE, "private key path")
	var dbserver = flag.String("dbserver", DEFAULT_DATABASE_SERVER, "address of your mongo db server")
	var dbdb = flag.String("dbdb", DEFAULT_DATABASE_DB, "default database name")
	flag.Parse()
	err := gcfg.ReadFileInto(&cfg, *configpath)

	if err != nil {
		log.Fatal(err)
	}

	if *listento != DEFAULT_NETWORK_ADDRESS {
		cfg.Network.ListenTo = *listento
	}
	if *loglevel != DEFAULT_LOGLEVEL {
		cfg.Logging.Level = *loglevel
	}
	if *enable_crypto != DEFAULT_CRYPTO {
		cfg.Crypto.Enabled = *enable_crypto
	}
	if *certificate != DEFAULT_CERTIFICATE {
		cfg.Crypto.Certificate = *certificate
	}
	if *keyfile != DEFAULT_KEYFILE {
		cfg.Crypto.KeyFile = *keyfile
	}
	if *dbserver != DEFAULT_DATABASE_SERVER {
		cfg.Database.Server = *dbserver
	}
	if *dbdb != DEFAULT_DATABASE_DB {
		cfg.Database.DB = *dbdb
	}
	if *pepper != DEFAULT_PEPPERFILE {
		cfg.Crypto.Pepperfile = *pepper
	}

	switch cfg.Logging.Level {
	case "Debug":
		log.SetLevel(log.DebugLevel)
	}

	// Test DB Connection
	log.Info("Test database connection …")
	s, err := mgo.Dial(cfg.Database.Server)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Connection successful")
	defer s.Close()

	backend.RegisterDatabase(s, cfg.Database.DB)
	backend.ReadPepper(cfg.Crypto.Pepperfile)
	defer backend.CloseIDGen()
	restful.DefaultContainer.Filter(restful.DefaultContainer.OPTIONSFilter)
	restful.Add(backend.NewItemService())
	restful.Add(backend.NewUserService())
	//	restful.Add(backend.NewPolicyService())

	log.WithFields(log.Fields{"Address": cfg.Network.ListenTo, "TLS": cfg.Crypto.Enabled}).
		Info("lsms started successfully")
	if cfg.Crypto.Enabled {
		log.Fatal(http.ListenAndServeTLS(cfg.Network.ListenTo, cfg.Crypto.Certificate, cfg.Crypto.KeyFile, nil))
	} else {
		log.Fatal(http.ListenAndServe(cfg.Network.ListenTo, nil))
	}
}
