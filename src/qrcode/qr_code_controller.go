package qrcode

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

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

type CoordinatesDto struct {
	Lat  *float64 `json:"lat"`
	Long *float64 `json:"long"`
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

	latitudeStr := c.GetHeader("X-User-Latitude")
	longitudeStr := c.GetHeader("X-User-Longitude")

	var latitude float64
	var longitude float64

	if latitudeStr != "" {
		latFloat, err := strconv.ParseFloat(latitudeStr, 64)
		if err == nil {
			latitude = latFloat
		}
	}
	if longitudeStr != "" {
		longFloat, err := strconv.ParseFloat(longitudeStr, 64)
		if err == nil {
			longitude = longFloat
		}
	}

	coordinates := CoordinatesDto{
		Lat:  &latitude,
		Long: &longitude,
	}

	qrCode, err := AccessQRCode(slug, coordinates, u.MongoClient)

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
		fmt.Printf("Corpo da requisição inválido | %v", err)
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
// @Param 			 maxDistance query int false "Maximum distance in meters (default: 3000 or 3km)"
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

	var maxDistance int64 = 3000

	maxDistanceQuery := c.Query("maxDistance")
	if maxDistanceQuery != "" {
		parsedDistance, errQuery := strconv.ParseInt(maxDistanceQuery, 10, 64)

		if errQuery != nil {
			fmt.Printf("Erro ao converter maxDistance: %v", errQuery)
			c.IndentedJSON(400, gin.H{
				"message": "Parâmetro maxDistance inválido.",
				"status":  400,
			})
			return
		}

		maxDistance = parsedDistance
	}

	fmt.Printf("\n Buscando scans próximos para o QR Code com slug: %s\n", slug)
	scans, err := FindNearScans(slug, maxDistance, u.MongoClient)

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
