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
	"database/sql"
	"flag"
	//	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	_ "github.com/mattn/go-sqlite3"
	"github.com/openlab-aux/lsms/backend"
	"net/http"
	//	"time"
	"code.google.com/p/gcfg"
)

type Config struct {
	Network struct {
		ListenTo string
	}
	Crypto struct {
		Enabled     bool
		Certificate string
		KeyFile     string
	}
	Database struct {
		Path string
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
)

func main() {
	log.WithFields(log.Fields{"Version": "0.1"}).Info("lsmsd starting")
	var cfg Config
	var configpath = flag.String("cfgpath", DEFAULT_CONFIG_PATH, "path to your config file")
	var listento = flag.String("listento", DEFAULT_NETWORK_ADDRESS, "listen to address and port")
	var loglevel = flag.String("loglevel", DEFAULT_LOGLEVEL, "verbosity")
	var enable_crypto = flag.Bool("enablecrypto", DEFAULT_CRYPTO, "Use TLS instead of plain text")
	var certificate = flag.String("certificate", DEFAULT_CERTIFICATE, "certificate path")
	var keyfile = flag.String("keyfile", DEFAULT_KEYFILE, "private key path")
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

	switch cfg.Logging.Level {
	case "Debug":
		log.SetLevel(log.DebugLevel)
	}

	db, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		log.Fatal(err)
	}

	backend.RegisterDatabase(db)
	restful.DefaultContainer.Filter(restful.DefaultContainer.OPTIONSFilter)
	restful.Add(backend.NewItemService())
	restful.Add(backend.NewUserService())
	restful.Add(backend.NewPolicyService())

	log.WithFields(log.Fields{"Address": cfg.Network.ListenTo, "TLS": cfg.Crypto.Enabled}).
		Info("lsms started successfully")
	if cfg.Crypto.Enabled {
		log.Fatal(http.ListenAndServeTLS(cfg.Network.ListenTo, cfg.Crypto.Certificate, cfg.Crypto.KeyFile, nil))
	} else {
		log.Fatal(http.ListenAndServe(cfg.Network.ListenTo, nil))
	}
}
