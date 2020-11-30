package internal

import (
	"fmt"
	"os"
)

func ExitOnError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
