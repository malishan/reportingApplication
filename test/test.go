package main

import (
	"fmt"
	"os"
)

func main() {
	err := os.Setenv("FOO", "1")
	if err != nil {
		fmt.Println(err)
	}
	val := os.Getenv("CassandraHost")

	fmt.Println("Value : " + val)
}
