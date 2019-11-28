//////////////////////////////////////////////////////////////////////
//
// File      : mysql-describe-2.go
//
// Author    : Barry Kimelman
//
// Created   : November 28, 2019
//
// Purpose   : Go MySQL describe table structure
//
// Notes     : adapted from mysql-describe.go
//
//////////////////////////////////////////////////////////////////////

package main

import (
    "database/sql"
    "log"
	"fmt"
	"os"

    _ "github.com/go-sql-driver/mysql"
)

func main() {
    argsWithProg := os.Args
	num_args := len(os.Args)
	if ( num_args < 2 ){
		log.Fatal("Usage : ",argsWithProg[0]," table_name\n")
	}
	table_name := argsWithProg[1]
	fmt.Println("\nDescription of Table ",table_name,"\n")

    db, err := sql.Open("mysql", "myusername:mypassword@(127.0.0.1:3306)/myschema?parseTime=true")
    if err != nil {
        log.Fatal(err)
    }
    if err := db.Ping(); err != nil {
        log.Fatal(err)
    }

    { // Query the columns of the named table
        var (
            ordinal        int
            colname        string
            isnull         string
            maxlen         string
			column_type    string
			extra          string
			column_key     string
			comment        string
        )
		var colnames []string
		var data_types []string
		var maxlens []string
		var nulls []string
		var col_keys []string
		var extras []string
		var comments []string
		var headers = [7]string{"Column Name", "Data Type", "Maxlen", "Nullable ?", "Key" , "Extra" , "Comment" }
		colname_maxlen := len(headers[0])
		data_type_maxlen := len(headers[1])
		maxlen_maxlen := len(headers[2])
		nullable_maxlen := len(headers[3])
		key_maxlen := len(headers[4])
		extra_maxlen := len(headers[5])
		query := "select ordinal_position ordinal, column_name colname,is_nullable isnull," +
					"ifnull(character_maximum_length,'--') maxlen,column_type,extra,column_key," +
					"ifnull(column_comment,'--') comment" +
					" from information_schema.columns where table_schema = 'qwlc' and " +
					"table_name = ?"
        rows, err := db.Query(query,table_name)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		num_cols := 0
		for rows.Next() {
			err := rows.Scan(&ordinal, &colname, &isnull, &maxlen, &column_type, &extra, &column_key, &comment)
			if err != nil {
				log.Fatal(err)
			}
			num_cols += 1
			// fmt.Println(ordinal, colname, isnull, maxlen, column_type, extra, column_key, comment)
			colnames = append(colnames, colname)
			data_types = append(data_types, column_type)
			maxlens = append(maxlens,maxlen)
			nulls = append(nulls,isnull)
			col_keys = append(col_keys,column_key)
			extras = append(extras,extra)
			comments = append(comments,comment)
			count := len(colname)
			if ( count > colname_maxlen ) {
				colname_maxlen = count
			}
			count = len(column_type)
			if ( count > data_type_maxlen ) {
				data_type_maxlen = count
			}
			count = len(maxlen)
			if ( count > maxlen_maxlen ) {
				maxlen_maxlen = count
			}
			count = len(isnull)
			if ( count > nullable_maxlen ) {
				nullable_maxlen = count
			}
			count = len(column_key)
			if ( count > key_maxlen ) {
				key_maxlen = count
			}
			count = len(extra)
			if ( count > extra_maxlen ) {
				extra_maxlen = count;}
		} // for
		var longest = [6]int{colname_maxlen, data_type_maxlen, maxlen_maxlen, nullable_maxlen, key_maxlen , extra_maxlen }

		fmt.Printf("\n")
		for index := 0 ; index < 6 ; index++ {
			fmt.Printf("%-*.*s ",longest[index],longest[index],headers[index])
		}
		
		fmt.Printf("%s\n",headers[6])
		for index := 0 ; index < num_cols ; index++ {
			fmt.Printf("%-*.*s ",longest[0],longest[0],colnames[index])
			fmt.Printf("%-*.*s ",longest[1],longest[1],data_types[index])
			fmt.Printf("%-*.*s ",longest[2],longest[2],maxlens[index])
			fmt.Printf("%-*.*s ",longest[3],longest[3],nulls[index])
			fmt.Printf("%-*.*s ",longest[4],longest[4],col_keys[index])
			fmt.Printf("%-*.*s ",longest[5],longest[5],extras[index])
			fmt.Printf("%s",comments[index])
			fmt.Printf("\n")
		}
    }

} // main
