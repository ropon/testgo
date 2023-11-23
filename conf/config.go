package conf

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ropon/logger"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const SERVERNAME = "work_api"

var (
	etcdHost       string
	configFileName string
	Cfg            Config
)

type LogCfg struct {
	Level     string
	FilePath  string
	FileName  string
	MaxSize   int64
	SplitFlag bool
	TimeDr    float64
}

type MysqlCfg struct {
	MysqlConn            string
	MysqlConnectPoolSize int
}

type RedisCfg struct {
	RedisConn   string
	RedisPasswd string
	RedisDb     int
}

// Config 配置文件结构体
type Config struct {
	LogCfg        LogCfg
	MysqlCfg      map[string]MysqlCfg
	RedisCfg      map[string]RedisCfg
	External      map[string]string
	ExternalInt64 map[string]int64
	Listen        string
}

type EtcdRes struct {
	Node struct {
		Key           string `json:"key"`
		Value         string `json:"value"`
		ModifiedIndex int    `json:"modifiedIndex"`
		CreatedIndex  int    `json:"createdIndex"`
	} `json:"node"`
}

func initConf() (err error) {
	flag.StringVar(&configFileName, "c", configFileName, "config file")
	flag.StringVar(&etcdHost, "etcd", etcdHost, "etcd addr")
	flag.Parse()
	if configFileName != "" {
		err = initConfigFile(configFileName, &Cfg)
	} else {
		err = loadCfgFromEtcd([]string{etcdHost}, SERVERNAME, &Cfg)
	}
	if err != nil {
		return
	}
	logCfg := logger.LogCfg(Cfg.LogCfg)
	err = logger.InitLog(&logCfg)
	return
}

func initConfigFile(filename string, cfg *Config) error {
	_, err := os.Stat(filename)
	if err != nil {
		return err
	}
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, cfg); err != nil {
		err = fmt.Errorf("unmarshal error :%s", string(bytes))
		return err
	}
	return nil
}

func loadCfgFromEtcd(addrs []string, service string, cfg interface{}) error {
	environment := strings.ToLower(os.Getenv("GOENV"))
	if environment == "" {
		environment = "online"
	}
	data, err := cfgFromEtcd(addrs, service, environment)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), cfg)
}

func cfgFromEtcd(addrs []string, service, env string) (string, error) {
	if len(addrs) == 0 || addrs[0] == "" {
		return "", fmt.Errorf("etcd地址不能为空")
	}
	addr := fmt.Sprintf("%s/v2/keys%s", addrs[0], etcdKey(service, env))
	resp, err := http.Get(addr)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", err
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	etcdRes := EtcdRes{}
	err = json.Unmarshal(data, &etcdRes)
	if err != nil {
		log.Printf("unmarshal [%s] failed:%v", data, err)
		return "", err
	}
	return etcdRes.Node.Value, nil
}

func etcdKey(service, env string) string {
	return fmt.Sprintf("/config/%s/%s", service, env)
}

// Init 初始化配置文件、Mysql、Redis、etcd等
func Init() (err error) {
	err = initConf()
	if err != nil {
		return
	}
	return nil
}
