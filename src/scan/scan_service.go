package scan

import (
	"context"
	"fmt"
	"qr-code-boost/src/mongo/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateScanDto struct {
	QRCodeId primitive.ObjectID `bson:"qrCodeId"`
	Lat      *float64           `bson:"lat"`
	Long     *float64           `bson:"long"`
}

type FindNearScansFilterDto struct {
	MaxDistance *int64
}

func Create(dto CreateScanDto, client *mongo.Client) (models.Scan, error) {
	fmt.Printf("Criando scan para QR Code ID: %s\n", dto.QRCodeId.Hex())

	_, err := FindQRCodeById(dto.QRCodeId, client)

	if err != nil {
		fmt.Printf("Erro ao encontrar QR Code: %v\n", err)
		return models.Scan{}, err
	}

	coll := client.Database("qr-code-boost").Collection("scans")

	newScan := models.Scan{
		QRCodeId: dto.QRCodeId,
		Location: models.Location{
			Type:        "Point",
			Coordinates: []float64{*dto.Long, *dto.Lat},
		},
		ScanedAt: time.Now(),
	}

	result, err := coll.InsertOne(context.TODO(), newScan)

	if err != nil {
		fmt.Printf("Erro ao inserir scan: %v\n", err)
		panic(err)
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		newScan.ID = oid
	}

	fmt.Printf("Scan criado com sucesso: %s\n", newScan.ID.Hex())

	return newScan, nil
}

func FindQRCodeById(qrCodeId primitive.ObjectID, client *mongo.Client) (models.QRCode, error) {
	coll := client.Database("qr-code-boost").Collection("qrcodes")

	fmt.Println("Buscando QR Code por ID:", qrCodeId)

	filter := bson.D{{Key: "_id", Value: qrCodeId}}

	var result models.QRCode
	err := coll.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Printf("QR Code não encontrado: %s\n", qrCodeId)
			return models.QRCode{}, err
		}
		panic(err)
	}

	return result, nil
}

func FindNearScans(filterDto FindNearScansFilterDto, qrCode models.QRCode, client *mongo.Client) ([]models.Scan, error) {
	coll := client.Database("qr-code-boost").Collection("scans")

	centerPoint := qrCode.Location

	var defaultDistance int64 = 3000
	maxDistance := defaultDistance
	if filterDto.MaxDistance != nil {
		maxDistance = *filterDto.MaxDistance
	}

	filter := bson.D{
		{
			Key: "location",
			Value: bson.D{
				{
					// 2. Usando o operador "$nearSphere" para GeoJSON
					Key: "$nearSphere",
					Value: bson.D{
						{Key: "$geometry", Value: centerPoint},
						{Key: "$maxDistance", Value: maxDistance}, // 3km
					},
				},
			},
		},
		{Key: "qrCodeId", Value: qrCode.ID},
	}

	var nearbyScans []models.Scan
	cursor, err := coll.Find(context.TODO(), filter)
	if err != nil {
		fmt.Printf("[SCAN SERVICE] Erro ao executar a busca de scans próximos: %v\n", err)
		return nil, err
	}

	if err = cursor.All(context.TODO(), &nearbyScans); err != nil {
		return nil, err
	}

	fmt.Printf("Encontrados %d scans próximos.\n", len(nearbyScans))
	return nearbyScans, nil
}
