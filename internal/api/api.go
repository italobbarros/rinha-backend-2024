// api.go
package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	postgres "github.com/italobbarros/rinha-backend-2024/internal/db"
	"github.com/italobbarros/rinha-backend-2024/internal/models"
	_ "github.com/lib/pq"
)

func NewApi(db *sql.DB) *Api {

	clientes := &Clientes{
		Sync: make(map[int]chan struct{}),
	}
	postgres.Init(db)
	return &Api{
		Clientes: clientes,
		db:       db,
		sync: Sync{
			semaphore: map[int]chan struct{}{
				1: make(chan struct{}, 1),
				2: make(chan struct{}, 1),
				3: make(chan struct{}, 1),
				4: make(chan struct{}, 1),
				5: make(chan struct{}, 1),
			},
		},
	}
}

func (a *Api) Acquire(id int) {
	a.sync.mutex.Lock()
	sync, ok := a.sync.semaphore[id]
	a.sync.mutex.Unlock()
	if !ok {
		return
	}
	sync <- struct{}{}

}

func (a *Api) Release(id int) {
	a.sync.mutex.Lock()
	sync, ok := a.sync.semaphore[id]
	a.sync.mutex.Unlock()
	if !ok {
		return
	}
	<-sync
}

// cadastrarTransacao cadastra uma nova transação para um cliente específico.
func (a *Api) cadastrarTransacao(c *gin.Context) {
	clienteIDStr := c.Param("id")
	clienteID, err := strconv.Atoi(clienteIDStr)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "ID do cliente não é um numero"})
		return
	}

	if clienteID > 5 || clienteID < 1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID do cliente não existe"})
		return
	}
	var transacao models.PostTransacaoRequest
	if err := c.ShouldBindJSON(&transacao); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	if transacao.Tipo != "c" && transacao.Tipo != "d" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "tipo da transação diferente de \"c\" e \"d\""})
		return
	}
	length := len(transacao.Descricao)
	if length < 1 || length > 10 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "descrição possui tamanho menor do que 1 ou maior do que 10"})
		return
	}
	a.Acquire(clienteID)
	defer a.Release(clienteID)
	var result models.PostTransacaoResponseSuccess
	if transacao.Tipo == "c" { //credito
		result, err = postgres.UpdateCreditTransationClient(a.db, clienteID, &transacao)
		if err != nil {
			fmt.Println("UpdateCreditTransationClient clienteId:", clienteID, err.Error())
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Error "})
			return
		}
	} else { //debito
		result, err = postgres.UpdateDebitTransationClient(a.db, clienteID, &transacao)
		if err != nil {
			fmt.Println("UpdateDebitTransationClient clienteId:", clienteID, err.Error())
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Erro ao exec a transação -" + err.Error()})
			return
		}

	}
	c.JSON(http.StatusOK, gin.H{
		"limite": result.Limite,
		"saldo":  result.Saldo,
	})
}

func (a *Api) getExtrato(c *gin.Context) {
	clienteIDStr := c.Param("id")
	clienteID, err := strconv.Atoi(clienteIDStr)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "ID do cliente não é um numero"})
		return
	}
	if clienteID > 5 || clienteID < 1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "ID do cliente não existe"})
		return
	}
	r, err := postgres.GetValueAndHist(a.db, clienteID)
	if err != nil {
		fmt.Println("ERROR GetValueAndHist error:", err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err})
		return
	}
	r.Saldo.Data = time.Now()

	c.JSON(http.StatusOK, r)
}

func (a *Api) Run() {
	defer postgres.Close()
	router := gin.Default()
	router.Use(corsHandler) // Adicionar o middleware CORS

	router.GET("/clientes/:id/extrato", a.getExtrato)
	router.POST("/clientes/:id/transacoes", a.cadastrarTransacao)

	router.Run(os.Getenv("API_SERVER_LISTEN")) //os.Getenv("API_SERVER_LISTEN")
}

func corsHandler(c *gin.Context) {
	// Configurar cabeçalhos CORS
	c.Header("Access-Control-Allow-Origin", "*") // Permitir qualquer origem
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
	c.Header("Content-Type", "application/json")

	if c.Request.Method == http.MethodOptions {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
}
