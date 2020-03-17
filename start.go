package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	yaml "github.com/goccy/go-yaml"
	//yaml "gopkg.in/yaml.v3"
)

const userPath = "assets/user.yaml"
const ladderPath = "assets/ladder.yaml"

type User struct {
	Id int `yaml:"id"`
	Hash string `yaml:"hash"`
	Level int `yaml:"level"`
}

type UserList struct {
	Users []User `yaml:"users"`
}

type Ladder struct {
	Id int `yaml:"id"`
	Name string `yaml:"name"`
	Url string `yaml:"url"`
	Level int `yaml:"level"`
}

type LadderList struct {
	Ladders []Ladder `yaml:"ladders"`
}

type ProxyConfig struct {
	encodeConfig string
	rawConfig string
	userLevel int
	userList UserList
	ladderList LadderList
}

func (s ProxyConfig) readFile(filePath string) []byte {
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("read fail", err)
	}

	return f
}

func (s *ProxyConfig) getUserLevel(userId string) {
	userRead := s.readFile(userPath)
	//fmt.Println(string(userRead))
	if err := yaml.Unmarshal(userRead, &(s.userList)); err != nil {
		fmt.Println("Failed to resolve user list.", err)
	}
	//fmt.Println(s.userList)

	flag := 0
	for _, user := range s.userList.Users{
		//fmt.Println(user.Hash)
		if strings.Compare(userId, user.Hash) == 0 {
			flag = 1
			s.userLevel = user.Level
		}
	}
	if flag == 0 {
		s.userLevel = -1
	}
}

func (s *ProxyConfig) getLadderList() {
	ladderRead := s.readFile(ladderPath)
	//fmt.Println(string(ladderRead))
	if err := yaml.Unmarshal(ladderRead, &(s.ladderList)); err != nil {
		fmt.Println("Failed to resolve ladder list.", err)
	}
	//fmt.Println(s.ladderList)
}

func (s *ProxyConfig) levelSelect(length string) {
	//fmt.Println(s.userLevel)
	ladders := s.ladderList.Ladders

	var buffer bytes.Buffer
	count := 0
	for _, ladder := range ladders{
		if ladder.Level >= s.userLevel {
			buffer.WriteString(ladder.Url)
			buffer.WriteString("\n")
		}
		count += 1
		if strings.Compare(length, strconv.Itoa(count)) == 0 {
			break
		}
	}
	s.rawConfig = buffer.String()
}

func (s *ProxyConfig) base64Encode() string {
	s.encodeConfig = base64.StdEncoding.EncodeToString([]byte(s.rawConfig))
	return s.encodeConfig
}

func main() {
	r := gin.Default()
	r.Static("/assets", "./assets")

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK,"Welcome! Hope the world peace.")
	})

	r.GET("/v1/typeV/:identityString", func(c *gin.Context) {
		identityString := c.Param("identityString")
		length := c.Query("length")

		v := &ProxyConfig{}
		v.getUserLevel(identityString)
		if (*v).userLevel == -1 {
			c.JSON(http.StatusOK, gin.H{
				"status": "refused",
				"id": identityString,
				"length": length,
			})
			return
		}
		v.getLadderList()
		v.levelSelect(length)
		ec := v.base64Encode()
		fmt.Println(ec)

		filename := "typeV_" + identityString + ".txt"
		c.Header("Content-Disposition","attachment; filename=" + filename)
		c.Data(http.StatusOK, "application/octet-stream; charset=utf-8", []byte(ec))
	})

	r.Run(":11001") // 监听并在 0.0.0.0:10001 上启动服务
}
