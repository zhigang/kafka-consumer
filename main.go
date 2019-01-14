package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	globalConfig  config.Config
	asyncProducer sarama.AsyncProducer
	startedAt     time.Time
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

	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	initKafkaProducer()
	defer closeProducer()

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
	apiV1.GET("/producer", producer)
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

	filter := c.QueryParam("filter")

	commit := c.QueryParam("commit")

	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	config.Consumer.Offsets.CommitInterval = 1 * time.Second

	// 是否启用SSL
	if globalConfig.Kafka.SSL.Enable {
		tlsConfig, err := newTLSConfig("cert/client.cer.pem", "cert/client.key.pem", "cert/server.cer.pem")
		if err != nil {
			fmt.Printf("Unable new TLS config. %v \n", err)
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

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

		if filter == "" {
			msgs = append(msgs, msg)
			idx++
		} else if strings.Contains(msg.Value, filter) {
			msgs = append(msgs, msg)
			idx++
		}

		c.Logger().Info(fmt.Sprintf("%s/%d/%d : %s", msg.Topic, msg.Partition, msg.Offset, msg.Value))
		if commit == "1" {
			consumer.MarkOffset(m, "") //MarkOffset 并不是实时写入kafka，有可能在程序crash时丢掉未提交的offset
		}
		if idx >= count {
			break
		}
	}

	return c.JSON(http.StatusOK, msgs)
}

func initKafkaProducer() {

	config := sarama.NewConfig()
	config.ClientID = "kafka-test-client"
	//等待服务器所有副本都保存成功后的响应
	config.Producer.RequiredAcks = sarama.WaitForAll
	//随机向partition发送消息
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	//是否等待成功和失败后的响应,只有上面的RequireAcks设置不是NoReponse这里才有用.
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	//设置使用的kafka版本,如果低于V0_10_0_0版本,消息中的timestrap没有作用.需要消费和生产同时配置
	//注意，版本设置不对的话，kafka会返回很奇怪的错误，并且无法成功发送消息
	// config.Version = sarama.V1_0_0_0
	// 是否启用SSL
	if globalConfig.Kafka.SSL.Enable {
		tlsConfig, err := newTLSConfig("cert/client.cer.pem", "cert/client.key.pem", "cert/server.cer.pem")
		if err != nil {
			fmt.Printf("Unable new TLS config. %v \n", err)
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	asyncProducer, err := sarama.NewAsyncProducer(strings.Split(globalConfig.Kafka.Brokers, ","), config)
	if err != nil {
		fmt.Printf("Unable to connect to kafka brokers (producer). %v \n", err)
	}

	//循环判断哪个通道发送过来数据.
	go func(p sarama.AsyncProducer) {
		for {
			select {
			case suc := <-p.Successes():
				if suc != nil {
					// 格式：\033[显示方式;前景色;背景色m
					// 说明：
					// 前景色            背景色           颜色
					// ---------------------------------------
					// 30                40              黑色
					// 31                41              红色
					// 32                42              绿色
					// 33                43              黃色
					// 34                44              蓝色
					// 35                45              紫红色
					// 36                46              青蓝色
					// 37                47              白色
					// 显示方式           意义
					// -------------------------
					// 0                终端默认设置
					// 1                高亮显示
					// 4                使用下划线
					// 5                闪烁
					// 7                反白显示
					// 8                不可见
					fmt.Printf("%s level: \x1b[1;32m%s\x1b[0m, topic: %s, partitions: %d, offset: %d, metadata: %v\n", time.Now().Format("2006-01-02 15:04:05"), "SUCCESS", suc.Topic, suc.Partition, suc.Offset, suc.Metadata)
				}
			case fail := <-p.Errors():
				if fail != nil {
					fmt.Printf("%s level: \x1b[1;31m%s\x1b[0m, topic: %s, partitions: %d, offset: %d, metadata: %v, error: %v\n", time.Now().Format("2006-01-02 15:04:05"), "ERROR", fail.Msg.Topic, fail.Msg.Partition, fail.Msg.Offset, fail.Msg.Metadata, fail.Err)
				}
			}
		}
	}(asyncProducer)
}

func closeProducer() {
	asyncProducer.AsyncClose()
}

func producer(c echo.Context) error {

	countStr := c.QueryParam("cnt")
	if countStr == "" {
		return c.JSON(http.StatusBadRequest, "'cnt' is required parameter. Type is 'int'.")
	}

	cnt, err := strconv.Atoi(countStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, fmt.Sprintf("'cnt' type is 'int'. %v", err))
	}

	for i := 0; i < cnt; i++ {
		time.Sleep(50 * time.Millisecond)

		log := &Test{Level: "INFO", Log: "this is a auto message " + strconv.Itoa(i) + ". " + time.Now().Format("2006-01-02 15:04:05")}
		body, _ := json.Marshal(log)
		// 发送的消息,主题。
		// 注意：这里的msg必须得是新构建的变量，不然你会发现发送过去的消息内容都是一样的，因为批次发送消息的关系。
		msg := &sarama.ProducerMessage{
			Topic: "test",
			Value: sarama.ByteEncoder(body),
		}
		if asyncProducer != nil {
			//使用通道发送
			asyncProducer.Input() <- msg
		} else {
			c.Logger().Info("asyncProducer is nil.")
		}

	}

	return c.JSON(http.StatusOK, "ok")
}

func newTLSConfig(clientCertFile, clientKeyFile, caCertFile string) (*tls.Config, error) {
	tlsConfig := tls.Config{}
	// Load client cert
	cert, err := tls.LoadX509KeyPair(clientCertFile, clientKeyFile)
	if err != nil {
		return &tlsConfig, err
	}

	tlsConfig.Certificates = []tls.Certificate{cert}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return &tlsConfig, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig.RootCAs = caCertPool

	tlsConfig.BuildNameToCertificate()
	return &tlsConfig, err
}
