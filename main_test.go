package main_test

import (
	"fmt"

	"github.com/ewwwwwqm/mig"
)

func ExampleCheckDriver() {
	err := CheckDriver("mysql")
	fmt.Println(err)
	// Output:
	// 
}