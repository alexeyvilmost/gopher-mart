package server

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress     string
	DBConnection   string
	AccrualAddress string
}

func NewConfig() *Config {
	result := new(Config)
	runAddress := flag.String("a", "localhost:8080", "Server run address")
	accrualAddress := flag.String("r", "http://localhost:9090", "Accrual system address")
	dbConnection := flag.String("d", "port=5432 user=app dbname=shortener password=app host=localhost", "Connection string for DB")

	flag.Parse()
	var ok bool
	if result.RunAddress, ok = os.LookupEnv("RUN_ADDRESS"); !ok {
		result.RunAddress = *runAddress
	}
	if result.AccrualAddress, ok = os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); !ok {
		result.AccrualAddress = *accrualAddress
	}
	if result.DBConnection, ok = os.LookupEnv("DATABASE_URI"); !ok {
		result.DBConnection = *dbConnection
	}
	return result
}
