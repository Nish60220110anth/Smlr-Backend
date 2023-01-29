package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath("./")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}
}

func GetPort() int32 {
	return viper.GetInt32("PORT")
}

//Only init and GetFilePath doesn't need error Loggers

func GetFilePath(fileName string) (filePath string) {
	myabspath, err := filepath.Abs("./")

	if err != nil {
		log.Fatalln(err.Error())
	}

	return filepath.Join(myabspath, fileName)
}

func CheckError(err error, eLog *log.Logger) {
	if err != nil {
		eLog.Fatalln(err.Error())
	}
}

func GetServerUpTime() int32 {
	return viper.GetInt32("SERVER_UP_TIME")
}

func GetErrorLogger() *log.Logger {
	fileHandle, err := os.OpenFile(GetFilePath(viper.GetString("ERROR_LOG_FILE")),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalln(err)
	}

	return log.New(fileHandle, fmt.Sprintf("PREFIX: "), log.Ltime|log.Lshortfile)
}

func GetLogger() *log.Logger {
	fileHandle, err := os.OpenFile(GetFilePath(viper.GetString("LOG_FILE")),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalln(err)
	}

	return log.New(fileHandle, fmt.Sprintf("PREFIX: "), log.Ltime|log.Lshortfile)
}
