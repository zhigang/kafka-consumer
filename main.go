package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zprojects/kafka-consumer/config"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/jinzhu/configor"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	llog "github.com/labstack/gommon/log"
)

var (
	globalConfig config.Config
	startedAt    time.Time
)

func main() {
	startedAt = time.Now()

	loadConfig()
	useProfiler()

	e := echo.New()
	s := &http.Server{
		Addr:         globalConfig.Service.Address,
		ReadTimeout:  20 * time.Minute,
		WriteTimeout: 20 * time.Minute,
	}

	initEchoServer(e)
	bindingAPI(e)

	e.Logger.Fatal(e.StartServer(s))
}

func loadConfig() {
	fmt.Println("Config loading...")
	configor.Load(&globalConfig, "config/config.yml")
	fmt.Printf("Use config: %+v\n", globalConfig)
}

func useProfiler() {
	if globalConfig.Profiler.Enable {
		fmt.Printf("Enable profiler on %s\n", globalConfig.Profiler.Listening)
		go func() {
			log.Println(http.ListenAndServe(globalConfig.Profiler.Listening, nil))
		}()
	}
}

func initEchoServer(e *echo.Echo) {

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Logger.SetLevel(llog.INFO)

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.GET("/info", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"name":      globalConfig.Service.Name,
			"version":   globalConfig.Service.Version,
			"startedAt": startedAt,
			"contacts":  globalConfig.Contacts,
		})
	})
}

func bindingAPI(e *echo.Echo) {
	apiV1 := e.Group("/v1")
	apiV1.GET("/consumer", consumer)
}

func consumer(c echo.Context) error {
	topics := c.QueryParam("topics")
	if topics == "" {
		return c.JSON(http.StatusBadRequest, "'topics' is required parameter.")
	}

	initial := c.QueryParam("initial")
	if initial == "" {
		return c.JSON(http.StatusBadRequest, "'initial' is required parameter. set 'new' or 'old'.")
	}

	countStr := c.QueryParam("count")
	if countStr == "" {
		return c.JSON(http.StatusBadRequest, "'count' is required parameter. Type is 'int'.")
	}

	count, err := strconv.Atoi(countStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("'count' type is 'int'. %v", err))
	}

	group := c.QueryParam("group")
	if group == "" {
		group = fmt.Sprintf("%s-%d", globalConfig.Kafka.Group, time.Now().Unix())
	}

	commit := c.QueryParam("commit")

	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.CommitInterval = 1 * time.Second

	if initial == "old" {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		config.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	consumer, err := cluster.NewConsumer(strings.Split(globalConfig.Kafka.Brokers, ","), group, strings.Split(topics, ","), config)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("Unable to connect to kafka brokers (consumer). %v", err))
	}

	defer consumer.Close()

	go func() {
		for err := range consumer.Errors() {
			c.Logger().Error(err)
		}
	}()

	go func() {
		for note := range consumer.Notifications() {
			c.Logger().Info(fmt.Sprintf("Rebalanced: %+v", note))
		}
	}()

	idx := 0
	var msgs []Msg
	for m := range consumer.Messages() {
		var msg Msg
		msg.Consumer = group
		msg.Topic = m.Topic
		msg.Partition = m.Partition
		msg.Offset = m.Offset
		msg.Value = string(m.Value)
		msgs = append(msgs, msg)
		c.Logger().Info(fmt.Sprintf("%s/%d/%d : %s", msg.Topic, msg.Partition, msg.Offset, msg.Value))
		if commit == "1" {
			consumer.MarkOffset(m, "") //MarkOffset 并不是实时写入kafka，有可能在程序crash时丢掉未提交的offset
		}
		idx++
		if idx >= count {
			break
		}
	}

	return c.JSON(http.StatusOK, msgs)
}
