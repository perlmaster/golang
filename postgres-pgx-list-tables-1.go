//////////////////////////////////////////////////////////////////////
//
// File      : postgres-pgx-list-tables-1.go
//
// Author    : Barry Kimelman
//
// Created   : January 26, 2020
//
// Purpose   : Use pgx package to display a list of tables under a database
//
// Notes     :
//
//////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx"
)

func main() {
	var (
		database_url	string
		schema			string
		name			string
		table_type		string
		owner			string
		size			string
		description		string
	)
	var headers = [6]string{ "Schema" , "Name" , "Type" , "Owner" , "Size" , "Description" }
	var underlines = [6]string{ "======" , "====" , "====" , "=====" , "====" , "===========" }
	var longest = [6]int { 0 , 0 , 0 , 0 , 0 , 0 }
	var index int
	var length int
	var schemas []string
	var names []string
	var types []string
	var owners []string
	var sizes []string
	var descriptions []string

	if len(os.Args) < 3 {
		fmt.Printf("Usage : %s dbname username password\n",os.Args[0])
		os.Exit(1)
	}

	for index = 0 ; index <= 5 ; index++ {
		longest[index] = len(headers[index])
	}

// DATABASE_URL looks like postgres://{user}:{password}@{hostname}:{port}/{database-name}
	dbname := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]
	fmt.Printf("\nList of Available Tables Under Database '%s'\n",dbname)
	database_url = fmt.Sprintf("postgres://%s:%s@localhost:5432/%s",username,password,dbname)

	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	query := "SELECT n.nspname as \"Schema\",   c.relname as \"Name\",   CASE c.relkind   WHEN 'r' THEN 'table'   WHEN 'v' THEN 'view'   WHEN 'm' THEN 'materialized view'   WHEN 'i' THEN 'index'   WHEN 'S' THEN 'sequence'   WHEN 's' THEN 'special'   WHEN 'f' THEN 'foreign table'   WHEN 'p' THEN 'partitioned table'   WHEN 'I' THEN 'partitioned index'   END as \"Type\",   pg_catalog.pg_get_userbyid(c.relowner) as \"Owner\",   pg_catalog.pg_size_pretty(pg_catalog.pg_table_size(c.oid)) as \"Size\",   coalesce( nullif(pg_catalog.obj_description(c.oid, 'pg_class'),'') , '---' ) as \"Description\" FROM pg_catalog.pg_class c      LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace WHERE c.relkind IN ('r','p','v','m','S','f','')       AND n.nspname <> 'pg_catalog'       AND n.nspname <> 'information_schema'       AND n.nspname !~ '^pg_toast'   AND pg_catalog.pg_table_is_visible(c.oid) ORDER BY 1,2;"

	rows, err := conn.Query(context.Background(),query)
	if err != nil {
		panic(err.Error()) // use proper error handling
	}
	count := 0
	for rows.Next() {
		count += 1
		err = rows.Scan(&schema,&name,&table_type,&owner,&size,&description)
		length = len(schema)
		schemas = append(schemas, schema)
		if length > longest[0] {
			longest[0] = length
		}
		length = len(name)
		names = append(names, name)
		if length > longest[1] {
			longest[1] = length
		}
		length = len(table_type)
		types = append(types,table_type)
		if length > longest[2] {
			longest[2] = length
		}
		length = len(owner)
		owners = append(owners,owner)
		if length > longest[3] {
			longest[3] = length
		}
		length = len(size)
		sizes = append(sizes,size)
		if length > longest[4] {
			longest[4] = length
		}
		length = len(description)
		descriptions = append(descriptions,description)
		if length > longest[5] {
			longest[5] = length
		}
		if err != nil {
			panic(err.Error()) // use proper error handling
		}
	} // FOR

	fmt.Printf("\n")
	for index := 0 ; index < 6 ; index++ {
		fmt.Printf("%-*.*s ",longest[index],longest[index],headers[index])
	}
	fmt.Printf("\n")
	for index := 0 ; index < 6 ; index++ {
		fmt.Printf("%-*.*s ",longest[index],longest[index],underlines[index])
	}
	fmt.Printf("\n")
	for index = 0 ; index < count ; index++ {
		fmt.Printf("%-*.*s ",longest[0],longest[0],schemas[index])
		fmt.Printf("%-*.*s ",longest[1],longest[1],names[index])
		fmt.Printf("%-*.*s ",longest[2],longest[2],types[index])
		fmt.Printf("%-*.*s ",longest[3],longest[3],owners[index])
		fmt.Printf("%-*.*s ",longest[4],longest[4],sizes[index])
		fmt.Printf("%s\n",description)
	}

} // main
