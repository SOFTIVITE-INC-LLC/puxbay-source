package main

import (
	"fmt"
	"time"
)

func main() {
	d, _ := time.ParseDuration("168h")
	fmt.Println("168h in seconds:", int(d.Seconds()))
}
