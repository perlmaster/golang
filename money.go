package main

import (
    "golang.org/x/text/language"
    "golang.org/x/text/message"
	"fmt"
)
func format_money(pennies int) {
	// var cents int
	var dollars int

	fmt.Printf("Format %d as a currency value\n",pennies)

	cents := pennies % 100
	dollars = pennies / 100
	fmt.Printf("dollars = %d , cents = %d\n",dollars,cents)
		fmt.Printf("%s","$")
	p := message.NewPrinter(language.English)
	p.Printf("%d",dollars)
	fmt.Printf("%s%02d\n",".",cents)

	return
} // format_money

func main() {

	format_money(150075)  // $1,500.75
}