package cmd

import (
	"context"
	"flag"
	"fmt"
	// "log"
	"time"
	"go.uber.org/zap"
	// "go.uber.org/zap/zapcore"

	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo" 
	"go.mongodb.org/mongo-driver/mongo/options" 

	"github.com/amsokol/mongo-go-driver-protobuf"
	// pmongo "github.com/amsokol/mongo-go-driver-protobuf/pmongo"

	"github.com/kenanya/jt-video-bank/pkg/logger"
	"github.com/kenanya/jt-video-bank/pkg/protocol/grpc"
	"github.com/kenanya/jt-video-bank/pkg/service/v1"
)

// ExtraConfig is additional configuration for Server
type ExtraConfig struct {
	// Log parameters section
	// LogLevel is global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	LogLevel int
	// LogTimeFormat is print time format for logger e.g. 2006-01-02T15:04:05Z07:00
	LogTimeFormat string
}

func ConnectToDB(ctx context.Context) (*mongo.Database, error) {
	
	reg := codecs.Register(bson.NewRegistryBuilder()).Build()
	// log.Printf("connecting to MongoDB...")
	logger.Log.Info("connecting to MongoDB...")

	uri := fmt.Sprintf(`mongodb://%s:%s@%s/%s`,
		ConfEnv.DatastoreDBUser,
		ConfEnv.DatastoreDBPassword,
		ConfEnv.DatastoreDBHost,
		ConfEnv.DatastoreDBSchema,
	)

	client, err := mongo.NewClient(
		options.Client().ApplyURI(uri),
		&options.ClientOptions{
			Registry: reg,
		})

	if err != nil {
		// log.Fatalf("failed to create new MongoDB client: %#v", err)
		logger.Log.Fatal("failed to create new MongoDB client", zap.String("reason", err.Error()))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// connect client
	if err = client.Connect(ctx); err != nil {
		// log.Fatalf("failed to connect to MongoDB: %#v", err)
		logger.Log.Fatal("failed to connect to MongoDB", zap.String("reason", err.Error()))
	}
	// log.Printf("connected successfully")
	logger.Log.Info("connected successfully")

	db := client.Database(ConfEnv.DatastoreDBSchema)
	return db, err
}

// RunServer runs gRPC server and HTTP gateway
func RunServer() error {

	var cfg ExtraConfig
	ctx := context.Background()
	
	// fmt.Println("\n##### selected config #####")
	// fmt.Println(ConfEnv)

	flag.IntVar(&cfg.LogLevel, "log-level", 0, "Global log level")
	flag.StringVar(&cfg.LogTimeFormat, "log-time-format", "2006-01-02T15:04:05Z07:00",
		"Print time format for logger e.g. 2006-01-02T15:04:05Z07:00")
	flag.Parse()

	// initialize logger
	if err := logger.Init(cfg.LogLevel, cfg.LogTimeFormat); err != nil {
		return fmt.Errorf("failed to initialize logger: %v", err)
	}

	if len(ConfEnv.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", ConfEnv.GRPCPort)
	}

	// Create protobuf Timestamp value from golang Time
	globalLoc, err := time.LoadLocation("Asia/Jakarta")
    if err != nil {				
		return fmt.Errorf("failed to load location %v", err)
	}

	db, err := ConnectToDB(ctx)
	if err != nil {		
		logger.Log.Fatal("failed initialize MongoDB connection", zap.String("reason", err.Error()))
		return fmt.Errorf("failed initialize MongoDB connection: %v", err)
	}
	v1API := v1.NewVideoBankServiceServer(db, globalLoc)
	return grpc.RunServer(ctx, v1API, ConfEnv.GRPCPort)
}


