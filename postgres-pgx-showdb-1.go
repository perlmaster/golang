//////////////////////////////////////////////////////////////////////
//
// File      : postgres-pgx-showdb.go
//
// Author    : Barry Kimelman
//
// Created   : January 25, 2020
//
// Purpose   : Use pgx package to display a list of databases
//
// Notes     :
//
//////////////////////////////////////////////////////////////////////

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx"
)

func main() {
	var (
		database_url	string
		name			string
		owner			string
		encoding		string
		collate			string
		ctype			string
		access_priv		string
	)
	var headers = [6]string{ "Name" , "Owner" , "Encoding" , "Collate" , "Ctype" , "Access privileges" }
	var underlines = [6]string{ "====" , "=====" , "========" , "=======" , "=====" , "=================" }
	var longest = [6]int { 0 , 0 , 0 , 0 , 0 , 0 }
	var index int
	var length int
	var names []string
	var owners []string
	var encodings []string
	var collates []string
	var ctypes []string
	var access_privs []string

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
	fmt.Printf("\nList of Available Databases\n")
	database_url = fmt.Sprintf("postgres://%s:%s@localhost:5432/%s",username,password,dbname)

	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	query := "SELECT d.datname as \"Name\", pg_catalog.pg_get_userbyid(d.datdba) as \"Owner\", pg_catalog.pg_encoding_to_char(d.encoding) as \"Encoding\", d.datcollate as \"Collate\", d.datctype as \"Ctype\", coalesce(nullif(pg_catalog.array_to_string(d.datacl, E'\\n'),''),'---') AS \"Access privileges\" FROM pg_catalog.pg_database d ORDER BY 1;"

	rows, err := conn.Query(context.Background(),query)
	if err != nil {
		panic(err.Error()) // use proper error handling
	}
	count := 0
	for rows.Next() {
		count += 1
		err = rows.Scan(&name,&owner,&encoding,&collate,&ctype,&access_priv)
		length = len(name)
		names = append(names, name)
		if length > longest[0] {
			longest[0] = length
		}
		length = len(owner)
		owners = append(owners, owner)
		if length > longest[1] {
			longest[1] = length
		}
		length = len(encoding)
		encodings = append(encodings,encoding)
		if length > longest[2] {
			longest[2] = length
		}
		length = len(collate)
		collates = append(collates,collate)
		if length > longest[3] {
			longest[3] = length
		}
		length = len(ctype)
		ctypes = append(ctypes,ctype)
		if length > longest[4] {
			longest[4] = length
		}
		length = len(access_priv)
		access_privs = append(access_privs,access_priv)
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
		fmt.Printf("%-*.*s ",longest[0],longest[0],names[index])
		fmt.Printf("%-*.*s ",longest[1],longest[1],owners[index])
		fmt.Printf("%-*.*s ",longest[2],longest[2],encodings[index])
		fmt.Printf("%-*.*s ",longest[3],longest[3],collates[index])
		fmt.Printf("%-*.*s ",longest[4],longest[4],ctypes[index])
		acc := strings.ReplaceAll(access_privs[index], "\n", "\\n")
		fmt.Printf("%s\n",acc)
	}

} // main
