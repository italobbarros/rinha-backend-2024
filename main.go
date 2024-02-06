package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Criação de uma instância do Gin
	router := gin.Default()

	// Definição de uma rota simples
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Bem-vindo à API Gin!"})
	})

	// Inicia o servidor na porta 8080
	router.Run(os.Getenv("API_SERVER_LISTEN"))
}
