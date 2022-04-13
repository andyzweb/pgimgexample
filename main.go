package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

var conn *pgx.Conn

func main() {
	var err error
	conn, err = pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	if len(os.Args) == 1 {
		help()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "ddl":
		err := ddl()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to Create Table: %v\n", err)
			os.Exit(1)
		}

	case "add":
		err := add(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Problem inserting record: %v\n", err)
			os.Exit(1)
		}

	case "read":
		width, height, err := read(os.Args[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Problem selecting record: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "Image %v dimensions: width %v, height %v\n", os.Args[2], width, height)
	default:
		help()
	}
}

func help() {
	fmt.Print(`Usage:
    ddl - creates tables in database
    add  [filename] - adds metadata about an image to the database
    read [filename] - retreives metadata about an image
`)
}

func ddl() error {
	fmt.Println("This will run CREATE TABLE etc etc")
	var ddlSql string = `CREATE TABLE IF NOT EXISTS images (
    filename text PRIMARY KEY,
    pixel_width integer NOT NULL,
    pixel_height integer NOT NULL
    )`
	_, err := conn.Exec(context.Background(), ddlSql)
	return err
}

func add(filename string) error {
	//open file -> get dims -> store to DB
	fmt.Println("This will add image: ", filename)
	width, height := getImageDimension(filename)
	insertSql := "INSERT INTO images (filename, pixel_width, pixel_height) VALUES ($1,$2,$3)"
	_, err := conn.Exec(context.Background(), insertSql, filename, width, height)
	return err
}

func read(filename string) (int, int, error) {
	fmt.Println("This will retreive metadata for image: ", filename)
	var width, height int
	selectSql := "SELECT pixel_width, pixel_height FROM images WHERE filename = $1"
	err := conn.QueryRow(context.Background(), selectSql, filename).Scan(&width, &height)
	return width, height, err
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	defer file.Close()

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", imagePath, err)
	}
	return image.Width, image.Height
}
