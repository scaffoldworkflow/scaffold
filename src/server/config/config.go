package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"scaffold/server/constants"
	"strconv"
)

const DEFAULT_CONFIG_PATH = "/home/scaffold/data/config.json"
const ENV_PREFIX = "SCAFFOLD_"

type ConfigObject struct {
	Host              string          `json:"host" env:"HOST"`
	Port              int             `json:"port" env:"PORT"`
	Protocol          string          `json:"protocol" env:"PROTOCOL"`
	WSPort            int             `json:"ws_port" env:"WS_PORT"`
	LogLevel          string          `json:"log_level" env:"LOG_LEVEL"`
	LogFormat         string          `json:"log_format" env:"LOG_FORMAT"`
	BaseURL           string          `json:"base_url" env:"BASE_URL"`
	Admin             UserObject      `json:"admin" env:"ADMIN"`
	DB                DBObject        `json:"db" env:"DB"`
	Node              NodeObject      `json:"node" env:"NODE"`
	HeartbeatInterval int             `json:"heartbeat_interval" env:"HEARTBEAT_INTERVAL"`
	HealthCheckLimit  int             `json:"health_check_limit" env:"HEALTH_CHECK_LIMIT"`
	Reset             ResetObject     `json:"reset" env:"RESET"`
	FileStore         FileStoreObject `json:"file_store" env:"FILE_STORE"`
	TLSEnabled        bool            `json:"tls_enabled" env:"TLS_ENABLED"`
	TLSSkipVerify     bool            `json:"tls_skip_verify" env:"TLS_SKIP_VERIFY"`
	TLSCrtPath        string          `json:"tls_crt_path" env:"TLS_CRT_PATH"`
	TLSKeyPath        string          `json:"tls_key_path" env:"TLS_KEY_PATH"`
}

type FileStoreObject struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Protocol  string `json:"protocol"`
}

type UserObject struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type DBObject struct {
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

func LoadConfig() {
	configPath := os.Getenv(ENV_PREFIX + "CONFIG_PATH")
	if configPath == "" {
		configPath = DEFAULT_CONFIG_PATH
	}

	Config = ConfigObject{
		Host:              "scaffold",
		Port:              2997,
		Protocol:          "http",
		WSPort:            8080,
		LogLevel:          constants.LOG_LEVEL_INFO,
		LogFormat:         constants.LOG_FORMAT_CONSOLE,
		BaseURL:           "http://localhost:2997",
		HeartbeatInterval: 500,
		HealthCheckLimit:  10,
		TLSEnabled:        false,
		TLSSkipVerify:     false,
		TLSCrtPath:        "/tmp/cert.crt",
		TLSKeyPath:        "/tmp/cert.key",
		Admin: UserObject{
			Username: "admin",
			Password: "admin",
		},
		DB: DBObject{
			Username: "myCoolMongoDBUsername",
			Password: "myCoolMongoDBPassword",
			Name:     "scaffold",
			Host:     "mongodb",
			Port:     27017,
		},
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
		},
	}

	jsonFile, err := os.Open(configPath)
	if err == nil {
		log.Printf("Successfully Opened %v", configPath)

		byteValue, _ := ioutil.ReadAll(jsonFile)

		json.Unmarshal(byteValue, &Config)
	}

	v := reflect.ValueOf(Config)
	t := reflect.TypeOf(Config)

	for i := 0; i < v.NumField(); i++ {
		field, found := t.FieldByName(v.Type().Field(i).Name)
		if !found {
			continue
		}

		value := field.Tag.Get("env")
		if value != "" {
			val, present := os.LookupEnv(ENV_PREFIX + value)
			if present {
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

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
}

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
