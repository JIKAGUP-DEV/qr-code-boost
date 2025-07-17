package qrcode

import (
	"qr-code-boost/src/middlewares"

	"github.com/gin-gonic/gin"
)

// @Summary      QR Code Routes
func QRCodesRouter(r *gin.Engine, qrCodeController *QRCodeController) {
	qrCodeRoutes := r.Group("/qr", middlewares.InternalOnlyMiddleware())
	{
		qrCodeRoutes.POST("/", qrCodeController.CreateQRCode)
		qrCodeRoutes.GET("/near/:slug", qrCodeController.FindNearScans)
		qrCodeRoutes.GET("/user/:userId", qrCodeController.FindAllQRCodes)
		qrCodeRoutes.GET("/:slug", qrCodeController.AccessQRCode)
	}
}
