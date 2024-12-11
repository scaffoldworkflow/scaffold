package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"reflect"
	"strconv"

	"github.com/jfcarter2358/go-logger"
)

const DEFAULT_CONFIG_PATH = "/home/scaffold/data/config.json"
const ENV_PREFIX = "SCAFFOLD_"

type ConfigObject struct {
	Host                     string          `json:"host"`
	Port                     int             `json:"port"`
	Protocol                 string          `json:"protocol"`
	WSPort                   int             `json:"ws_port" env:"WS_PORT"`
	LogLevel                 string          `json:"log_level" env:"LOG_LEVEL"`
	LogFormat                string          `json:"log_format" env:"LOG_FORMAT"`
	BaseURL                  string          `json:"base_url" env:"BASE_URL"`
	PodmanOpts               string          `json:"podman_opts" env:"PODMAN_OPTS"`
	Admin                    UserObject      `json:"admin" env:"ADMIN"`
	DBConnectionString       string          `json:"db_connection_string" env:"DB_CONNECTION_STRING"`
	DB                       DBObject        `json:"db"`
	Node                     NodeObject      `json:"node" env:"NODE"`
	HeartbeatInterval        int             `json:"heartbeat_interval" env:"HEARTBEAT_INTERVAL"`
	HeartbeatBackoff         int             `json:"heartbeat_backoff" env:"HEARTBEAT_BACKOFF"`
	Reset                    ResetObject     `json:"reset" env:"RESET"`
	FileStore                FileStoreObject `json:"file_store" env:"FILESTORE"`
	TLSEnabled               bool            `json:"tls_enabled" env:"TLS_ENABLED"`
	TLSSkipVerify            bool            `json:"tls_skip_verify" env:"TLS_SKIP_VERIFY"`
	TLSCrtPath               string          `json:"tls_crt_path" env:"TLS_CRT_PATH"`
	TLSKeyPath               string          `json:"tls_key_path" env:"TLS_KEY_PATH"`
	RabbitMQConnectionString string          `json:"rabbitmq_connection_string" env:"RABBITMQ_CONNECTION_STRING"`
	ManagerQueueName         string          `json:"manager_queue_name" env:"MANAGER_QUEUE_NAME"`
	WorkerQueueName          string          `json:"worker_queue_name" env:"WORKER_QUEUE_NAME"`
	KillQueueName            string          `json:"kill_queue_name" env:"KILL_QUEUE_NAME"`
	PingHealthyThreshold     int             `json:"ping_healthy_threshold" env:"PING_HEALTHY_THRESHOLD"`
	PingUnknownThreshold     int             `json:"ping_unknown_threshold" env:"PING_UNKNOWN_THRESHOLD"`
	PingDownThreshold        int             `json:"ping_down_threshold" env:"PING_DOWN_THRESHOLD"`
	CheckInterval            int             `json:"check_interval" env:"CHECK_INTERVAL"`
	RestartPeriod            int             `json:"restart_period" env:"RESTART_PERIOD"`
	RunPruneCron             string          `json:"run_prune_cron" env:"RUN_PRUNE_CRON"`
	RunPruneDuration         int             `json:"run_prune_duration" env:"RUN_PRUNE_DURATION"`
}

type FileStoreObject struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Protocol  string `json:"protocol"`
	Type      string `json:"type"`
}

type UserObject struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type DBObject struct {
	Protocol string `json:"protocol"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

type ResetObject struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Host     string `json:"mail_host"`
	Port     int    `json:"mail_port"`
}

type NodeObject struct {
	Type            string `json:"type"`
	ManagerHost     string `json:"manager_host"`
	ManagerPort     int    `json:"manager_port"`
	JoinKey         string `json:"join_key"`
	PrimaryKey      string `json:"primary_key"`
	ManagerProtocol string `json:"manager_protocol"`
}

var Config ConfigObject

// Load configuration from either a local JSON file or via ENV variables
// ENV variables will override settings in the JSON file
func LoadConfig() {
	configPath := os.Getenv(ENV_PREFIX + "CONFIG_PATH")
	if configPath == "" {
		configPath = DEFAULT_CONFIG_PATH
	}

	// Default configuration
	Config = ConfigObject{
		Host:              "",
		Port:              -1,
		Protocol:          "",
		BaseURL:           "http://localhost:2997",
		WSPort:            8080,
		LogLevel:          logger.LOG_LEVEL_INFO,
		LogFormat:         logger.LOG_FORMAT_CONSOLE,
		HeartbeatInterval: 1000,
		HeartbeatBackoff:  10,
		TLSEnabled:        false,
		TLSSkipVerify:     false,
		TLSCrtPath:        "/tmp/certs/cert.crt",
		TLSKeyPath:        "/tmp/certs/cert.key",
		PodmanOpts:        "--security-opt label=disabled --network=host",
		Admin: UserObject{
			Username: "admin",
			Password: "admin",
		},
		DBConnectionString: "mongodb://MyCoolMongoDBUsername:MyCoolMongoDBPassword@mongodb:27017/scaffold",
		DB:                 DBObject{},
		Reset: ResetObject{
			Email:    "",
			Password: "",
			Host:     "smtp.gmail.com",
			Port:     587,
		},
		Node: NodeObject{
			Type:            "manager",
			ManagerHost:     "scaffold-manager",
			ManagerPort:     2997,
			ManagerProtocol: "http",
			JoinKey:         "MyCoolJoinKey12345",
			PrimaryKey:      "MyCoolPrimaryKey12345",
		},
		FileStore: FileStoreObject{
			AccessKey: "MyCoolMinIOAccessKey",
			SecretKey: "MyCoolMinIOSecretKey",
			Host:      "minio",
			Port:      9000,
			Bucket:    "scaffold",
			Region:    "default-region",
			Protocol:  "http",
			Type:      "s3",
		},
		RabbitMQConnectionString: "amqp://guest:guest@rabbitmq:5672",
		ManagerQueueName:         "scaffold_manager",
		WorkerQueueName:          "scaffold_worker",
		KillQueueName:            "scaffold_kill",
		PingHealthyThreshold:     3,
		PingUnknownThreshold:     6,
		PingDownThreshold:        9,
		CheckInterval:            2000,
		RestartPeriod:            86400,         // 24 hours
		RunPruneCron:             "0 0 * * * *", // every day at midnight
		RunPruneDuration:         24,            // 24 hour run lifetime
	}

	// Load JSON if exists
	jsonFile, err := os.Open(configPath)
	if err == nil {
		log.Printf("Successfully Opened %v", configPath)

		byteValue, _ := ioutil.ReadAll(jsonFile)

		json.Unmarshal(byteValue, &Config)
	}

	v := reflect.ValueOf(Config)
	t := reflect.TypeOf(Config)

	// Go through config object fields and check for ENV variable existence/populate
	//   configuration with the present values
	for i := 0; i < v.NumField(); i++ {
		field, found := t.FieldByName(v.Type().Field(i).Name)
		if !found {
			continue
		}

		value := field.Tag.Get("env")
		if value != "" {
			val, present := os.LookupEnv(ENV_PREFIX + value)
			if present {
				// log.Printf("Found ENV var %s with value %s", ENV_PREFIX+value, val)
				w := reflect.ValueOf(&Config).Elem().FieldByName(t.Field(i).Name)
				x := getAttr(&Config, t.Field(i).Name).Kind().String()
				if w.IsValid() {
					switch x {
					case "int", "int64":
						i, err := strconv.ParseInt(val, 10, 64)
						if err == nil {
							w.SetInt(i)
						}
					case "int8":
						i, err := strconv.ParseInt(val, 10, 8)
						if err == nil {
							w.SetInt(i)
						}
					case "int16":
						i, err := strconv.ParseInt(val, 10, 16)
						if err == nil {
							w.SetInt(i)
						}
					case "int32":
						i, err := strconv.ParseInt(val, 10, 32)
						if err == nil {
							w.SetInt(i)
						}
					case "string":
						w.SetString(val)
					case "float32":
						i, err := strconv.ParseFloat(val, 32)
						if err == nil {
							w.SetFloat(i)
						}
					case "float", "float64":
						i, err := strconv.ParseFloat(val, 64)
						if err == nil {
							w.SetFloat(i)
						}
					case "bool":
						i, err := strconv.ParseBool(val)
						if err == nil {
							w.SetBool(i)
						}
					default:
						objValue := reflect.New(field.Type)
						objInterface := objValue.Interface()
						err := json.Unmarshal([]byte(val), objInterface)
						obj := reflect.ValueOf(objInterface)
						if err == nil {
							w.Set(reflect.Indirect(obj).Convert(field.Type))
						} else {
							log.Println(err)
						}
					}
				}
			}
		}
	}

	defer jsonFile.Close()

	breakConfigFields()
}

// get object field by name
func getAttr(obj interface{}, fieldName string) reflect.Value {
	pointToStruct := reflect.ValueOf(obj) // addressable
	curStruct := pointToStruct.Elem()
	if curStruct.Kind() != reflect.Struct {
		panic("not struct")
	}
	curField := curStruct.FieldByName(fieldName) // type: reflect.Value
	if !curField.IsValid() {
		panic("not found:" + fieldName)
	}
	return curField
}

// Break the DB config and base URL into separate fields for ease of use later
func breakConfigFields() {
	baseURL, err := url.Parse(Config.BaseURL)
	if err != nil {
		panic(err)
	}
	baseHost, basePortString, err := net.SplitHostPort(baseURL.Host)
	if err != nil {
		panic(err)
	}
	basePort, err := strconv.Atoi(basePortString)
	if err != nil {
		panic(err)
	}
	baseProtocol := baseURL.Scheme

	Config.Host = baseHost
	Config.Port = basePort
	Config.Protocol = baseProtocol

	mongoURL, err := url.Parse(Config.DBConnectionString)
	if err != nil {
		panic(err)
	}

	mongoHost, mongoPortString, err := net.SplitHostPort(baseURL.Host)
	if err != nil {
		panic(err)
	}

	mongoPort, err := strconv.Atoi(mongoPortString)
	if err != nil {
		panic(err)
	}
	mongoProtocol := baseURL.Scheme

	mongoUsername := mongoURL.User.Username()
	mongoPassword, isSet := mongoURL.User.Password()
	if !isSet {
		panic(errors.New("credentials not provided to DB connection string"))
	}

	mongoName := mongoURL.Path[1:len(mongoURL.Path)]

	Config.DB.Host = mongoHost
	Config.DB.Port = mongoPort
	Config.DB.Username = mongoUsername
	Config.DB.Password = mongoPassword
	Config.DB.Name = mongoName
	Config.DB.Protocol = mongoProtocol
}
