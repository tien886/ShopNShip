package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "up",
			"service": "delivery-service",
		})
	})

    // Scalar documentation
    r.GET("/docs", func(c *gin.Context) {
        c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
            <!doctype html>
            <html>
              <head>
                <title>Delivery Service API Reference</title>
                <meta charset="utf-8" />
                <meta name="viewport" content="width=device-width, initial-scale=1" />
              </head>
              <body>
                <script id="api-reference" data-url="https://cdn.jsdelivr.net/npm/@scalar/galaxy/dist/latest.json"></script>
                <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
              </body>
            </html>
        `))
    })

	r.Run(":8082")
}
