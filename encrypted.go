//////////////////////////////////////////////////////////////////////
//
// File      : encrypted.go
//
// Author    : Barry Kimelman
//
// Created   : December 25, 2019
//
// Purpose   : Go CGI script to manage my encrypted mysql data table
//
// Notes     :
//
//////////////////////////////////////////////////////////////////////

package main

import (
    "fmt"
	"time"
    "os"
	"strings"
	"bufio"
	"database/sql"
	"path/filepath"
	"regexp"
	"net/url"
	"strconv"

   _ "github.com/go-sql-driver/mysql"

)

var script_name string
var top_level_href string

var fields_count int
var fields_map map[string]string
var function string
var debug_mode bool = false
var request_method string
var request_uri string
var post_data string
var encrypted_data_table = "my_encrypted"
var encrypted_data_control_table = "my_encrypted_control"
var title string = "Encrypted Data Management"
var dummy_test_string = "DUMMY TEST STRING"

var mysql_connect_string string = "myusername:mypassword@(127.0.0.1:3306)/mydatabase?parseTime=true"

type EncryptedRec struct {
	id				int
	modified_date	string
	encrypted_data	string  // decrypted data
}
var encrypted_rec EncryptedRec
var encrypted_records []EncryptedRec
var num_encrypted_records int
var encrypted_map map[int]int  // key is "id" , value is index into history[]
var onmouseover string

//////////////////////////////////////////////////////////////////////
//
// Function  : debug_print
//
// Purpose   : Generate a DEBUG message if debugging mode is active
//
// Inputs    : message string - the message to be displayed
//
// Output    : (none)
//
// Returns   : nothing
//
// Example   : debug_print()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func debug_print(message string) {

	if debug_mode {
		fmt.Printf("<H3>DEBUG : %s</H3>\n",message)
	}

	return
} // debug_print

//////////////////////////////////////////////////////////////////////
//
// Function  : format_now
//
// Purpose   : Format the current time and date into a string
//
// Inputs    : (none)
//
// Output    : (none)
//
// Returns   : formatted time/date string
//
// Example   : now := format_now()
//
// Notes     : string format will be YYYY-MM-DD HH:MM:SS
//
///////////////////////////////////////////////////////////////////////

func format_now() string {
	var now string

	t := time.Now()
	now = fmt.Sprintf("%d:%d:%d\n",t.Hour(),t.Minute(),t.Second())
	now += " " + fmt.Sprintf("%04d-%02d-%02d\n",t.Year(),t.Month(),t.Day())

	return(now)
} // format_now

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
	var value string
	var post_method bool

	fields_count = 0
	fields_map = make(map[string]string)  // initialize the map
	request_method , _ = os.LookupEnv("REQUEST_METHOD")
	if request_method == "POST" {
		post_method = true
	} else {
		post_method = false
	}
	request_uri , _ = os.LookupEnv("REQUEST_URI")

	post_data = ""
	if post_method {
		reader := bufio.NewReader(os.Stdin)
		for {
			looptext, _ := reader.ReadString('\n')
			looptext , _ = url.QueryUnescape(looptext)
			count := len(looptext)
			if count == 0 {
				break
			}
			post_data += looptext + "\n"
			post_vars := strings.Split(looptext, "&")
			for _, pvar := range post_vars {
				pair := strings.SplitN(pvar, "=", 2)
				fields_count += 1
				fields_map[pair[0]] = pair[1]
			}
		} // FOR over lines of POST data
		debug_print("POST method data is<BR><PRE>" + post_data + "</PRE>")
	} else {
		value , _ = os.LookupEnv("QUERY_STRING")
		if value == "" {
			// fmt.Printf("<H3>DEBUG : query string is empty</H3>\n")
		} else {
			value , _ = url.QueryUnescape(value)
			debug_print("GET method data is<BR><PRE>" + value + "</PRE>")
			query_vars := strings.Split(value, "&")
			for _, qvar := range query_vars {
				pair := strings.SplitN(qvar, "=", 2)
				fields_map[pair[0]] = pair[1]
				fields_count += 1
			} // FOR over GET input fields
		}
	}

	return
} // parse_fields

//////////////////////////////////////////////////////////////////////
//
// Function  : database_error
//
// Purpose   : display a database error message and terminate execution
//
// Inputs    : format_string string - format string for fmt.Printf
//             err error - error info object
//
// Output    : (none)
//
// Returns   : nothing
//
// Example   : database_error("<H2>database error. failed to connect<BR>%q</H2>",err)
//
// Notes     : Execution is terminated
//
///////////////////////////////////////////////////////////////////////

func database_error(format_string string, err error) {
	fmt.Printf(format_string,err)
	fmt.Println("</BODY></HTML>")
	os.Exit(0)
} // database_error

//////////////////////////////////////////////////////////////////////
//
// Function  : connect_to_database
//
// Purpose   : Connect to the database
//
// Inputs    : 
//
// Output    : (none)
//
// Returns   : database pointer/handle
//
// Example   : db = connect_to_database()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func connect_to_database() *sql.DB {
	var db *sql.DB
	var err1 error

	db, err1 = sql.Open("mysql", mysql_connect_string)
    if err1 != nil {
		database_error("<H2>database error. failed to connect<BR>%q</H2>",err1)
    }
    if err1 = db.Ping(); err1 != nil {
		database_error("<H2>database error. failed to ping<BR>%q</H2>",err1)
    }

	return db
} // connect_to_database

//////////////////////////////////////////////////////////////////////
//
// Function  : display_parameters
//
// Purpose   : Display input parameters passed to CGI script
//
// Inputs    : (none)
//
// Output    : input fields
//
// Returns   : nothing
//
// Example   : display_parameters()
//
// Notes     : (none)
//
///////////////////////////////////////////////////////////////////////

func display_parameters() {
	fmt.Println("<BR>")
	fmt.Printf("<H3>The list of %d parameters received by %s</H3>\n",fields_count,script_name)
	fmt.Println("<BR>")

	if fields_count > 0 {
		fmt.Println("<TABLE border='1' cellspacing='0' cellpadding='3'>")
		fmt.Println("<THEAD>")
		fmt.Println("<TR class='th'><TD>Name</TD><TD>Value</TD></TR>")
		fmt.Println("</THEAD>")
		fmt.Println("<TBODY style='font-size: 12px;'>")
		for key, value := range fields_map {
			fmt.Println("<TR><TD>" + key + "</TD><TD>" + value + "</TD></TR>")
		}
		fmt.Println("</TBODY>")
		fmt.Println("</TABLE>")
	} // IF
	fmt.Println("<BR><BR>")

} // display_parameters

//////////////////////////////////////////////////////////////////////
//
// Function  : decrypt_table
//
// Purpose   : Read contents of Smart Users History table into an array of structures
//
// Inputs    : secret_key string - decryption key for database table
//
// Output    : (none)
//
// Returns   : nothing
//
// Example   : decrypt_table()
//
// Notes     : The retrieved records are left in a global variable
//
///////////////////////////////////////////////////////////////////////

func decrypt_table(secret_key string) {
	var query string
	var count int
	var buffer string
	var db *sql.DB
	var err1 error

	encrypted_map = make(map[int]int)  // initialize the map
	num_encrypted_records = 0


	db = connect_to_database()

	query = "select count(*) num_records from " + encrypted_data_control_table
	if err1 := db.QueryRow(query).Scan(&count); err1 != nil {
		database_error("<H2>database error. counting query failed on control table<BR>%q</H2>",err1)
	}
	if count <= 0 { // IF no data in control table
		query = "INSERT INTO " + encrypted_data_control_table + " ( encrypted_data ) "  +
				"VALUES ( aes_encrypt('" + dummy_test_string + "','" + secret_key + "')"
		stmtIns, err5 := db.Prepare(query)
		if err5 != nil {
			database_error("<H2>database error. prepare failed for insert into control table<BR>%q</H2>",err5)
		}
		_, err6 := stmtIns.Exec()
		if err6 != nil {
			database_error("<H2>database error. insert into control table failed<BR>%q</H2>",err6)
		}
		stmtIns.Close()
	} else {
		query = "select aes_decrypt(encrypted_data,'" + secret_key + "') " +
					"from " + encrypted_data_control_table
		if err4 := db.QueryRow(query).Scan(&buffer); err4 != nil {
			buffer = "<H2>database error. data query failed on control table<BR>" + query +
						"<BR>%q</H2>"
			database_error(buffer,err4)
		}
		if buffer != dummy_test_string {
			fmt.Println("<H3>Invalid secret key. Mismatch against control data string</H3>")
			fmt.Println("</BODY></HTML>")
		}
	} // ELSE data already in control table

	query = "select id,aes_decrypt(encrypted_data,'" + secret_key + "'),modified_date" +
				" from " + encrypted_data_table

	results2, err2 := db.Query(query)
	if err2 != nil {
		buffer = "<H2>database error. data query failed on table<BR>" + query +
					"<BR>%q</H2>"
		database_error(buffer,err2)
	}

	for results2.Next() {
		num_encrypted_records += 1
		err1 = results2.Scan(&encrypted_rec.id,&encrypted_rec.encrypted_data,&encrypted_rec.modified_date)
		if err1 != nil {
			database_error("<H2>database error. failed to scan record from data table<BR>%q</H2>",err1)
		}
		encrypted_records = append(encrypted_records,encrypted_rec)
		encrypted_map[encrypted_rec.id] = num_encrypted_records - 1 // save index into encrypted_records
	} // FOR
	results2.Close()

	return
} // decrypt_table

//////////////////////////////////////////////////////////////////////
//
// Function  : generate_action_button
//
// Purpose   : Generate an action button form
//
// Inputs    : id int - id number of encryption entry which is used to generate form name
//             function_word string - function word for form
//             submit_label string - label for submit button
//             form_number int - form number (1 or 2)
//             secret_key string - encryption key
//
// Output    : (none)
//
// Returns   : the requested action button form
//
// Example   : generate_action_button(22,"delete","Delete",1,secret_key)
//             generate_action_button(22,"modify","Modify",2,secret_key)
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func generate_action_button(id int, function_word string, submit_label string, form_number int, secret_key string) string {
	var form_name string
	var action_button_form string
	var onclick_prompt string

	onclick_prompt = fmt.Sprintf("%s this record", function_word)
	form_name = fmt.Sprintf("data_form_%d_%d",form_number,id)
	action_button_form = "<div style='float: left; width: 65px;'>\n"
	action_button_form += fmt.Sprintf("<FORM class='button_form' name='%s' id='%s' method='POST' action='%s'>\n",
		form_name,form_name,script_name)
	action_button_form += fmt.Sprintf("<input type='hidden' id='id' name='id' value='%d' />\n",id)
	action_button_form += fmt.Sprintf("<input type='hidden' id='function' name='function' value='%s' />\n",function_word)
	action_button_form += fmt.Sprintf("<input type='hidden' id='secret_key' name='secret_key' value='%s' />\n",secret_key)
	action_button_form += fmt.Sprintf("<input type='submit' id='submit1' name='submit1' value='%s' \n",submit_label)
	action_button_form += fmt.Sprintf("onclick=\"return show_confirm_2(data_form_%d_%d,'%s');\" ",form_number,id,onclick_prompt)
	action_button_form += "/>\n"
	action_button_form += "</FORM>\n"
	action_button_form += "</div>\n"

	return action_button_form
} // generate_action_button

//////////////////////////////////////////////////////////////////////
//
// Function  : generate_query_screen
//
// Purpose   : Generate the query screen
//
// Inputs    : (none)
//
// Output    : the query screen
//
// Returns   : nothing
//
// Example   : generate_query_screen()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func generate_query_screen() {
	var form_name string = "query_record_form"

	fmt.Println("<H3>Note that the query will be performed in case insensitive mode</H3>")
	fmt.Printf("<FORM name='%s' id='%s' method='POST' action='%s'>\n",form_name,form_name,script_name)
	fmt.Println("<input type='hidden' name='function' value='query_records'>")
	fmt.Printf("<TABLE border='1' cellspacing='0' cellpadding='5' rules='groups' frame='box'>\n")
	fmt.Println("<THEAD class='bg_silver'>")
	fmt.Println("<TR><TH colspan='2'>Enter Query Information</TH></TR>")
	fmt.Println("</HEAD>")
	fmt.Println("<TBODY class='bg_wheat'>")
	fmt.Println("<TR><TD>Secret Key</TD>")
	fmt.Println("<TD><input type='password' title='Secret Key' name='secret_key' id='secret_key' size='20' autofocus /></TD>")
	fmt.Println("</TR>")
	fmt.Println("<TR><TD>Query Term</TD>")
	fmt.Println("<TD><input type='text' title='Record Data' name='query_data' id='query_data' size='60' /></TD>")
	fmt.Println("</TR>")
	fmt.Println("</TBODY></TABLE><BR><BR>")
	fmt.Println("<input type='submit' class='submit1' value='Search Records' />")

	fmt.Printf("</FORM>\n")

	return
} // generate_query_screen

//////////////////////////////////////////////////////////////////////
//
// Function  : generate_list_all_records_screen
//
// Purpose   : Generate the list_all_records screen
//
// Inputs    : (none)
//
// Output    : the list_all_records screen
//
// Returns   : nothing
//
// Example   : generate_list_all_records_screen()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func generate_list_all_records_screen() {
	var form_name string = "listall_records_form"


	fmt.Printf("<FORM name='%s' id='%s' method='POST' action='%s'>\n",form_name,form_name,script_name)
	fmt.Println("<input type='hidden' name='function' value='listall2'>")
	fmt.Printf("<TABLE border='1' cellspacing='0' cellpadding='5' rules='groups' frame='box'>\n")
	fmt.Println("<THEAD class='bg_silver'>")
	fmt.Println("<TR><TH colspan='2'>Enter Decryption Key</TH></TR>")
	fmt.Println("</HEAD>")
	fmt.Println("<TBODY class='bg_wheat'>")
	fmt.Println("<TR><TD>Secret Key</TD>")
	fmt.Println("<TD><input type='password' title='Secret Key' name='secret_key' id='secret_key' size='20' autofocus /></TD>")
	fmt.Println("</TR>")
	fmt.Println("</TBODY></TABLE><BR><BR>")
	fmt.Println("<input type='submit' class='submit1' value='List All Records' />")

	fmt.Printf("</FORM>\n")


	return
} // generate_list_all_records_screen

//////////////////////////////////////////////////////////////////////
//
// Function  : display_one_record
//
// Purpose   : Display the contents of the record
//
// Inputs    : record EncryptedRec - the record to be displayed
//             secret_key string - the decryption key
//
// Output    : the specified record
//
// Returns   : nothing
//
// Example   : display_one_record(encrypted_records[index],secret_key)
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func display_one_record (record EncryptedRec, secret_key string) {

	var delete_action string
	var modify_action string
	var actions string

	delete_action = generate_action_button(record.id,"delete","Delete",1,secret_key)
	modify_action = generate_action_button(record.id,"modify","Modify",2,secret_key)
	actions = "<div style='width: 160px;'>" + delete_action + modify_action + "</div>"
	fmt.Printf("<TR %s><TD>%d</TD><TD>%s</TD><TD>%s</TD><TD>%s</TD></TR>\n",
					onmouseover,record.id,record.modified_date,record.encrypted_data,actions)

	return
} // display_one_record

//////////////////////////////////////////////////////////////////////
//
// Function  : list_all_records
//
// Purpose   : List the decrypted data from all of the records
//
// Inputs    : (none)
//
// Output    : all of the records
//
// Returns   : nothing
//
// Example   : list_all_records()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func list_all_records() {
	var secret_key string
	var ok bool
	var index int

	secret_key , ok = fields_map["secret_key"]
	if !ok {
		fmt.Println("<H3>Error : No value was specified for the secret key</H3>")
		return
	}
	decrypt_table(secret_key)

	fmt.Printf("<H3>%d records were retrieved</H3>\n",num_encrypted_records)
	fmt.Println("<TABLE border='1' cellspacing='0' cellpadding='3'>")
	fmt.Println("<THEAD>")
	fmt.Printf("<TR class='th'><TH>Id</TH><TH>Date</TH><TH>Data</TH><TH>Actions</TH></TR>\n")
	fmt.Println("</THEAD>")
	fmt.Println("<TBODY>")

	for index = 0 ; index < num_encrypted_records ; index ++ {
		display_one_record(encrypted_records[index],secret_key)
	}
	fmt.Printf("</TBODY></TABLE>\n")

	return
} // list_all_records

//////////////////////////////////////////////////////////////////////
//
// Function  : query_records
//
// Purpose   : List the decrypted data from all of the records
//
// Inputs    : (none)
//
// Output    : all of the records
//
// Returns   : nothing
//
// Example   : query_records()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func query_records() {
	var secret_key string
	var query_data string
	var pattern string
	var data string
	var ok bool
	var index int
	var num_matched int

	secret_key , ok = fields_map["secret_key"]
	if !ok {
		fmt.Println("<H3>Error : No value was specified for the secret key</H3>")
		return
	}
	query_data , ok = fields_map["query_data"]
	if !ok {
		fmt.Println("<H3>Error : No value was specified for the query term</H3>")
		return
	}
	pattern = fmt.Sprintf("(?i)%s",query_data)
	re := regexp.MustCompile(pattern)

	decrypt_table(secret_key)
	fmt.Printf("<H3>%d records were retrieved</H3>\n",num_encrypted_records)
	num_matched = 0
	for index = 0 ; index < num_encrypted_records ; index ++ {
		data = encrypted_records[index].encrypted_data
		if re.Match([]byte(data)) {
			num_matched += 1
			if num_matched == 1 {
				fmt.Println("<TABLE border='1' cellspacing='0' cellpadding='3'>")
				fmt.Println("<THEAD>")
				fmt.Printf("<TR class='th'><TH>Id</TH><TH>Date</TH><TH>Data</TH><TH>Actions</TH></TR>\n")
				fmt.Println("</THEAD>")
				fmt.Println("<TBODY>")
			}
			display_one_record(encrypted_records[index],secret_key)
		} // IF a match
	} // FOR
	if num_matched > 0 {
		fmt.Println("</TBODY></TABLE>")
	}
	fmt.Printf("<BR><H3>For '%s' the number of matched records = %d</H3>\n",query_data,num_matched)

	return
} // query_records

//////////////////////////////////////////////////////////////////////
//
// Function  : describe_table
//
// Purpose   : Display the description of one database table
//
// Inputs    : (none)
//
// Output    : the description for the named table
//
// Returns   : nothing
//
// Example   : describe_table("mytable")
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func describe_table(table_name string) {
	var (
		ordinal        int
		colname        string
		isnull         string
		maxlen         string
		column_type    string
		extra          string
		column_key     string
		comment        string
		db				*sql.DB
		err1			error
	)

	fmt.Printf("<H3>Description of Table %s</H3>\n",table_name)
	db = connect_to_database()

	query := "select ordinal_position ordinal, column_name colname,is_nullable isnull," +
					"ifnull(character_maximum_length,'--') maxlen,column_type,extra,column_key," +
					"ifnull(column_comment,'--') comment" +
					" from information_schema.columns where table_schema = 'qwlc' and " +
					"table_name = ?"
	rows, err2 := db.Query(query,table_name)
	if err2 != nil {
		database_error("<H2>database error. db.Query failed<BR>%q</H2>",err2)
	}
	defer rows.Close()
	num_cols := 0
	for rows.Next() {
		err1 = rows.Scan(&ordinal, &colname, &isnull, &maxlen, &column_type, &extra, &column_key, &comment)
		if err1 != nil {
			database_error("<H2>database error. rows.Scan failed<BR>%q</H2>",err1)
		}
		num_cols += 1
		if num_cols == 1 {
			fmt.Println("<TABLE border='1' cellspacing='0' cellpadding='3'>")
			fmt.Println("<THEAD>")
			fmt.Printf("<TR class='th'><TH>Ordinal</TH><TH>Column Name</TH><TH>Data Type</TH><TH>Maxlen</TH><TH>Nullable ?</TH><TH>Key</TH><TH>Extra</TH><TH>Comment</TH></TR>\n")
			fmt.Println("</THEAD>")
			fmt.Println("<TBODY>")
		}
		fmt.Println("<TR>")
		fmt.Printf("<TD>%d</TD><TD>%s</TD><TD>%s</TD><TD>%s</TD><TD>%s</TD><TD>%s</td><TD>%s</TD><TD>%s</TD>\n",
						ordinal,colname,column_type,maxlen,isnull,column_key,extra,comment)
		fmt.Println("</TR>")
	} // FOR
	if num_cols > 0 {
		fmt.Println("</TBODY></TABLE>")
	} else {
		fmt.Printf("<H3>No information found for '%s'</H3>\n",table_name)
	}
	fmt.Println("<BR>")

	return
} // describe_table

//////////////////////////////////////////////////////////////////////
//
// Function  : list_metadata
//
// Purpose   : Display the metadata
//
// Inputs    : (none)
//
// Output    : the metadata for the two tables
//
// Returns   : nothing
//
// Example   : list_metadata()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func list_metadata() {

	describe_table("my_encrypted")
	describe_table("my_encrypted_control")

	return
} // list_metadata

//////////////////////////////////////////////////////////////////////
//
// Function  : generate_add_record_screen
//
// Purpose   : Generate the add new record screen
//
// Inputs    : (none)
//
// Output    : the add new record screen
//
// Returns   : nothing
//
// Example   : generate_add_record_screen()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func generate_add_record_screen() {
	var form_name string = "new_record_form"

	fmt.Printf("<FORM class='form2' name='%s' id='%s' method='POST' action='%s'>\n",form_name,form_name,script_name)
	fmt.Println("<input type='hidden' name='function' value='add_new_record'>")
	fmt.Printf("<TABLE border='1' cellspacing='0' cellpadding='5' rules='groups' frame='box'>\n")
	fmt.Println("<THEAD class='bg_silver'>")
	fmt.Println("<TR><TH colspan='2'>Enter New Record Information</TH></TR>")
	fmt.Println("</HEAD>")
	fmt.Println("<TBODY class='bg_wheat'>")
	fmt.Println("<TR><TD>Secret Key</TD>")
	fmt.Println("<TD><input type='password' title='Secret Key' name='secret_key' id='secret_key' size='20' autofocus /></TD>")
	fmt.Println("</TR>")
	fmt.Println("<TR><TD>Record Data</TD>")
	fmt.Println("<TD><input type='text' title='Record Data' name='record_data' id='record_data' size='100' /></TD>")
	fmt.Println("</TR>")
	fmt.Println("</TBODY></TABLE><BR><BR>")
	fmt.Println("<input type='submit' class='submit1' value='Add New Record' />")

	fmt.Printf("</FORM>\n")

	return
} // generate_add_record_screen

//////////////////////////////////////////////////////////////////////
//
// Function  : add_new_record
//
// Purpose   : Process a request to add a new record
//
// Inputs    : (none)
//
// Output    : appropriate messages
//
// Returns   : nothing
//
// Example   : add_new_record()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func add_new_record() {
	var secret_key string
	var record_data string
	var ok bool
	var num_errors int = 0
	var insert_stmt string
	var db *sql.DB

	secret_key , ok = fields_map["secret_key"]
	if !ok || 0 == len(secret_key) {
		fmt.Println("<H3>Error : No value was specified for the secret key</H3>")
		num_errors += 1
	}

	record_data , ok = fields_map["record_data"]
	if !ok || 0 == len(record_data) {
		fmt.Println("<H3>Error : No value was specified for the record data</H3>")
		num_errors += 1
	}
	if num_errors > 0 {
		return
	}
	fmt.Printf("<H3>secret_key = '%s' , record_data = '%s'</H3>\n",secret_key,record_data)

// insert record into encrypted table

	db = connect_to_database()

	insert_stmt = "INSERT INTO my_encrypted " + "(modified_date , encrypted_data) " +
					"VALUES (now() , aes_encrypt(' " + record_data + "','" + secret_key + "') )"
	debug_print("SQL for insert is<BR>" + insert_stmt)
	stmtIns, err3 := db.Prepare(insert_stmt)
	if err3 != nil {
		fmt.Printf("<H3>Database Error. Failed to prepare sql<BR>%s<BR>%q</H3></BODY></HTML>\n",insert_stmt,err3)
		db.Close()
		os.Exit(0)
	}

	debug_print("Execute insert request")
	_, err4 := stmtIns.Exec()
	if err4 != nil {
		fmt.Printf("<H3>Database Error. Failed to insert record<BR>%s<BR>%q</H3></BODY></HTML>\n",insert_stmt,err4)
		stmtIns.Close()
		db.Close()
		os.Exit(0)
	}
	fmt.Println( "<H3>Insert was successfull</H3>")
	stmtIns.Close()
	db.Close()
	return
} // add_new_record

//////////////////////////////////////////////////////////////////////
//
// Function  : generate_modify_record_screen
//
// Purpose   : Generate the modify record screen
//
// Inputs    : (none)
//
// Output    : the modify record screen
//
// Returns   : nothing
//
// Example   : generate_modify_record_screen()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func generate_modify_record_screen() {
	var form_name string = "old_record_form"
	var secret_key string
	var id string
	var id2 int
	var num_errors int = 0
	var ok bool
	var record EncryptedRec
	var index int

	secret_key , ok = fields_map["secret_key"]
	if !ok || 0 == len(secret_key) {
		fmt.Println("<H3>Error : No value was specified for the secret key</H3>")
		num_errors += 1
	}
	id , ok = fields_map["id"]
	if !ok || 0 == len(id) {
		fmt.Println("<H3>Error : No value was specified for the id</H3>")
		num_errors += 1
	}
	id2, _ = strconv.Atoi(id)
	if num_errors < 0 {
		return
	}
	decrypt_table(secret_key)

	// encrypted_records = append(encrypted_records,encrypted_rec)
	// encrypted_map[encrypted_rec.id] = num_encrypted_records - 1 // save index into encrypted_records
	index = encrypted_map[id2]
	record = encrypted_records[index]


	fmt.Printf("<FORM class='form2' name='%s' id='%s' method='POST' action='%s'>\n",form_name,form_name,script_name)
	fmt.Println("<input type='hidden' name='function' value='modify2'>")
	fmt.Printf("<input type='hidden' name='id' value='%s'>\n",id)
	fmt.Printf("<TABLE border='1' cellspacing='0' cellpadding='5' rules='groups' frame='box'>\n")
	fmt.Println("<THEAD class='bg_silver'>")
	fmt.Printf("<TR><TH colspan='2'>Modify Current Information for Record %s</TH></TR>\n",id)
	fmt.Println("</HEAD>")
	fmt.Println("<TBODY class='bg_wheat'>")
	fmt.Println("<TR><TD>Secret Key</TD>")
	fmt.Println("<TD><input type='password' title='Secret Key' name='secret_key' id='secret_key' size='20' autofocus /></TD>")
	fmt.Println("</TR>")
	fmt.Println("<TR><TD>Record Data</TD>")
	fmt.Printf("<TD><input type='text' title='Record Data' name='record_data' id='record_data' size='100' value='%s'/></TD>\n",record.encrypted_data)
	fmt.Println("</TR>")
	fmt.Println("</TBODY></TABLE><BR><BR>")
	fmt.Println("<input type='submit' class='submit1' value='Modify Current Record Data' />")

	fmt.Printf("</FORM>\n")

	return
} // generate_modify_record_screen

//////////////////////////////////////////////////////////////////////
//
// Function  : modify_existing_record
//
// Purpose   : Process a request to modify a record
//
// Inputs    : (none)
//
// Output    : appropriate messages
//
// Returns   : nothing
//
// Example   : modify_existing_record()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func modify_existing_record() {
	var secret_key string
	var id string
	var id2 int
	var num_errors int = 0
	var ok bool
	var record_data string
	var db *sql.DB
	var sql_update string

	secret_key , ok = fields_map["secret_key"]
	if !ok || 0 == len(secret_key) {
		fmt.Println("<H3>Error : No value was specified for the secret key</H3>")
		num_errors += 1
	}
	id , ok = fields_map["id"]
	if !ok || 0 == len(id) {
		fmt.Println("<H3>Error : No value was specified for the id</H3>")
		num_errors += 1
	}
	id2, _ = strconv.Atoi(id)
	if num_errors < 0 {
		return
	}
	record_data , ok = fields_map["record_data"]
	if !ok || 0 == len(record_data) {
		fmt.Println("<H3>Error : No value was specified for the record_data</H3>")
		num_errors += 1
	}
	if num_errors > 0 {
		return
	}
	// decrypt_table(secret_key)

	fmt.Printf("<H3>New data for record %d is<BR>%s</H3>\n",id2,record_data)
	db = connect_to_database()

	// Prepare statement for updating data
	sql_update = fmt.Sprintf("UPDATE my_encrypted set encrypted_data = aes_encrypt('%s','%s') , modified_date = now() WHERE id = %d",
					record_data,secret_key,id2)
	debug_print("SQL for UPDATE operation is<BR>" + sql_update)
	stmt, err := db.Prepare(sql_update)
	if err != nil {
		database_error("<H2>database error. failed to prepare sql for update to record_data</H2>",err);
	}

	debug_print("Exec request to update record_data")
	res, err2 := stmt.Exec()
	if err2 != nil {
		database_error("<H2>database error. failed to update record_data</H2>",err2);
	}
	affect, err3 := res.RowsAffected()
	if err3 != nil {
		database_error("<H2>database error. possibly failed to update record_data",err3);
	}
	fmt.Printf("<H3>Number of updated rows = %d</H3>\n",affect)

	debug_print("Close stmt and db handles used for record update")
	stmt.Close()
	db.Close()

	return
} // modify_existing_record

//////////////////////////////////////////////////////////////////////
//
// Function  : delete_record
//
// Purpose   : Process a request to delete a record
//
// Inputs    : (none)
//
// Output    : appropriate messages
//
// Returns   : nothing
//
// Example   : delete_record()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func delete_record() {
	var id string
	var id2 int
	var num_errors int = 0
	var ok bool
	var db *sql.DB
	var sql_delete string

	id , ok = fields_map["id"]
	if !ok || 0 == len(id) {
		fmt.Println("<H3>Error : No value was specified for the id</H3>")
		num_errors += 1
	}
	id2, _ = strconv.Atoi(id)
	if num_errors < 0 {
		return
	}
	db = connect_to_database()

	// Prepare statement for updating data
	sql_delete = fmt.Sprintf("DELETE from my_encrypted WHERE id = %d",id2)
	debug_print("SQL for ZDELETE operation is<BR>" + sql_delete)
	stmt, err := db.Prepare(sql_delete)
	if err != nil {
		database_error("<H2>database error. failed to prepare sql for delete of record</H2>",err);
	}

	debug_print("Exec request to delete record")
	res, err2 := stmt.Exec()
	if err2 != nil {
		database_error("<H2>database error. failed to delete record</H2>",err2);
	}
	affect, err3 := res.RowsAffected()
	if err3 != nil {
		database_error("<H2>database error. possibly failed to delete record",err3);
	}
	fmt.Printf("<H3>Number of deleted rows = %d</H3>\n",affect)

	debug_print("Close stmt and db handles used for record delete")
	stmt.Close()
	db.Close()

	return
} // delete_record

//////////////////////////////////////////////////////////////////////
//
// Function  : generate_menu_entry
//
// Purpose   : Generate a menu entry for the main screen
//
// Inputs    : menu_title string - menu title
//             function_word string - command function word
//
// Output    : the main screen menu entry
//
// Returns   : nothing
//
// Example   : generate_menu_entry("List Data","listall")
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func generate_menu_entry(menu_title string, function_word string) {
	var form_name string
	var div_name string

	form_name = "form_" + function_word
	div_name = "div_" + function_word
	fmt.Println("<BR>")
	fmt.Printf("<FORM name='%s' id='%s' method='POST' action='%s'>\n",form_name,div_name,script_name)
	fmt.Printf("<input type='hidden' name='function' id='function' value='%s'>\n",function_word)
	fmt.Printf("<input class='submit1' type='submit' value='%s'>\n",menu_title)
	fmt.Println("</FORM>")

	return
} // generate_menu_entry

//////////////////////////////////////////////////////////////////////
//
// Function  : generate_main_screen
//
// Purpose   : Generate the main screen
//
// Inputs    : (none)
//
// Output    : the main screen
//
// Returns   : nothing
//
// Example   : generate_main_screen()
//
// Notes     :
//
///////////////////////////////////////////////////////////////////////

func generate_main_screen() {

	generate_menu_entry("List All Records","listall")
	generate_menu_entry("Query Records","query")
	generate_menu_entry("Add New Record","add")
	generate_menu_entry("List Metadata","meta")

	return
} // generate_main_screen

//////////////////////////////////////////////////////////////////////
//
// Function  : main
//
// Purpose   : Go CGI script to generate XML output for Smart Contracts
//
// Inputs    : (none)
//
// Output    : XML processing results
//
// Returns   : nothing
//
// Example   : main()
//
// Notes     : (none)
//
///////////////////////////////////////////////////////////////////////

func main() {
	var function string
	var ok bool
	var debug string

    fmt.Printf("Content-Type: text/html\n\n");
	fmt.Println("<HTML>")
	fmt.Println("<HEAD>")
	fmt.Printf("<TITLE>%s</TITLE>\n",title)
	fmt.Println("<link rel='stylesheet' type='text/css' href='/styles.css'>")
	fmt.Println("<link rel='stylesheet' type='text/css' href='/fieldset.css'>")
	fmt.Println("<script src='/show_confirm_2.js' type='text/javascript'></script>")
	fmt.Println("<style type='text/css' media='print,screen'>")
	fmt.Println(".bg_wheat { background-color: wheat; }")
	fmt.Println(".bg_silver { background-color: lightgrey; }")
	fmt.Println("</style>")
	fmt.Println("</HEAD>")
	fmt.Println("<BODY>")
	fmt.Println("<div style='padding-left: 20px;'>")
	fmt.Printf("<H2>%s</H2>\n",title)
	script_name = filepath.Base(os.Args[0])
	top_level_href = fmt.Sprintf("<A class='boldlink3' href='%s'>Return to Main Screen</A>",script_name)
	onmouseover = "onMouseOver=\"this.style.background='peru';this.style.color='white';this.style.fontWeight=900;return true;\"" +
			"onMouseOut=\"this.style.backgroundColor=''; this.style.color='black';this.style.fontWeight=500;\""

	defer func() { //catch or finally
        if err := recover(); err != nil { //catch
			fmt.Printf("<BR>PANIC : %v<BR>",err)
			fmt.Printf("</BODY></HTML>")
			os.Exit(0)
       }
    }()
	
	parse_fields()
	debug , ok = fields_map["secret_key"]
	if ok {
		switch debug {
			case "on":
				debug_mode = true
				break
			case "off":
				debug_mode = false
				break
			default:
				fmt.Printf("<H3>Invaluid value '%s' specified for debug parameter.</H3>\n",debug)
		}
	}
	if debug_mode {
		display_parameters()
	}
	function, ok = fields_map["function"]
	if ok {
		switch function {
		case "listall":
			generate_list_all_records_screen()
			break
		case "listall2":
			list_all_records()
			break
		case "query":
			generate_query_screen()
			break
		case "query_records":
			query_records()
			break
		case "meta":
			list_metadata()
			break
		case "add":
			generate_add_record_screen()
			break
		case "add_new_record":
			add_new_record()
			break
		case "modify":
			generate_modify_record_screen()
			break
		case "modify2":
			modify_existing_record()
			break
		case "delete":
			delete_record()
			break
		default:
			fmt.Printf("<H3>Error : '%s' is not a supported function</H3>\n",function)
		} // switch
		fmt.Printf("<BR>%s<BR><BR>\n",top_level_href)
	} else {
		generate_main_screen()
	}

	fmt.Println("</div>")
	fmt.Println("</BODY></HTML>")
	os.Exit(0)
} // main
