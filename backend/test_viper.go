package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()

	d := viper.GetDuration("JWT_ACCESS_EXPIRY")
	fmt.Println("JWT_ACCESS_EXPIRY duration:", d)
	fmt.Println("Seconds:", int(d.Seconds()))
}
