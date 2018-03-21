package cli

import (
	"fmt"
	"time"
)

func InputLoop() {
	for {
		time.Sleep(5 * time.Second)
		fmt.Println("Looping...")
	}
}
