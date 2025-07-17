package main

import (
	"encoding/json"
	"fmt"
	"log"
	"qr-code-boost/src/mongo"
	"qr-code-boost/src/postgres"
	"qr-code-boost/src/qrcode"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"

	docs "qr-code-boost/docs"

	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {
	// Carrega as variáveis de .env para o ambiente
	err := godotenv.Load()
	if err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}
}

// @title           QR Code Boost API
// @version         1.0
// @description     This is an API for managing users and QR codes.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /
func main() {
	router := gin.Default()
	docs.SwaggerInfo.BasePath = "/"

	postgresClient, postgresConnectionErr := postgres.ConnectionPostgres()

	if postgresConnectionErr != nil {
		log.Fatal("Erro ao se conectar com o postgresql: ", postgresConnectionErr)
	}

	mongoClient, mongoConnectionErr := mongo.ConnectMongoDB()

	if mongoConnectionErr != nil {
		log.Fatal("Erro ao se conectar com o mongodb: ", mongoConnectionErr)
	}

	mongo.CreateIndexes(mongoClient)

	router.Static("/images", "./static/images")

	stats, _ := mongo.GetDBStats(mongoClient)

	json, errJson := json.MarshalIndent(stats, "", " ")
	if errJson != nil {
		fmt.Println("Erro ao identar os stats")
	}

	fmt.Println("---------------------------")
	fmt.Println(string(json))
	fmt.Println("---------------------------")

	qrCodeController := &qrcode.QRCodeController{
		MongoClient:    mongoClient,
		PostgresClient: postgresClient,
	}

	qrcode.QRCodesRouter(router, qrCodeController)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	fmt.Printf("\n %sAPI running on http://localhost:8080 \n", "\x1b[32m")
	fmt.Printf("\n Docs available on http://localhost:8080/swagger/index.html%s \n \n", "\x1b[0m")
	router.Run(":8080")

}
