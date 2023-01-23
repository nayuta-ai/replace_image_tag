package test

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	fmt.Println("load env file for test")
	// When running GitHub Actions, the below line must not run.
	// err := godotenv.Load("../.env")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	sample := os.Getenv("TEST_VARIABLE")
	if sample != "test" {
		fmt.Printf("assertion error: Got %s, but expected test", sample)
		os.Exit(1)
	}
	fmt.Println("start test")
	code := m.Run()
	os.Exit(code)
}
