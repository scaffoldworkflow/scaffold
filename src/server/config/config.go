package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strconv"
)

const DEFAULT_CONFIG_PATH = "/home/scaffold/data/config.json"
const ENV_PREFIX = "SCAFFOLD_"

type ConfigObject struct {
	HTTPHost          string          `json:"http_host" env:"HTTP_HOST"`
	HTTPPort          int             `json:"http_port" env:"HTTP_PORT"`
	BaseURL           string          `json:"base_url" env:"BASE_URL"`
	Admin             UserObject      `json:"admin" env:"ADMIN"`
	DB                DBObject        `json:"db" env:"DB"`
	Node              NodeObject      `json:"node" env:"NODE"`
	HeartbeatInterval int             `json:"heartbeat_interval" env:"HEARTBEAT_INTERVAL"`
	Reset             ResetObject     `json:"reset" env:"RESET"`
	FileStore         FileStoreObject `json:"file_store" env:"file_store"`
}

type FileStoreObject struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
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
	Type        string `json:"type"`
	ManagerHost string `json:"manager_host"`
	ManagerPort int    `json:"manager_port"`
	JoinKey     string `json:"join_key"`
	PrimaryKey  string `json:"primary_key"`
}

var Config ConfigObject

func LoadConfig() {
	configPath := os.Getenv(ENV_PREFIX + "CONFIG_PATH")
	if configPath == "" {
		configPath = DEFAULT_CONFIG_PATH
	}

	Config = ConfigObject{
		HTTPHost:          "scaffold",
		HTTPPort:          2997,
		BaseURL:           "http://localhost:2997",
		HeartbeatInterval: 500,
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
			Type:        "manager",
			ManagerHost: "scaffold-manager",
			ManagerPort: 2997,
			JoinKey:     "MyCoolJoinKey12345",
			PrimaryKey:  "MyCoolPrimaryKey12345",
		},
		FileStore: FileStoreObject{
			AccessKey: "MyCoolMinIOAccessKey",
			SecretKey: "MyCoolMinIOSecretKey",
			Host:      "minio",
			Port:      9000,
			Bucket:    "scaffold",
			Region:    "default-region",
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
