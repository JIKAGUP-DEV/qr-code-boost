package mongo

import (
	"context"
	"fmt"
	"qr-code-boost/config"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoAlbum struct {
	ID     string `bson:"id,omitempty"`
	Title  string `bson:"title"`
	Artist string `bson:"artist"`
	Price  string `bson:"price"`
}

type DBStats struct {
	DatabaseName          string  `json:"databaseName" bson:"db"`
	DatabaseDataSize      float64 `json:"databaseDataSize" bson:"dataSize"`
	DatabaseStorageSize   float64 `json:"databaseStorageSize" bson:"storageSize"`
	CollectionName        string  `json:"collectionName" bson:"-"`
	CollectionDataSize    int32   `json:"collectionDataSize" bson:"size"`
	CollectionStorageSize int32   `json:"collectionStorageSize" bson:"storageSize"`
	DocumentCount         int32   `json:"documentCount" bson:"count"`
}

func CreateIndexes(client *mongo.Client) {
	qrCodesCollection := client.Database("qr-code-boost").Collection("qrcodes")

	qrCodeIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "location", Value: "2dsphere"}},
			Options: nil,
		},
	}

	_, errQRCodes := qrCodesCollection.Indexes().CreateMany(context.Background(), qrCodeIndexes)

	if errQRCodes != nil {
		fmt.Printf("Erro ao criar índices para 'qrcodes': %v\n", errQRCodes)
	} else {
		fmt.Println("Índices da coleção 'qrcodes' verificados/criados.")
	}

	scansCollection := client.Database("qr-code-boost").Collection("scans")

	scanLocationIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	}

	_, errScans := scansCollection.Indexes().CreateOne(context.Background(), scanLocationIndex)
	if errScans != nil {
		fmt.Printf("Erro ao criar índice para 'scans': %v\n", errScans)
	} else {
		fmt.Println("Índice da coleção 'scans' verificado/criado.")
	}
}

func ConnectMongoDB() (*mongo.Client, error) {
	databaseURL, err := config.GetEnvVariable("MONGODB_URL")

	if err != nil {
		return nil, err
	}

	clientOptions := options.Client().ApplyURI(databaseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("falha ao verificar conexão: %v", err)
	}

	fmt.Println("Conexão com MongoDB estabelecida!")
	return client, nil
}

func GetDBStats(client *mongo.Client) (DBStats, error) {
	ctx := context.Background()

	var result DBStats
	var dbStats bson.M
	var collStats bson.M

	dbCommand := bson.D{{Key: "dbStats", Value: 1}, {Key: "scale", Value: 1024}}
	collCommand := bson.D{{Key: "collStats", Value: "qrcodes"}, {Key: "scale", Value: 1024}}

	errDBCommand := client.Database("qr-code-boost").RunCommand(ctx, dbCommand).Decode(&dbStats)
	errCollCommand := client.Database("qr-code-boost").RunCommand(ctx, collCommand).Decode(&collStats)

	if errDBCommand != nil {
		fmt.Printf("Erro ao pegar estatísticas do banco Mongo DB: %v", errDBCommand)
		return DBStats{}, errDBCommand
	}
	if errCollCommand != nil {
		fmt.Printf("Erro ao pegar estatísticas do banco Mongo DB: %v", errCollCommand)
		return DBStats{}, errCollCommand
	}

	fmt.Println("Type size", reflect.TypeOf(collStats["size"]))
	fmt.Println(collStats["size"])

	if dataSize, ok := dbStats["dataSize"].(float64); ok {
		result.DatabaseDataSize = dataSize
	}
	if storageSize, ok := dbStats["storageSize"].(float64); ok {
		result.DatabaseStorageSize = storageSize
	}
	if size, ok := collStats["size"].(int32); ok {
		result.CollectionDataSize = size
	}
	if storageSize, ok := collStats["storageSize"].(int32); ok {
		result.CollectionStorageSize = storageSize
	}
	if count, ok := collStats["count"].(int32); ok {
		result.DocumentCount = count
	}

	result.DatabaseName = "qr-code-boost"
	result.CollectionName = "qrcodes"

	return result, nil
}
