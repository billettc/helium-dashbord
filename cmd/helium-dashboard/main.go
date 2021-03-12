package main

import (
	"fmt"
	"os"

	"github.com/billettc/helium-dashbord/dashboard"
)

func main() {

	addresses := os.Args[1:]

	d := dashboard.NewDashboard(addresses)
	if err := d.Run(); err != nil {
		panic(err)
	}
	fmt.Println("Goodbye")
}
