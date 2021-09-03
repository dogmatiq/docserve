package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.Getenv("GITHUB_USER_TOKEN"))
}
