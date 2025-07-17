package qrcode

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"qr-code-boost/src/config"
	"qr-code-boost/src/user"

	"qr-code-boost/src/mongo/models"

	"qr-code-boost/src/scan"

	qrcode "github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type QRCodeWithURL struct {
	models.QRCode
	Url string `bson:"url"`
}

func Create(dto CreateQRCodeDto, mongoClient *mongo.Client, postgresClient *sql.DB) (QRCodeWithURL, error) {
	webURL, envErr := config.GetEnvVariable("WEB_URL")

	if envErr != nil {
		return QRCodeWithURL{}, envErr
	}

	user, err := user.FindById(dto.UserId, postgresClient)

	if err != nil {
		fmt.Println("Erro ao buscar usuário.")
		return QRCodeWithURL{}, err
	}

	if user == nil || user.Name == "" {
		fmt.Println("Usuário não encontrado.")
		return QRCodeWithURL{}, errors.New("user not found")
	}

	collection := mongoClient.Database("qr-code-boost").Collection("qrcodes")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // QUANDO TERMINAR DE UTILIZAR O CONTEXTO A CONEXÃO É FECHADA

	buffer, err := generateQRCode(dto.Link)

	if err != nil {
		fmt.Printf("Error writing file: %v", err)
		return QRCodeWithURL{}, nil
	}

	id := primitive.NewObjectID()

	filename := fmt.Sprintf("%s.png", id.Hex())
	filepath := fmt.Sprintf("./static/images/%s", filename)
	errSavingFile := saveStaticFile(buffer, filepath)

	qrCode := models.QRCode{
		ID:   id,
		Slug: dto.Slug,
		Link: dto.Link,
		Location: models.Location{
			Type:        "Point",
			Coordinates: []float64{dto.Long, dto.Lat},
		},
		UserId:    dto.UserId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, errCreating := collection.InsertOne(ctx, qrCode)

	if errCreating != nil {
		fmt.Println("Erro ao inserir QR Code na collection.")
		return QRCodeWithURL{}, errCreating
	}

	if errSavingFile != nil {
		fmt.Printf("Erro ao salvar arquivo estático: %v", errSavingFile)
		collection.DeleteOne(context.Background(), bson.M{"_id": id})
		return QRCodeWithURL{}, nil
	}

	fmt.Printf("Código QR criado com sucesso! ID: %s\n", id.Hex())

	qrCodeCreated := QRCodeWithURL{
		QRCode: qrCode,
		Url:    webURL + "/qr/" + dto.Slug,
	}

	return qrCodeCreated, nil
}

func generateQRCode(link string) ([]byte, error) {
	var buffer []byte
	buffer, err := qrcode.Encode(link, qrcode.Medium, 256)

	if err != nil {
		return nil, err
	}

	return buffer, nil
}

func saveStaticFile(buffer []byte, filepath string) error {
	errSavingFile := os.WriteFile(filepath, buffer, 0644)
	return errSavingFile
}

func FindAll(userId string, client *mongo.Client) ([]QRCodeWithURL, error) {
	webURL, err := config.GetEnvVariable("WEB_URL")

	if err != nil {
		return nil, err
	}

	collection := client.Database("qr-code-boost").Collection("qrcodes")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // QUANDO TERMINAR DE UTILIZAR O CONTEXTO A CONEXÃO É FECHADA

	filter := bson.D{{Key: "userId", Value: userId}}

	cursor, err := collection.Find(ctx, filter)

	if err != nil {
		fmt.Println("Erro ao buscar QR Codes na collection.")
		return nil, err
	}

	defer cursor.Close(ctx)

	var qrCodesWithURL []QRCodeWithURL

	for cursor.Next(ctx) {
		log.Println("Documento encontrado...")

		var qrCode models.QRCode
		var qrCodeWithURL QRCodeWithURL

		err := cursor.Decode(&qrCode)

		if err != nil {
			fmt.Printf("Erro ao buscando QR Codes na collection: %v", err)
			return nil, err
		}

		qrCodeWithURL.QRCode = qrCode

		qrCodeWithURL.Url = webURL + "/qr/" + qrCode.Slug

		qrCodesWithURL = append(qrCodesWithURL, qrCodeWithURL)
	}

	return qrCodesWithURL, nil
}

func FindById(qrCodeId string, client *mongo.Client) (models.QRCode, error) {
	coll := client.Database("qr-code-boost").Collection("qrcodes")

	filter := bson.D{{Key: "_id", Value: qrCodeId}}

	var result models.QRCode
	err := coll.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return models.QRCode{}, err
		}
		panic(err)
	}

	return result, nil
}

func AccessQRCode(slug string, dto CoordinatesDto, client *mongo.Client) (models.QRCode, error) {
	qrCode, err := FindBySlug(slug, client)
	if err != nil {
		fmt.Printf("\n\n [QRCODE SERVICE AccessQRCode] Erro ao encontrar QR Code: %v\n\n", err)
		return models.QRCode{}, err
	}

	fmt.Printf("QR Code encontrado: %s\n", qrCode.Slug)
	_, err = scan.Create(scan.CreateScanDto{
		QRCodeId: qrCode.ID,
		Lat:      dto.Lat,
		Long:     dto.Long,
	}, client)

	if err != nil {
		fmt.Printf("\n\n [QRCODE SERVICE AccessQRCode] Erro ao criar scan: %v\n\n", err)
		return models.QRCode{}, err
	}

	return qrCode, nil
}

func FindBySlug(slug string, client *mongo.Client) (models.QRCode, error) {
	coll := client.Database("qr-code-boost").Collection("qrcodes")

	filter := bson.D{{Key: "slug", Value: slug}}

	var result models.QRCode
	err := coll.FindOne(context.TODO(), filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Printf("\n\n [QRCODE SERVICE FindBySlug] QR Code não encontrado: %s\n\n", slug)
			return models.QRCode{}, err
		}
		panic(err)
	}

	return result, nil
}

func FindNearScans(slug string, maxDistance int64, client *mongo.Client) ([]models.Scan, error) {
	qrCode, err := FindBySlug(slug, client)

	if err != nil {
		fmt.Printf("\n\n [QRCODE SERVICE FindNearScans] Erro ao encontrar QR Code: %v\n\n", err)
		return nil, err
	}

	findNearScansFilterDto := scan.FindNearScansFilterDto{
		MaxDistance: &maxDistance,
	}

	scans, err := scan.FindNearScans(findNearScansFilterDto, qrCode, client)

	if err != nil {
		fmt.Printf("\n\n [QRCODE SERVICE FindNearScans] Erro ao buscar scans próximos: %v\n\n", err)
		return nil, err
	}

	return scans, nil
}
