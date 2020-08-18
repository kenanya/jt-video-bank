// +build unit_video_bank_service

package v1_test

import (
	"context"
	// "errors"
	"flag"
	"reflect"
	"testing"
	"time"
	"fmt"
	"log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"os"
	"strings"
	// "strconv"
	// "github.com/sparrc/go-ping"
	// "math/rand"

	tspb "github.com/golang/protobuf/ptypes/timestamp"
	// "github.com/golang/protobuf/ptypes"
	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo" 
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/amsokol/mongo-go-driver-protobuf"
	// pmongo "github.com/amsokol/mongo-go-driver-protobuf/pmongo"

	pb_gen_v1 "github.com/kenanya/jt-video-bank/pkg/api/v1"
	svc_v1 "github.com/kenanya/jt-video-bank/pkg/service/v1"
	"github.com/kenanya/jt-video-bank/pkg/logger"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
var ConfEnv SelectedConfig

// ExtraConfig is additional configuration for Server
type ExtraConfig struct {
	// Log parameters section
	// LogLevel is global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	LogLevel int
	// LogTimeFormat is print time format for logger e.g. 2006-01-02T15:04:05Z07:00
	LogTimeFormat string
}

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

const (
	apiVersion = "v1"
  )

var globalID string
var globalGenre string
var globalCreatedAt *tspb.Timestamp
var globalContentID string
var globalTime time.Time

var curDB *mongo.Database
var curLoc *time.Location
var ctx context.Context
  
func Test_videoBankService(t *testing.T) {

	// #open config
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
  	
	filepath := path.Join(path.Dir(dir), "../config/configGlobal.yaml")
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
	default:
		log.Fatal("No environment defined.")
	}	


	// #testing
	var dbErr error
	ctx = context.Background()
	curDB, dbErr = ConnectToDBForTest(ctx)
	if dbErr != nil {
		log.Fatalf("failed initialize MongoDB connection: %#v", dbErr)
	}

	var cfg ExtraConfig
	flag.IntVar(&cfg.LogLevel, "log-level", 0, "Global log level")	
	flag.StringVar(&cfg.LogTimeFormat, "log-time-format", "2006-01-02T15:04:05Z07:00",
		"Print time format for logger e.g. 2006-01-02T15:04:05Z07:00")
	flag.Parse()

	// initialize logger
	if err := logger.Init(cfg.LogLevel, cfg.LogTimeFormat); err != nil {
		t.Errorf("failed to initialize logger: %v", err)
	}

	// Create protobuf Timestamp value from golang Time
	curLoc, err = time.LoadLocation("Asia/Jakarta")
    if err != nil {
		t.Errorf("failed to load location: %v", err)
	}

	// fmt.Println(globalDB)
	t.Run("CreateVideoBank", createVideoBank_should_insert_videoBank_into_mongo)
	t.Run("CreateBulkVideoBank", createBulkVideoBank_should_insert_videoBank_into_mongo)
	// t.Run("UpsertBulkVideoBank", upsertBulkVideoBank_should_upsert_videoBank_into_mongo)
	t.Run("ReadVideoBank", readVideoBank_should_retrieve_one_videoBank)
	t.Run("ReadAllVideoBank", readAllVideoBank_should_retrieve_all_videoBank)

	t.Run("GetGenreListVideoBank", getGenreList_should_retrieve_all_genre_videoBank)
	// t.Run("UpdateVideoBank", updateVideoBank_should_update_one_videoBank)		
	// t.Run("DeleteVideoBank", deleteVideoBank_should_delete_one_videoBank)	
}

func ConnectToDBForTest(ctx context.Context) (*mongo.Database, error) {
	
	reg := codecs.Register(bson.NewRegistryBuilder()).Build()
	uri := fmt.Sprintf(`mongodb://%s:%s@%s/%s`,
		ConfEnv.DatastoreDBUser,
		ConfEnv.DatastoreDBPassword,
		ConfEnv.DatastoreDBHost,
		ConfEnv.DatastoreDBSchemaTest,
	)

	client, err := mongo.NewClient(
		options.Client().ApplyURI(uri),
		&options.ClientOptions{
			Registry: reg,
		})

	if err != nil {
		log.Fatalf("failed to create new MongoDB client: %#v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// connect client
	if err = client.Connect(ctx); err != nil {
		log.Fatalf("failed to connect to MongoDB: %#v", err)
	}

	db := client.Database(ConfEnv.DatastoreDBSchemaTest)
	// fmt.Println(db)
	return db, err
}
  
func createVideoBank_should_insert_videoBank_into_mongo(t *testing.T) {
	  
	svc_v1_server := svc_v1.NewVideoBankServiceServer(curDB, curLoc)


	strVideoType := "THAI_MV"
	arrGenre := []string{"ANIME"}
	arrCast := []string{"Kevin Bacon", "Elisabeth Shue"}
	strContentType := "FREE"
	tagList := []string{"anime2019", "japan", "school"}
	directorList := []string{"Noriyaki Akitaya", "Dominic Sena"}

	valVideoType, ok := pb_gen_v1.MV_VideoTypeOpt_value[strVideoType]
	if !ok {
		t.Error("Invalid enum value of video type")		
	}
	
	//tambah genre
	genreList := []pb_gen_v1.MV_GenreOpt{}
	for _, a := range arrGenre {			
		valGenreType, ok := pb_gen_v1.MV_GenreOpt_value[a]
		if !ok {
			t.Error("Invalid enum value of genre")		
		}
		genreList = append(genreList, pb_gen_v1.MV_GenreOpt(valGenreType))
		if globalGenre == "" {
			globalGenre = a
			fmt.Printf("Global Genre: `%s` \n\n", globalGenre)
		}
	}


	valContentType, ok := pb_gen_v1.MV_ContentTypeOpt_value[strContentType]
	if !ok {		
		t.Error("Invalid enum value of content type")
	}

	xValVideoType := pb_gen_v1.MV_VideoTypeOpt(valVideoType)
	xValContentType := pb_gen_v1.MV_ContentTypeOpt(valContentType)

	// Call Create
	req1 := pb_gen_v1.MV_CreateRequest{
		Api: apiVersion,
		VideoBank: &pb_gen_v1.VideoBank{	
			ContentId: "20000001",		
			Provider: "Sushiroll",
			ProviderShort: "SRO",			
			ProviderLabel: "https://assets.genflix.co.id/poster/movies/303/cu_TotalChaos.1280.jpg",   // image_url
			Title: "Ms. vampire who lives in my neighborhood Eps 1",
			// TitlePackage: "Ms. vampire who lives in my neighborhood.",
    		VideoType: xValVideoType,
			Genre: genreList,
			Tags: tagList,
			Year: 2018,
			Duration: "02:10:20", 
			Synopsis: "Human girl Amano Akari is rescued by a vampire, Sophie Twilight, and falls in love with her. She forces herself into Sophie’s home and begins living with her. Though Sophie is a vampire, she never attacks humans, instead buying blood and anime merchandise online like any ordinary person. A modern-day vampire comedy!",
			Cast: arrCast,
			PlayerUrl: "http://sushiroll.co.id/smartfren/play/662",

			Poster: &pb_gen_v1.MV_PosterData{
				S: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
                M: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
			},

			Director: directorList,
			ContentType: xValContentType,
			Availability: "SVOD",
			ContentAs: "MOVIE",
			ContentLevel: 10,
			IsActive: true,				
			IsValid: true,				
		},	
	}


	res1, err := svc_v1_server.Create(ctx, &req1)
	// Assert
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}

	id := res1.Id
	globalID = id.Value
	expectedType := reflect.TypeOf(pb_gen_v1.VideoBank{}.Id)
	if reflect.TypeOf(pb_gen_v1.VideoBank{}.Id) != reflect.TypeOf(id) {
		t.Errorf("Incorrect Type Of Id `%s`. Expected `%s`, Got: `%s`", globalID, expectedType, reflect.TypeOf(id))
	}
	fmt.Printf("Inserted ID: `%s` \n\n", globalID)
}

func createBulkVideoBank_should_insert_videoBank_into_mongo(t *testing.T) {
	  
	svc_v1_server := svc_v1.NewVideoBankServiceServer(curDB, curLoc)


	strVideoType := remapVideoType("HOL_MV")
	arrGenre := []string{"ANIME", "SCI_FI_FANTASY", "MARTIAL_ARTS"}
	arrCast := []string{"Kevin Bacon", "Elisabeth Shue"}
	strContentType := "FREE"
	tagList := []string{"anime2019", "japan", "school"}
	directorList := []string{"Noriyaki Akitaya", "Dominic Sena"}

	valVideoType, ok := pb_gen_v1.MV_VideoTypeOpt_value[strVideoType]
	if !ok {
		t.Error("Invalid enum value of video type")		
	}
	
	//tambah genre
	genreList := []pb_gen_v1.MV_GenreOpt{}
	for _, a := range arrGenre {			
		valGenreType, ok := pb_gen_v1.MV_GenreOpt_value[a]
		if !ok {
			t.Error("Invalid enum value of genre")		
		}
		genreList = append(genreList, pb_gen_v1.MV_GenreOpt(valGenreType))
		if globalGenre == "" {
			globalGenre = a
			fmt.Printf("Global Genre: `%s` \n\n", globalGenre)
		}
	}


	valContentType, ok := pb_gen_v1.MV_ContentTypeOpt_value[strContentType]
	if !ok {		
		t.Error("Invalid enum value of content type")
	}

	xValVideoType := pb_gen_v1.MV_VideoTypeOpt(valVideoType)
	xValContentType := pb_gen_v1.MV_ContentTypeOpt(valContentType)

	// Call Create
	req1 := pb_gen_v1.MV_CreateBulkRequest{
		Api: apiVersion,
		VideoBankList: []*pb_gen_v1.VideoBank{	
			&pb_gen_v1.VideoBank{
				ContentId: "20000002",							
				Provider: "Sushiroll",
				ProviderShort: "SRO",				
				ProviderLabel: "https://assets.genflix.co.id/poster/movies/303/cu_TotalChaos.1280.jpg",   // image_url
				Title: "SRO 1 - Ms. vampire who lives in my neighborhood Eps 1",
				// TitlePackage: "SRO 1 - Ms. vampire who lives in my neighborhood.",
				VideoType: xValVideoType,
				Genre: genreList,
				Tags: tagList,
				Year: 2010,
				Duration: "01:20:20", 				
				Synopsis: "SRO 1 - Human girl Amano Akari is rescued by a vampire, Sophie Twilight, and falls in love with her. She forces herself into Sophie’s home and begins living with her. Though Sophie is a vampire, she never attacks humans, instead buying blood and anime merchandise online like any ordinary person. A modern-day vampire comedy!",
				Cast: arrCast,
				PlayerUrl: "http://sushiroll.co.id/smartfren/play/662",
			
				Poster: &pb_gen_v1.MV_PosterData{
					S: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
					M: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
				},
							
				Director: directorList,
				ContentType: xValContentType,
				Availability: "SVOD",
				ContentAs: "MOVIE",
				ContentLevel: 10,
				IsActive: true,				
				IsValid: true,	
			},	
			&pb_gen_v1.VideoBank{			
				ContentId: "20000003",
				Provider: "Genflix",
				ProviderShort: "GEN",				
				ProviderLabel: "https://assets.genflix.co.id/poster/movies/303/cu_TotalChaos.1280.jpg",   // image_url
				Title: "GEN - Ms. vampire who lives in my neighborhood Eps 1",
				// TitlePackage: "GEN - Ms. vampire who lives in my neighborhood.",
				VideoType: xValVideoType,
				Genre: genreList,
				Tags: tagList,
				Year: 2011,
				Duration: "01:30:20", 				
				Synopsis: "GEN - Human girl Amano Akari is rescued by a vampire, Sophie Twilight, and falls in love with her. She forces herself into Sophie’s home and begins living with her. Though Sophie is a vampire, she never attacks humans, instead buying blood and anime merchandise online like any ordinary person. A modern-day vampire comedy!",
				Cast: arrCast,
				PlayerUrl: "http://sushiroll.co.id/smartfren/play/662",
			
				Poster: &pb_gen_v1.MV_PosterData{
					S: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
					M: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
				},
			
				Director: directorList,
				ContentType: xValContentType,
				Availability: "SVOD",
				ContentAs: "MOVIE",
				ContentLevel: 10,
				IsActive: true,				
				IsValid: true,
			},
			&pb_gen_v1.VideoBank{	
				ContentId: "20000004",		
				Provider: "Genflix",
				ProviderShort: "GEN",				
				ProviderLabel: "https://assets.genflix.co.id/poster/movies/303/cu_TotalChaos.1280.jpg",   // image_url
				Title: "GEN 2 - Ms. vampire who lives in my neighborhood Eps 1",
				// TitlePackage: "GEN 2 - Ms. vampire who lives in my neighborhood.",
				VideoType: xValVideoType,
				Genre: genreList,
				Tags: tagList,
				Year: 2012,
				Duration: "00:50:20", 				
				Synopsis: "GEN 2 - Human girl Amano Akari is rescued by a vampire, Sophie Twilight, and falls in love with her. She forces herself into Sophie’s home and begins living with her. Though Sophie is a vampire, she never attacks humans, instead buying blood and anime merchandise online like any ordinary person. A modern-day vampire comedy!",
				Cast: arrCast,
				PlayerUrl: "http://sushiroll.co.id/smartfren/play/662",
			
				Poster: &pb_gen_v1.MV_PosterData{
					S: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
					M: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
				},
			
				Director: directorList,
				ContentType: xValContentType,
				Availability: "SVOD",
				ContentAs: "MOVIE",
				ContentLevel: 10,
				IsActive: true,				
				IsValid: true,	
			},
			&pb_gen_v1.VideoBank{	
				ContentId: "20000005",	
				Provider: "Sushiroll",
				ProviderShort: "SRO",				
				ProviderLabel: "https://assets.genflix.co.id/poster/movies/303/cu_TotalChaos.1280.jpg",   // image_url
				Title: "SRO 2 - Ms. vampire who lives in my neighborhood Eps 1",
				// TitlePackage: "SRO 2 - Ms. vampire who lives in my neighborhood.",
				VideoType: xValVideoType,
				Genre: genreList,
				Tags: tagList,
				Year: 2013,
				Duration: "01:00:40", 				
				Synopsis: "SRO 2 - Human girl Amano Akari is rescued by a vampire, Sophie Twilight, and falls in love with her. She forces herself into Sophie’s home and begins living with her. Though Sophie is a vampire, she never attacks humans, instead buying blood and anime merchandise online like any ordinary person. A modern-day vampire comedy!",
				Cast: arrCast,
				PlayerUrl: "http://sushiroll.co.id/smartfren/play/662",
			
				Poster: &pb_gen_v1.MV_PosterData{
					S: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
					M: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
				},
			
				Director: directorList,
				ContentType: xValContentType,
				Availability: "SVOD",
				ContentAs: "MOVIE",
				ContentLevel: 10,
				IsActive: true,				
				IsValid: true,	
			},	
		},
	}


	res1, err := svc_v1_server.CreateBulk(ctx, &req1)	
	// Assert
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	// fmt.Printf("result: <%+v>\n\n", res1)
	// fmt.Println("end of create bulk test")
	
	for counter, xId := range res1.Ids {
		id := xId
		// globalID = id.Value
		expectedType := reflect.TypeOf(pb_gen_v1.VideoBank{}.Id)
		if reflect.TypeOf(pb_gen_v1.VideoBank{}.Id) != reflect.TypeOf(id) {
			t.Errorf("Incorrect Type Of Id `%s`. Expected `%s`, Got: `%s`", id, expectedType, reflect.TypeOf(id))
		}
		fmt.Printf("Inserted Bulk ID #%d: `%s` \n\n", counter, id.Value)
	}
	
}

// func upsertBulkVideoBank_should_upsert_videoBank_into_mongo(t *testing.T) {
	  
// 	svc_v1_server := svc_v1.NewVideoBankServiceServer(curDB, curLoc)


// 	strVideoType := "TV_SERIES"
// 	arrGenre := []string{"ANIME"}
// 	arrCast := []string{"Kevin Bacon", "Elisabeth Shue"}
// 	strContentType := "FREE"
// 	tagList := []string{"anime2019", "japan", "school"}
// 	directorList := []string{"Noriyaki Akitaya", "Dominic Sena"}

// 	valVideoType, ok := pb_gen_v1.MV_VideoTypeOpt_value[strVideoType]
// 	if !ok {
// 		t.Error("Invalid enum value of video type")		
// 	}
	
// 	//tambah genre
// 	genreList := []pb_gen_v1.MV_GenreOpt{}
// 	for _, a := range arrGenre {			
// 		valGenreType, ok := pb_gen_v1.MV_GenreOpt_value[a]
// 		if !ok {
// 			t.Error("Invalid enum value of genre")		
// 		}
// 		genreList = append(genreList, pb_gen_v1.MV_GenreOpt(valGenreType))
// 		if globalGenre == "" {
// 			globalGenre = a
// 			fmt.Printf("Global Genre: `%s` \n\n", globalGenre)
// 		}
// 	}


// 	valContentType, ok := pb_gen_v1.MV_ContentTypeOpt_value[strContentType]
// 	if !ok {		
// 		t.Error("Invalid enum value of content type")
// 	}

// 	xValVideoType := pb_gen_v1.MV_VideoTypeOpt(valVideoType)
// 	xValContentType := pb_gen_v1.MV_ContentTypeOpt(valContentType)

// 	// Call Upsert
// 	req1 := pb_gen_v1.MV_UpsertBulkRequest{
// 		Api: apiVersion,
// 		VideoBankList: []*pb_gen_v1.VideoBank{	
// 			&pb_gen_v1.VideoBank{		
// 				Provider: "Sushiroll",
// 				ProviderShort: "SRO",
// 				ContentId: 100000002,
// 				ProviderLabel: "https://assets.genflix.co.id/poster/movies/303/cu_TotalChaos.1280.jpg",   // image_url
// 				Title: "[UPDATE] SRO 1 - Ms. vampire who lives in my neighborhood Eps 1",
// 				// TitlePackage: "SRO 1 - Ms. vampire who lives in my neighborhood.",
// 				VideoType: xValVideoType,
// 				Genre: genreList,
// 				Tags: tagList,
// 				Year: 2010,
// 				Duration: "01:20:20", 
// 				ContentId: 662, 
// 				Synopsis: "SRO 1 - Human girl Amano Akari is rescued by a vampire, Sophie Twilight, and falls in love with her. She forces herself into Sophie’s home and begins living with her. Though Sophie is a vampire, she never attacks humans, instead buying blood and anime merchandise online like any ordinary person. A modern-day vampire comedy!",
// 				Cast: arrCast,
// 				PlayerUrl: "http://sushiroll.co.id/smartfren/play/662",
			
// 				Poster: &pb_gen_v1.MV_PosterData{
// 					S: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
// 					M: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
// 				},
							
// 				Director: directorList,
// 				ContentType: xValContentType,
// 				IsActive: true,				
// 			},	
// 			&pb_gen_v1.VideoBank{			
// 				Provider: "Genflix",
// 				ProviderShort: "GEN",
// 				ContentId: 100000001,
// 				ProviderLabel: "https://assets.genflix.co.id/poster/movies/303/cu_TotalChaos.1280.jpg",   // image_url
// 				Title: "[UPDATE] GEN - Ms. vampire who lives in my neighborhood Eps 1",
// 				// TitlePackage: "GEN - Ms. vampire who lives in my neighborhood.",
// 				VideoType: xValVideoType,
// 				Genre: genreList,
// 				Tags: tagList,
// 				Year: 2011,
// 				Duration: "01:30:20", 
// 				ContentId: 662, 
// 				Synopsis: "GEN - Human girl Amano Akari is rescued by a vampire, Sophie Twilight, and falls in love with her. She forces herself into Sophie’s home and begins living with her. Though Sophie is a vampire, she never attacks humans, instead buying blood and anime merchandise online like any ordinary person. A modern-day vampire comedy!",
// 				Cast: arrCast,
// 				PlayerUrl: "http://sushiroll.co.id/smartfren/play/662",
			
// 				Poster: &pb_gen_v1.MV_PosterData{
// 					S: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
// 					M: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
// 				},
			
// 				Director: directorList,
// 				ContentType: xValContentType,
// 				IsActive: true,				
// 			},
// 			&pb_gen_v1.VideoBank{			
// 				Provider: "Genflix",
// 				ProviderShort: "GEN",
// 				ContentId: 100000002,
// 				ProviderLabel: "https://assets.genflix.co.id/poster/movies/303/cu_TotalChaos.1280.jpg",   // image_url
// 				Title: "[UPDATE] GEN 2 - Ms. vampire who lives in my neighborhood Eps 1",
// 				// TitlePackage: "GEN 2 - Ms. vampire who lives in my neighborhood.",
// 				VideoType: xValVideoType,
// 				Genre: genreList,
// 				Tags: tagList,
// 				Year: 2012,
// 				Duration: "00:50:20", 
// 				ContentId: 662, 
// 				Synopsis: "GEN 2 - Human girl Amano Akari is rescued by a vampire, Sophie Twilight, and falls in love with her. She forces herself into Sophie’s home and begins living with her. Though Sophie is a vampire, she never attacks humans, instead buying blood and anime merchandise online like any ordinary person. A modern-day vampire comedy!",
// 				Cast: arrCast,
// 				PlayerUrl: "http://sushiroll.co.id/smartfren/play/662",
			
// 				Poster: &pb_gen_v1.MV_PosterData{
// 					S: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
// 					M: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
// 				},
			
// 				Director: directorList,
// 				ContentType: xValContentType,
// 				IsActive: true,				
// 			},
// 			&pb_gen_v1.VideoBank{		
// 				Provider: "Sushiroll",
// 				ProviderShort: "SRO",
// 				ContentId: 100000004,
// 				ProviderLabel: "https://assets.genflix.co.id/poster/movies/303/cu_TotalChaos.1280.jpg",   // image_url
// 				Title: "[INSERT] SRO 2 - Ms. vampire who lives in my neighborhood Eps 1",
// 				// TitlePackage: "SRO 2 - Ms. vampire who lives in my neighborhood.",
// 				VideoType: xValVideoType,
// 				Genre: genreList,
// 				Tags: tagList,
// 				Year: 2013,
// 				Duration: "01:00:40", 
// 				ContentId: 662, 
// 				Synopsis: "SRO 2 - Human girl Amano Akari is rescued by a vampire, Sophie Twilight, and falls in love with her. She forces herself into Sophie’s home and begins living with her. Though Sophie is a vampire, she never attacks humans, instead buying blood and anime merchandise online like any ordinary person. A modern-day vampire comedy!",
// 				Cast: arrCast,
// 				PlayerUrl: "http://sushiroll.co.id/smartfren/play/662",
			
// 				Poster: &pb_gen_v1.MV_PosterData{
// 					S: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
// 					M: "http://sushiroll.co.id/assets/_preview/vid/___662___/vampire_eps_1.jpg.png",
// 				},
			
// 				Director: directorList,
// 				ContentType: xValContentType,
// 				IsActive: true,				
// 			},	
// 		},
// 	}


// 	res1, err := svc_v1_server.UpsertBulk(ctx, &req1)	
// 	// Assert
// 	if err != nil {
// 		t.Errorf("Upsert failed: %v", err)
// 	}
// 	fmt.Printf("result: <%+v>\n\n", res1)
// 	fmt.Println("end of Upsert bulk test")
	
// 	// for counter, xId := range res1.Ids {
// 	// 	id := xId
// 	// 	// globalID = id.Value
// 	// 	expectedType := reflect.TypeOf(pb_gen_v1.VideoBank{}.Id)
// 	// 	if reflect.TypeOf(pb_gen_v1.VideoBank{}.Id) != reflect.TypeOf(id) {
// 	// 		t.Errorf("Incorrect Type Of Id `%s`. Expected `%s`, Got: `%s`", id, expectedType, reflect.TypeOf(id))
// 	// 	}
// 	// 	fmt.Printf("Upserted Bulk ID #%d: `%s` \n\n", counter, id.Value)
// 	// }
	
// }

func readVideoBank_should_retrieve_one_videoBank(t *testing.T) {
	
	svc_v1_server := svc_v1.NewVideoBankServiceServer(curDB, curLoc)
	req1 := pb_gen_v1.MV_ReadRequest{
		Api: apiVersion,
		Id:  globalID,
		Idx: "",
	}
	res1, err := svc_v1_server.Read(ctx, &req1)
	getID := res1.VideoBank.Id.Value
	globalContentID = res1.VideoBank.ContentId
	globalCreatedAt = res1.VideoBank.CreatedAt
	// globalGenre = res1.GenreOpt_value

	//Assert
	if err != nil {
		t.Errorf("Read failed: %v", err)
	}

	if getID != globalID {
	  t.Errorf("Incorrect ID. Expected `%s`, Got: `%s`", globalID, getID)
	}

	if (globalContentID == "") {
		t.Errorf("Invalid Value of ContentId: %v", err)
	}
}

func readAllVideoBank_should_retrieve_all_videoBank(t *testing.T) {

	svc_v1_server := svc_v1.NewVideoBankServiceServer(curDB, curLoc)
	req1 := pb_gen_v1.MV_ReadAllRequest{
		Api: apiVersion,
		Title: "",
		VideoType: "", 
		Genre: globalGenre,
		Skip: 0,
		Limit: 100,
	}
	res1, err := svc_v1_server.ReadAll(ctx, &req1)
	//Assert
	if err != nil {
		t.Errorf("ReadAll failed: %v", err)
	}
	// log.Printf("ReadAll result: <%+v>\n\n", res1)

	items := []pb_gen_v1.VideoBank{}
	for _, a := range res1.VideoBanks {				
		vb := pb_gen_v1.VideoBank{
			Id:   a.Id,
			// Mdn : a.Mdn,
			// User: a.User,
			ContentId: a.ContentId, 
			Idx: a.Idx,
			Provider: a.Provider,
			ProviderShort: a.ProviderShort,
			ProviderLabel: a.ProviderLabel,
			Title: a.Title,
    		VideoType: a.VideoType,
			Genre: a.Genre,
			Tags: a.Tags,
			Year: a.Year,
			Duration: a.Duration, 			
			Synopsis: a.Synopsis,
			Cast: a.Cast,
			PlayerUrl: a.PlayerUrl,
			Poster: a.Poster,
			Director: a.Director,
			ContentType: a.ContentType,
			Availability: a.Availability,
			ContentAs: a.ContentAs,
			ContentLevel: a.ContentLevel,
			IsActive: a.IsActive,	
			IsValid: a.IsValid,	
		}
		items = append(items, vb)
	}
	// log.Printf("ReadAll results log: <%+v>\n\n", items)
	fmt.Printf("\nReadAll results fmt: %+v\n", items)
}

func getGenreList_should_retrieve_all_genre_videoBank(t *testing.T) {
	
	svc_v1_server := svc_v1.NewVideoBankServiceServer(curDB, curLoc)
	req1 := pb_gen_v1.MV_SetGenreRequest{
		Api: apiVersion,
	}
	res1, err := svc_v1_server.GetGenreList(ctx, &req1)
	
	//Assert
	if err != nil {
		t.Errorf("Get Genre List failed: %v", err)
	}
	fmt.Printf("Genre List: %+v\n\n", res1)
}


func remapVideoType(tempType string) string {
	finalType := strings.ToUpper(tempType) 
	if finalType == "HOL_MV" {
		finalType = "ENGLISH_MV"
	}
	if finalType == "IND_MV" {
		finalType = "INDONESIAN_MV"
	}
	return finalType
}