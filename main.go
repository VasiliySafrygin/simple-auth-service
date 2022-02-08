package main

import (
	"auth-service/api"
	"auth-service/db"
	"auth-service/function"
	"auth-service/rpc"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/go-zookeeper/zk"

	"github.com/getsentry/sentry-go"
)

type Service struct {
	db *sql.DB
}

type RegistryData struct {
	Api     string    `json:"api"`
	Rpc     string    `json:"rpc"`
	Created time.Time `json:"created"`
}

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		headersParts := strings.Split(authHeader, " ")
		if len(headersParts) == 2 {
			userId, err := function.VerifyToken(headersParts[1])
			if err == nil {
				c.Set("userId", userId)
			}
		}
		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func SetZkVars(data []byte, key string, host string) *zk.Conn {
	c, _, err := zk.Connect([]string{host}, time.Second*100)
	if err != nil {
		fmt.Println("Connect error", err)
		panic(err)
	}

	var flags int32 = 0
	var acls = zk.WorldACL(zk.PermAll)

	// create
	exist, s, err := c.Exists(key)
	if err != nil {
		fmt.Println(err)
		return c
	}
	fmt.Println("exist:", exist)
	if exist {
		s, err = c.Set(key, data, s.Version)
		if err != nil {
			fmt.Println(err)
			return c
		}
		fmt.Println("Set:", s)
	} else {
		p, err_create := c.Create(key, data, flags, acls)
		if err_create != nil {
			fmt.Println("Create error", err_create)
			return c
		}
		fmt.Println("created:", p)
	}
	return c
}

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:     "http://2d70d6da024f409a9b84f21d0194039f@192.168.49.166/4",
		Release: "auth-service@0.0.1",
		Debug:   true,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	defer sentry.Flush(2 * time.Second)

	var (
		DbHost     = os.Getenv("DB_HOST")
		DbPort, _  = strconv.ParseInt(os.Getenv("DB_PORT"), 10, 64)
		DbUser     = os.Getenv("DB_USER")
		DbPassword = os.Getenv("DB_PASSWORD")
		DbName     = os.Getenv("DB_NAME")
		ZkHost     = os.Getenv("ZK_HOST")
		ServiceApi = os.Getenv("SERVICE_API")
		ServiceKey = os.Getenv("SERVICE_KEY")
		ServiceRpc = os.Getenv("SERVICE_RPC")
	)
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		DbHost, DbPort, DbUser, DbPassword, DbName)

	registryData := RegistryData{}
	registryData.Api = fmt.Sprintf("http://%s/api/v1/", ServiceApi)
	registryData.Rpc = ServiceRpc
	registryData.Created = time.Now().UTC()
	data, err := json.Marshal(registryData)

	if err != nil {
		panic(err)
	}

	c := SetZkVars(data, ServiceKey, ZkHost)
	defer c.Close()
	defer func(c *zk.Conn, path string, version int32) {
		err := c.Delete(path, version)
		if err != nil {
			fmt.Printf("Error delete %s %v\n", path, err)
		}
	}(c, ServiceKey, -1)

	router := gin.Default()
	db.Connect(psqlInfo)
	rpc.StartGRPCServer(ServiceRpc)

	router.Use(CORSMiddleware())
	router.GET("/about/", api.About)
	router.GET("/api/v1/user/", api.User)
	router.POST("/api/v1/create/", api.Create)
	router.POST("/api/v1/login/", api.Login)
	router.POST("/api/v1/logout/", api.Logout)

	fmt.Println("Ready")
	log.Fatal(router.Run(ServiceApi))
}
