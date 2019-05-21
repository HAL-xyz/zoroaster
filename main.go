package main

import (
	"fmt"
	"log"
	"zoroaster/aws"
)

func main() {

	aws.InitDB()

	triggers, err := aws.LoadTriggersFromDB()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("I got %d triggers\n", len(triggers))
	for _, v := range triggers {
		fmt.Println(v.TriggerType)
		fmt.Println(v.Filters)
	}
}
