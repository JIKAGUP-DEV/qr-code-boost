package qrcode

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type CreateQRCodeDto struct {
	Slug   string  `json:"slug" binding:"required,min=2,max=20"`
	Link   string  `json:"link" binding:"required,url"`
	Lat    float64 `json:"lat" binding:"required,latitude"`
	Long   float64 `json:"long" binding:"required,longitude"`
	UserId string  `json:"userId" binding:"required,uuid"`
}

type QRCodeController struct {
	MongoClient    *mongo.Client
	PostgresClient *sql.DB
}

// @Summary      Find QR Code by Slug
// @Tags         QR Codes
// @Accept       json
// @Produce      json
// @Param        slug path string true "QR Code Slug"
// @Success      200 {object} qrcode.QRCodeWithURL
// @Router       /qr/{slug} [get]
func (u *QRCodeController) AccessQRCode(c *gin.Context) {
	slug := c.Param("slug")

	if slug == "" {
		c.IndentedJSON(400, gin.H{
			"message": "Slug não pode ser vazio.",
			"status":  400,
		})
		return
	}

	qrCode, err := AccessQRCode(slug, u.MongoClient)

	if err != nil {
		fmt.Printf("Erro ao buscar QR Code: %v", err)
		c.IndentedJSON(500, gin.H{
			"message": "Erro ao buscar QR Code",
			"error":   err.Error(),
		})
		return
	}

	c.IndentedJSON(200, qrCode)
}

// @Summary      List QR Codes from a specific user
// @Tags         QR Codes
// @Accept       json
// @Produce      json
// @Param        userId   path      string  true  "User ID"
// @Success      200  {array}   qrcode.QRCodeWithURL
// @Router       /qr/user/{userId} [get]
func (u *QRCodeController) FindAllQRCodes(c *gin.Context) {
	os.MkdirAll("./static/images", os.ModePerm)

	userId := c.Param("userId")

	if userId == "" {
		c.IndentedJSON(400, gin.H{
			"message": "User ID não pode ser vazio.",
			"status":  400,
		})
		return
	}

	qrCodes, err := FindAll(userId, u.MongoClient)

	if err != nil {
		fmt.Printf("Erro ao listar QR Codes: %v", err)
		c.IndentedJSON(500, err)
		return
	}

	c.IndentedJSON(200, qrCodes)
}

// @Summary      Create a QR Code
// @Tags         QR Codes
// @Accept       json
// @Produce      json
// @Param        request body qrcode.CreateQRCodeDto true "QR Code Payload"
// @Success      201  {object}  qrcode.QRCodeWithURL
// @Router       /qr [post]
func (u *QRCodeController) CreateQRCode(c *gin.Context) {
	var createQRCodeDto CreateQRCodeDto

	err := c.ShouldBindJSON(&createQRCodeDto)

	if err != nil {
		fmt.Println("Corpo da requisição inválido")
		c.IndentedJSON(400, gin.H{
			"message": "Corpo da requisição inválido.",
			"status":  400,
		})
		return
	}

	qrCodeWithURL, errCreating := Create(createQRCodeDto, u.MongoClient, u.PostgresClient)

	if errCreating != nil {
		fmt.Printf("Erro ao criar QR Code: %v", errCreating)
		c.IndentedJSON(500, gin.H{
			"message": "Erro ao criar QR Code",
			"error":   errCreating.Error(),
		})
		return
	}

	c.IndentedJSON(200, qrCodeWithURL)
}

// @Summary      Find scans near a QR Code
// @Tags         QR Codes
// @Accept       json
// @Produce      json
// @Param        slug path string true "QR Code Slug"
// @Success      200 {array} models.Scan
// @Router       /qr/near/{slug} [get]
func (u *QRCodeController) FindNearScans(c *gin.Context) {
	slug := c.Param("slug")

	if slug == "" {
		c.IndentedJSON(400, gin.H{
			"message": "Slug não pode ser vazio.",
			"status":  400,
		})
		return
	}

	fmt.Printf("\n Buscando scans próximos para o QR Code com slug: %s\n", slug)
	scans, err := FindNearScans(slug, u.MongoClient)

	if err != nil {
		fmt.Printf("Erro ao buscar scans próximos: %v", err)
		c.IndentedJSON(500, gin.H{
			"message": "Erro ao buscar scans próximos",
			"error":   err.Error(),
		})
		return
	}

	c.IndentedJSON(200, scans)
}
