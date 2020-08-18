package v1

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"math"
	"math/rand"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/bson"
	// "go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// "go.mongodb.org/mongo-driver/mongo/readconcern"
	// "go.mongodb.org/mongo-driver/mongo/readpref"
	// "go.mongodb.org/mongo-driver/mongo/writeconcern"


	"github.com/amsokol/mongo-go-driver-protobuf/pmongo"

	"github.com/kenanya/jt-video-bank/pkg/api/v1"
	"github.com/kenanya/jt-video-bank/pkg/logger"
)

type key string

const (
	apiVersion  = "v1"
	hostKey     = key("hostKey")
	usernameKey = key("usernameKey")
	passwordKey = key("passwordKey")
	databaseKey = key("databaseKey")
)

const collName = "jt_video_bank"

type videoBankServiceServer struct {
	db *mongo.Database
	globalLoc *time.Location
}

func NewVideoBankServiceServer(db *mongo.Database, globalLoc *time.Location) v1.VideoBankServiceServer {
	return &videoBankServiceServer{db: db, globalLoc: globalLoc}
}

func (s *videoBankServiceServer) checkAPI(api string) error {
	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented,
				"unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

func (s *videoBankServiceServer) Create(ctx context.Context, req *v1.MV_CreateRequest) (*v1.MV_CreateResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	curDB := s.db
	coll := curDB.Collection("jt_video_bank")

	//get new seq Id
	// freshId := s.GetFreshContentID(ctx, req.VideoBank.ProviderShort)
	// freshId := int64(3)

	t := time.Now().In(s.globalLoc)
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		// log.Fatalf("failed to convert golang Time to protobuf Timestamp: %#v", err)
		logger.Log.Error("failed to convert golang Time to protobuf Timestamp", zap.String("reason", err.Error()))
	}

	//copy map
	in := req.VideoBank
	in.CreatedAt = ts
	// in.ContentId = freshId
	in.Idx = in.ProviderShort + "_" + in.ContentId

	// log.Printf("insert data into collection <jt_db.jt_video_bank>...")
	// fmt.Printf("%+v\n", in)
	// fmt.Printf("%+v\n", v1.VideoBank_BuyOption)

	// Insert data into the collection
	res, err := coll.InsertOne(ctx, &in)
	if err != nil {
		// log.Fatalf("insert data into collection <%v.jt_video_bank>: %#v", curDB, err)
		logger.Log.Error("insert data into collection <jt_db.jt_video_bank>", zap.String("reason", err.Error()))
	}
	id := pmongo.NewObjectId(res.InsertedID.(primitive.ObjectID))

	return &v1.MV_CreateResponse{
		Api: apiVersion,
		Id:  id,
	}, nil
}

func (s *videoBankServiceServer) CreateBulk(ctx context.Context, reqs *v1.MV_CreateBulkRequest) (*v1.MV_CreateBulkResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(reqs.Api); err != nil {
		return nil, err
	}

	t := time.Now().In(s.globalLoc)
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		logger.Log.Error("failed to convert golang Time to protobuf Timestamp", zap.String("reason", err.Error()))
	}

	curDB := s.db
	coll := curDB.Collection("jt_video_bank_"+t.Format("20060102"))

	inBulk := []interface{}{}
	// var freshId = make(map[string]int64)
	idxList := []string{}
	for _, xReq := range reqs.VideoBankList {		

		xReq.CreatedAt = ts				
		if xReq.Idx == "" {
			xReq.Idx = xReq.ProviderShort + "_" + xReq.ContentId		
		}	
		inBulk = append(inBulk, xReq)

		// collect all id to be deleted
		idxList = append(idxList, xReq.Idx)
	}
	delfilter := bson.D{{
		"idx",
		bson.D{{
			"$in",
			idxList}}}}

	// bson.D{{"status", bson.D{{"$in", bson.A{"A", "D"}}}}}
	fmt.Printf("delfilter: %+v\n\n\n", delfilter)
	deleteResult, err := coll.DeleteMany(ctx, delfilter)
	if err != nil {
		logger.Log.Error("failed remove idx ", zap.String("reason", err.Error()))
	}
	fmt.Printf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)

	insertManyResult, err := coll.InsertMany(ctx, inBulk)
	if err != nil {		
		logger.Log.Error("insert bulk data into collection <jt_db.jt_video_bank>", zap.String("reason", err.Error()))	
	}
	// fmt.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)
	// id := pmongo.NewObjectId(res.InsertedID.(primitive.ObjectID))
	var insertIDs []*pmongo.ObjectId
	for _, xInsertedID := range insertManyResult.InsertedIDs {
		insertIDs = append(insertIDs, pmongo.NewObjectId(xInsertedID.(primitive.ObjectID)))
	}

	return &v1.MV_CreateBulkResponse{
		Api: apiVersion,
		Ids: insertIDs,
	}, nil
}

func (s *videoBankServiceServer) Read(ctx context.Context, req *v1.MV_ReadRequest) (*v1.MV_ReadResponse, error) {

	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	curDB := s.db
	coll := curDB.Collection("jt_video_bank")

	// Create filter and output structure to read data from collection
	var out *v1.VideoBank
	var filter = bson.D{}
	if req.Id != "" {
		id, err := primitive.ObjectIDFromHex(req.Id)
		if err != nil {
			logger.Log.Error("Fail converting Hex to ObjectID", zap.String("reason", err.Error()))
		}
		filter = bson.D{
			{Key: "_id", Value: id},
		}
	} else if req.Idx != "" {
		filter = bson.D{
			{Key: "idx", Value: req.Idx},
		}
	} else {
		logger.Log.Error("Either of 'Id' or 'Idx' Required")
	}
	fmt.Printf("%+v\n\n\n", filter)

	// // Read data from collection
	err := coll.FindOne(ctx, filter).Decode(&out)
	if err != nil {
		// logger.Log.Info("filter data =%v", zap.String("reason", filter.(string)))
		logger.Log.Error("Fail to read data from collection <jt_db.jt_video_bank>", zap.String("reason", err.Error()))
		return nil, err
	}
	// fmt.Printf("%#v\n\n\n", out)

	return &v1.MV_ReadResponse{
		Api:       apiVersion,
		VideoBank: out,
	}, nil
}

func (s *videoBankServiceServer) ReadAll(ctx context.Context, req *v1.MV_ReadAllRequest) (*v1.MV_ReadAllResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	tNow := time.Now().In(s.globalLoc)

	curDB := s.db
	coll := curDB.Collection("jt_video_bank")
	findOptions := options.Find()
	findOptions.SetLimit(req.Limit)
	findOptions.SetSkip(req.Skip)

	filter, arrCondition := bson.D{}, bson.A{}
	arrCondition = append(arrCondition, bson.D{{"isactive", true},})
	arrCondition = append(arrCondition, bson.D{{"expireddate", bson.D{{"$gt", tNow}} },})

	if req.Title != "" {
		arrCondition = append(arrCondition, bson.D{{"title", primitive.Regex{Pattern: req.Title, Options: "i"}}})
	}	
	if req.VideoType != "" {
		valVideoType, ok := v1.MV_VideoTypeOpt_value[req.VideoType]
		if !ok {
			logger.Log.Error("readAll: Invalid enum value of video type")
			return nil, status.Errorf(codes.InvalidArgument, "readAll: Invalid enum value of video type '%s'", req.VideoType)
		}
		arrCondition = append(arrCondition, bson.D{{"videotype", valVideoType},})
	} 	
	if req.Genre != "" {
		valGenreType, ok := v1.MV_GenreOpt_value[req.Genre]
		if !ok {
			logger.Log.Error("readAll: Invalid enum value of genre")
			return nil, status.Errorf(codes.InvalidArgument, "readAll: Invalid enum value of genre '%s'", req.Genre)
		}
		arrCondition = append(arrCondition, bson.D{{"genre", valGenreType},})
	}
	filter = bson.D{{"$and", arrCondition}}
	fmt.Printf("filter: %+v\n\n\n", filter)

	c, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		logger.Log.Error("readAll: couldn't list all video bank", zap.String("reason", err.Error()))
		return nil, status.Errorf(codes.Unknown, "readAll: couldn't list all video bank '%s'", err.Error())
	}
	defer c.Close(ctx)

	// fmt.Println("==================raw c=================")
	// fmt.Printf("%+v\n\n\n", c)
	var videoBankList []*v1.VideoBank
	for c.Next(ctx) {

		// fmt.Printf("//////////////////// pre c /////////////////////////\n\n")
		// fmt.Printf("%+v\n\n", c)

		var elem *v1.VideoBank
		if err = c.Decode(&elem); err != nil {
			logger.Log.Error("readAll: couldn't make video bank item ready for display ", zap.String("reason", err.Error()))
			return nil, status.Errorf(codes.Unknown, "readAll: couldn't make video bank item ready for display '%s'", err.Error())
		}
		// id := res.InsertedID.(primitive.ObjectID).Hex()
		videoBankList = append(videoBankList, elem)
		// fmt.Printf("//////////////////// single elem /////////////////////////\n\n\n")
		// fmt.Printf("%+v\n", elem)
	}
	if err = c.Err(); err != nil {
		logger.Log.Error("readAll: all video bank items couldn't be listed", zap.String("reason", err.Error()))
		return nil, status.Errorf(codes.Unknown, "readAll: all video bank items couldn't be listed '%s'", err.Error())
	}

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		logger.Log.Error("readAll: cannot count all video bank", zap.String("reason", err.Error()))
		return nil, status.Errorf(codes.Unknown, "readAll: cannot count all video bank '%s'", err.Error())
	}
	xTotalData := int32(count)
	xTotalPage := int32(math.Ceil(float64(count) / float64(req.Limit)))

	//shuffle array
	if count > 0 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(videoBankList), func(i, j int) { videoBankList[i], videoBankList[j] = videoBankList[j], videoBankList[i] })
	}

	fmt.Printf("xTotalData: %d, xTotalPage: %d", xTotalData, xTotalPage)

	return &v1.MV_ReadAllResponse{
		Api:       apiVersion,
		TotalData: xTotalData,
		TotalPage: xTotalPage,
		// CurrentPage: xCurrentPage,
		VideoBanks: videoBankList,
	}, nil
}

func (s *videoBankServiceServer) Update(ctx context.Context, req *v1.MV_UpdateRequest) (*v1.MV_UpdateResponse, error) {

	// 	if err := s.checkAPI(req.Api); err != nil {
	// 		return nil, err
	// 	}

	// 	curDB := s.db
	// 	coll := curDB.Collection("jt_video_bank")

	// t := time.Now().In(s.globalLoc)
	// 	ts, err := ptypes.TimestampProto(t)
	// 	if err != nil {
	// 		// log.Fatalf("failed to convert golang Time to protobuf Timestamp: %#v", err)
	// 		logger.Log.Error("failed to convert golang Time to protobuf Timestamp", zap.String("reason", err.Error()))
	// 	}

	// 	//copy map
	// 	up := req.VideoBank
	// 	up.UpdatedAt = ts

	// 	objId, err := primitive.ObjectIDFromHex(req.Id)
	// 	if err != nil {
	// 		// log.Fatalf("Fail converting Hex to ObjectID: %v", err)
	// 		logger.Log.Error("Fail converting Hex to ObjectID", zap.String("reason", err.Error()))
	// 	}
	// 	filter := bson.D{{"_id", objId}}
	// 	updateRec := bson.D{
	// 		{"$set", &up},
	// 		// {"$currentDate", bson.D{{"modifiedAt", true}}},
	// 	}

	// 	// Update data in the collection
	// 	res, err := coll.UpdateOne(ctx, filter, updateRec)
	// 	if err != nil {
	// 		// log.Fatalf("Update data in collection <jt_db.jt_video_bank>: %#v", err)
	// 		logger.Log.Error("Update data in collection <jt_db.jt_video_bank>", zap.String("reason", err.Error()))
	// 	}
	// 	// fmt.Printf("Matched %v documents and updated %v documents.\n", res.MatchedCount, res.ModifiedCount)

	// 	var upId *pmongo.ObjectId
	// 	if res.ModifiedCount > 0 {
	// 		upId = pmongo.NewObjectId(objId)
	// 	}

	//testing
	apiVersion := "1"
	var upId *pmongo.ObjectId

	return &v1.MV_UpdateResponse{
		Api:  apiVersion,
		UpId: upId,
	}, nil
}

func (s *videoBankServiceServer) Delete(ctx context.Context, req *v1.MV_DeleteRequest) (*v1.MV_DeleteResponse, error) {

	// 	if err := s.checkAPI(req.Api); err != nil {
	// 		return nil, err
	// 	}

	// 	curDB := s.db
	// 	coll := curDB.Collection("jt_video_bank")

	// t := time.Now().In(s.globalLoc)
	// 	ts, err := ptypes.TimestampProto(t)
	// 	if err != nil {
	// 		// log.Fatalf("failed to convert golang Time to protobuf Timestamp: %#v", err)
	// 		logger.Log.Error("failed to convert golang Time to protobuf Timestamp", zap.String("reason", err.Error()))
	// 	}

	// 	objId, err := primitive.ObjectIDFromHex(req.Id)
	// 	if err != nil {
	// 		// log.Fatalf("Fail converting Hex to ObjectID: %v", err)
	// 		logger.Log.Error("Fail converting Hex to ObjectID", zap.String("reason", err.Error()))
	// 	}

	// 	filter := bson.D{{"_id", objId}}
	// 	delRec := bson.D{
	// 		{"$set", bson.D{
	// 			{"isactive", false},
	// 			{"updatedat", ts},
	// 			{"updatedby", req.UpdatedBy},
	// 		}},
	// 		// {"$currentDate", bson.D{{"modifiedAt", true}}},
	// 	}

	// 	// Delete data in the collection, change isActive to true
	// 	res, err := coll.UpdateOne(ctx, filter, delRec)
	// 	if err != nil {
	// 		// log.Fatalf("Delete data in collection <jt_db.jt_video_bank>: %#v", err)
	// 		logger.Log.Error("Delete data in collection <jt_db.jt_video_bank>", zap.String("reason", err.Error()))
	// 	}
	// 	// fmt.Printf("Matched %v documents and updated %v documents.\n", res.MatchedCount, res.ModifiedCount)

	// 	var delId *pmongo.ObjectId
	// 	if res.ModifiedCount > 0 {
	// 		delId = pmongo.NewObjectId(objId)
	// 	}
	// 	// fmt.Printf("Delete %s.\n", delId)

	//testing
	apiVersion := "1"
	var delId *pmongo.ObjectId

	return &v1.MV_DeleteResponse{
		Api:   apiVersion,
		DelId: delId,
	}, nil
}

func (s *videoBankServiceServer) GetGenreList(ctx context.Context, req *v1.MV_SetGenreRequest) (*v1.MV_GetGenreResponse, error) {

	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	curDB := s.db
	coll := curDB.Collection("jt_video_bank")

	filter, arrCondition := bson.D{}, bson.A{}
	arrCondition = append(arrCondition, bson.D{{"isactive", true},})
	filter = bson.D{{"$and", arrCondition}}

	genreResp, err := coll.Distinct(ctx, "genre", filter)
	if err != nil {		
		logger.Log.Error("Fail to get genre list from collection <jt_db.jt_video_bank>", zap.String("reason", err.Error()))
		return nil, err
	}

	var genreMap []string
	for _, genreCode := range genreResp {	
		if genreCode != nil {
			genreMap = append(genreMap, v1.MV_GenreOpt_name[genreCode.(int32)])	
		}			
	}

	return &v1.MV_GetGenreResponse{
		Api:   apiVersion,
		Genre: genreMap,
	}, nil
}

func (s *videoBankServiceServer) RefreshCollection(ctx context.Context, req *v1.MV_RefreshRequest) (*v1.MV_RefreshResponse, error) {

	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}
	curDB := s.db
	isSuccess := true

	if req.SourceColName == "" || req.TargetColName == "" {
		logger.Log.Error("'SourceColName' and 'TargetColName' are Required")
		return nil, status.Error(codes.InvalidArgument, "'SourceColName' and 'TargetColName' are Required")
	}

	//check if SourceCol exist
	opts := options.EstimatedDocumentCount().SetMaxTime(1 * time.Second)
	count, err := curDB.Collection(req.SourceColName).EstimatedDocumentCount(ctx, opts)
	if err != nil {
	    logger.Log.Error("ERROR count source collection '"+req.SourceColName, zap.String("reason", err.Error()))
		return nil, err
	}
	fmt.Printf("estimated document count: %v\n", count)
	if count == 0 {
		logger.Log.Error("ERROR count source collection '"+req.SourceColName+ ", maybe the collection doesn't exists.")
		return nil, status.Errorf(codes.NotFound, "ERROR count source collection : '%s'", req.SourceColName)
	}

	//drop TargetColName first
	err = curDB.Collection(req.TargetColName).Drop(ctx)
	if err != nil {
		logger.Log.Error("ERROR MOVIE collection '"+req.TargetColName+ " cannot be dropped.", zap.String("reason", err.Error()))
		return nil, err
	} else {	logger.Log.Info(req.TargetColName+" MOVIE has been dropped.")	}

	result := curDB.RunCommand(ctx, bson.D{
		{"cloneCollectionAsCapped", req.SourceColName},
		{"toCollection", req.TargetColName},
		{"size", 10000000}, //movie
	})
	fmt.Printf("\nresult : <%+v>\n\n", result)

	var document bson.M
	err = result.Decode(&document)
	if err !=nil {
		logger.Log.Error("ERROR clone from '"+req.SourceColName+"' to collection '"+req.TargetColName+ " ~~ ", zap.String("reason", err.Error()))
		return nil, err
	}

	// ## compare two collection size
	resultSize := curDB.RunCommand(ctx, bson.M{"collStats":req.SourceColName})
	var docSize bson.M
	err = resultSize.Decode(&docSize)
	if err !=nil {	
		logger.Log.Error("ERROR counting '"+req.SourceColName+"' ~~ ", zap.String("reason", err.Error()))	
	} else {	fmt.Printf("\n%s size source: %v Bytes\n", req.SourceColName, docSize["size"])		}
	
	resultSize2 := curDB.RunCommand(ctx, bson.M{"collStats":req.TargetColName})
	var docSize2 bson.M
	err = resultSize2.Decode(&docSize2)
	if err !=nil {	
		logger.Log.Error("ERROR counting '"+req.TargetColName+"' ~~ ", zap.String("reason", err.Error()))	
	} else {	fmt.Printf("\n%s size target: %v Bytes\n", req.TargetColName, docSize2["size"])	}
	
	return &v1.MV_RefreshResponse{
		Api:       apiVersion,
		IsSuccess: isSuccess,
	}, nil
}