package main

import (
    "golang.org/x/text/language"
    "golang.org/x/text/message"
	"fmt"
)
func format_money(pennies int) string {

	cents := pennies % 100
	dollars := pennies / 100
	p := message.NewPrinter(language.English)
	
	dol_value := p.Sprintf("%d",dollars)
	result := fmt.Sprintf("$%s%s%02d",dol_value,".",cents)

	return(result)
} // format_money

func main() {

	value := 150075
	result := format_money(value)  // $1,500.75
	fmt.Printf("\nThe formatted money value of %d = %s\n",value,result)
}
