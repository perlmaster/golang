//////////////////////////////////////////////////////////////////////
//
// File      : go-envdump.go
//
// Author    : Barry Kimelman
//
// Created   : November 29, 2019
//
// Purpose   : Go CGI script to display environment
//
// Notes     :
//
//////////////////////////////////////////////////////////////////////

package main

import (
    "fmt"
    "os"
	"strings"
	"bufio"
)

var fields_count int
var fields_map map[string]string

//////////////////////////////////////////////////////////////////////
//
// Function  : parse_fields
//
// Purpose   : Parse all the input fields for the CGI script
//
// Inputs    : (none)
//
// Output    : (none)
//
// Returns   : nothing
//
// Example   : parse_fields()
//
// Notes     : Check ENV{"QUERY_STRING"} to determine GET or POST
//
///////////////////////////////////////////////////////////////////////

func parse_fields() {
	fields_count = 0
	fields_map = make(map[string]string)  // initialize the map
	value , exists := os.LookupEnv("QUERY_STRING")
	post_method := true
	if exists {
		if value == "" {
			// fmt.Println("QUERY_STRING is empty")
		} else {
			// fmt.Println("Value of QUERY_STRING => " + value)
			post_method = false
		}
	} else {
		fmt.Println("QUERY_STRING is not an environment variable")
	}
	if post_method {
		// fmt.Println("<BR>Read POST method data<BR>")
		reader := bufio.NewReader(os.Stdin)
		for {
			looptext, _ := reader.ReadString('\n')
			count := len(looptext)
			if count == 0 {
				break
			}
			// fmt.Print("Without newline [" + looptext + "]\n")
			post_vars := strings.Split(looptext, "&")
			for _, pvar := range post_vars {
				pair := strings.SplitN(pvar, "=", 2)
				// fmt.Println("<BR>POST Var : name = " + pair[0] + " , value = " + pair[1])
				fields_count += 1
				fields_map[pair[0]] = pair[1]
			}
		}

	} else {
		// fmt.Println("<BR>Parse QUERY_STRING fields<BR>")
		query_vars := strings.Split(value, "&")
		for _, qvar := range query_vars {
			pair := strings.SplitN(qvar, "=", 2)
			// fmt.Println("<BR>Query Var : name = " + pair[0] + " , value = " + pair[1])
			fields_map[pair[0]] = pair[1]
			fields_count += 1
		}
	}

} // parse_fields

//////////////////////////////////////////////////////////////////////
//
// Function  : display_input_fields
//
// Purpose   : Display input fields passed to CGI script
//
// Inputs    : (none)
//
// Output    : input fields
//
// Returns   : nothing
//
// Example   : display_input_fields()
//
// Notes     : (none)
//
///////////////////////////////////////////////////////////////////////

func display_input_fields() {
	fmt.Println("<BR>")
	fmt.Printf("<H3>The list of %d fields</H3>\n",fields_count)
	fmt.Println("<BR>")

	fmt.Println("<TABLE border='1' cellspacing='0' cellpadding='3'>")
	fmt.Println("<THEAD>")
	fmt.Println("<TR><TD>Name</TD><TD>Value</TD></TR>")
	fmt.Println("</THEAD>")
	fmt.Println("<TBODY style='font-size: 12px;'>")
	for key, value := range fields_map {
        fmt.Println("<TR><TD>" + key + "</TD><TD>" + value + "</TD></TR>")
	}

	fmt.Println("</TBODY>")
	fmt.Println("</TABLE>")
	fmt.Println("<BR><BR>")

} // display_input_fields

//////////////////////////////////////////////////////////////////////
//
// Function  : display_env_vars
//
// Purpose   : Display environment variables passed to CGI script
//
// Inputs    : (none)
//
// Output    : input fields
//
// Returns   : nothing
//
// Example   : display_env_vars()
//
// Notes     : (none)
//
///////////////////////////////////////////////////////////////////////

func display_env_vars() {
	fmt.Println("<H2>Environment</H2>")

	fmt.Println("<TABLE border='1' cellspacing='0' cellpadding='3'>")
	fmt.Println("<THEAD>")
	fmt.Println("<TR><TD>Name</TD><TD>Value</TD></TR>")
	fmt.Println("</THEAD>")
	fmt.Println("<TBODY style='font-size: 12px;'>")
    for _, e := range os.Environ() {
        pair := strings.SplitN(e, "=", 2)
        fmt.Println("<TR><TD>" + pair[0] + "</TD><TD>" + pair[1] + "</TD></TR>")
    }

	fmt.Println("</TBODY>")
	fmt.Println("</TABLE>")
	fmt.Println("<BR><BR>")

} // display_env_vars

//////////////////////////////////////////////////////////////////////
//
// Function  : main
//
// Purpose   : Display CGI environment
//
// Inputs    : (none)
//
// Output    : env vars and field values
//
// Returns   : nothing
//
// Example   : main()
//
// Notes     : (none)
//
///////////////////////////////////////////////////////////////////////

func main() {

    fmt.Println("Content-Type: text/html\n");
    fmt.Println("<HTML>")
    fmt.Println("<HEAD>")
	fmt.Println("<TITLE>Environment</TITLE>")
	fmt.Println("</HEAD>")
	fmt.Println("<BODY>")
	display_env_vars()

	parse_fields()
	display_input_fields()

	fmt.Println("</BODY>")
	fmt.Println("</HTML>")
}