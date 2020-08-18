package cmd

import (
	// "context"
	// "flag"
	"fmt"
	"log"
	// "time"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	// "runtime"
	"os"
	"net"
	"strconv"
	// "time"
)

var dbHostGlobal, dbSchemaGlobal, grpcPortGlobal string
var ConfEnv SelectedConfig

type SelectedConfig struct {		
	DatastoreDBHost string		`yaml:"db_host"`
	DatastoreDBUser string		`yaml:"db_user"`
	DatastoreDBPassword string	`yaml:"db_password"`
	DatastoreDBSchema string	`yaml:"db_schema"`
	DatastoreDBSchemaTest string	`yaml:"db_schema_test"`
	GRPCPort string				`yaml:"grpc_port"`
}

// Config is configuration for Server
type GlobalConfig struct {
	
	Local_Conf struct {		
		DatastoreDBHost string		`yaml:"db_host"`
		DatastoreDBUser string		`yaml:"db_user"`
		DatastoreDBPassword string	`yaml:"db_password"`
		DatastoreDBSchema string	`yaml:"db_schema"`
		DatastoreDBSchemaTest string	`yaml:"db_schema_test"`
		GRPCPort string				`yaml:"grpc_port"`
	}

	Staging_Conf struct {		
		DatastoreDBHost string		`yaml:"db_host"`
		DatastoreDBUser string		`yaml:"db_user"`
		DatastoreDBPassword string	`yaml:"db_password"`
		DatastoreDBSchema string	`yaml:"db_schema"`
		DatastoreDBSchemaTest string	`yaml:"db_schema_test"`
		GRPCPort string				`yaml:"grpc_port"`
	}

	Production_Conf struct {		
		DatastoreDBHost string		`yaml:"db_host"`
		DatastoreDBUser string		`yaml:"db_user"`
		DatastoreDBPassword string	`yaml:"db_password"`
		DatastoreDBSchema string	`yaml:"db_schema"`
		DatastoreDBSchemaTest string	`yaml:"db_schema_test"`
		GRPCPort string				`yaml:"grpc_port"`
	}
}

func init() {

	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
  	
	filepath := path.Join(path.Dir(dir), "../pkg/config/configGlobal.yaml")
	yamlFile, err := ioutil.ReadFile(filepath)
	
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	} 
	globalConfig := GlobalConfig{}
	err = yaml.Unmarshal([]byte(yamlFile), &globalConfig)
	if err != nil {
			log.Fatalf("error: %v", err)
	}

	switch os.Getenv("APP_ENV") {
	case "local":
		ConfEnv = globalConfig.Local_Conf
	case "staging":
		ConfEnv = globalConfig.Staging_Conf
	case "production":
		ConfEnv = globalConfig.Production_Conf
	}	
	// ConfEnv.GRPCPort = checkPort()
	// fmt.Printf("\nnew port: " + ConfEnv.GRPCPort)
}

func checkPort() string {
	// #port selected
	curPort := ConfEnv.GRPCPort
	intPort, err := strconv.ParseInt(curPort, 10, 64)
	ln, err := net.Listen("tcp", ":" + curPort)
	
	for err != nil {
		fmt.Fprintf(os.Stderr, "Can't listen on port %q: %s", curPort, err)		
		intPort += 1
		curPort = strconv.Itoa(int(intPort))			
		ln, err = net.Listen("tcp", ":" + curPort)		
	}	
	_ = ln.Close()
	return curPort
}