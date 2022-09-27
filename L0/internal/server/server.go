package server

import (
	"L0/internal/repository"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/nats-io/stan.go"
	"log"
	"net/http"
	"sync"
)

type Server struct {
	ordersCache map[string]repository.Order
	conn        *sqlx.DB
}

var OrdersCache = make(map[string]repository.Order)
var conn *sqlx.DB

func StartAPI() {

	router := gin.Default()

	var err error
	conn, err = repository.PgxConnection()
	if err != nil {
		fmt.Println(err)
		return
	}

	loadCacheFromDB(conn)

	router.GET("/", func(context *gin.Context) {
		context.IndentedJSON(http.StatusOK, "Используйте роутинг для поиска: ./orders/:id")
	})

	router.GET("/orders/:id", sendOrderToApi)

	err = router.Run()
	if err != nil {
		return
	}
}

func loadCacheFromDB(conn *sqlx.DB) {
	ordersLoad, err := repository.GetAllOrders(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, order := range ordersLoad {
		OrdersCache[order.OrderUID] = order
	}
}

func sendOrderToApi(ctx *gin.Context) {
	id := ctx.Param("id")
	if order, status := OrdersCache[id]; status {
		ctx.IndentedJSON(http.StatusOK, order)
	} else {
		ctx.IndentedJSON(http.StatusNotFound, "Данный элемент не найден")
	}
}

func ListenToNutsStreaming() {

	sc, err := stan.Connect("test-cluster", "sub", stan.NatsURL("0.0.0.0:4223"))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()
	sc.Subscribe("test", insertOrder)
	if err != nil {
		fmt.Println(err)
	}

	Block()
}

func Block() {
	w := sync.WaitGroup{}
	w.Add(1)
	w.Wait()
}

func insertOrder(m *stan.Msg) {

	var order repository.Order
	err := json.Unmarshal(m.Data, &order)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(order.OrderUID)
	if _, status := OrdersCache[order.OrderUID]; !status {
		OrdersCache[order.OrderUID] = order
		err = repository.InsertOrderToPgx(order, conn)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Println("Order is already created")
	}

}
