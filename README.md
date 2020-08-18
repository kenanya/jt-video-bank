## Description
This is Video Bank Service. This service provide basic CRUD function and some others function.
The data is stored to MongoDB. gRPC is used here for the communication.

Stack Tech: Golang, MongoDB, gRPC, docker


## Clone/Download Repo
git clone https://github.com/kenanya/jt-video-bank.git

## Configuration
The config file was stored at jt-video-bank/pkg/config/configGlobal.yaml. You can change the values according to your config. I set the grpc port to 9603, so when you start this service, it will run on port 9603.

## How to Start
cd jt-video-bank/cmd/server<br/>
go build .<br/>
APP_ENV=local ./server

## Consume Service
We can use <a href="https://appimage.github.io/BloomRPC/">BloomRPC</a> to test consuming this service. After you download and install the BloomRPC, you have to import the protobuf file at jt-video-bank/api/proto/v1/video_bank_service.proto. As the default, you will get the initial random value as sample request when the protobuf file has been imported. You can use the initial value or yours to test the service. 

### Sample Create Request
{
  "api": "v1",
  "videoBank": {
    "contentid" : "103",
    "idx" : "GFX_103",
    "provider" : "GENFLIX",
    "providershort" : "GFX",
    "providerlabel" : "https://sinarmas-vod.s3-ap-southeast-1.amazonaws.com/poster/movies/103/Unleashed1280.jpg",
    "title" : "Unleashed",
    "tags": [
      "Hello"
    ],
    "videoType": 0,
    "genre": [
      0
    ],
   "year" : 0,
    "duration" : "02:20:00",
    "synopsis" : "Danny adalah seorang petarung kuat yang dibesarkan bagaikan seekor anjing oleh pemiliknya, seorang gangster. Danny adalah predator dia akan bertarung dan membunuh siapapun sesuai instruksi majikannya. Pikiran dan kepribadiannya seperti anak kecil dan dia tak pernah menjalani kehidupan normal. Dalam suatu kejadian, Danny terluka parah dan koma, lalu dia dirawat oleh orang-orang baik. Akankah dia bisa berubah menjadi normal ataukah naluri buasnya hidup lagi?",
    "cast" : [ 
        "Jet Li", 
        " Bob Hoskins", 
        " Morgan Freeman"
    ],
    "playerurl" : "https://old.genflix.co.id/smartfren/player/movies/unleashed",
    "poster" : {
        "s" : "https://sinarmas-vod.s3-ap-southeast-1.amazonaws.com/poster/movies/103/Unleashed300.jpg",
        "m" : "https://sinarmas-vod.s3-ap-southeast-1.amazonaws.com/poster/movies/103/Unleashed400.jpg",
        "l" : "https://sinarmas-vod.s3-ap-southeast-1.amazonaws.com/poster/movies/103/Unleashed1280.jpg",
        "xxx_nounkeyedliteral" : {},
        "xxx_unrecognized" : null,
        "xxx_sizecache" : 0
    },
    "director" : [ 
        "a"
    ],
    "contenttype" : 1,
    "availability" : "",
    "contentas" : "",
    "contentlevel" : 0,
    "isactive" : true,
    "isvalid" : true,
    "createdAt": {
      "seconds": 20,
      "nanos": 10
    },
    "expiredDate": {
      "seconds": 20,
      "nanos": 10
    }
  }
}

### Sample Read Request
{
  "api": "v1",
  "idx": "GFX_103"
}
