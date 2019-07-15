package main

import "fmt"

func processNotifications(updates []*UpdatedServer) {
	fmt.Println("changes detected!")
	for _, update := range updates {
		fmt.Println("Server:", update.Server)
		fmt.Println("Key:", update.Key)
		fmt.Println("From:", update.From)
		fmt.Println("To:", update.To)
	}
}
