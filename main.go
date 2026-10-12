package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/loomlabs-ai/loom/internal/config"
	"github.com/loomlabs-ai/loom/internal/server"
	"github.com/loomlabs-ai/loom/internal/server/store"
)

func main() {
	port := flag.String("port", ":8080", "Web server port")

	flag.Parse()

	// Create folder
	err := config.Init()

	if err != nil {
		fmt.Println("config: failed to create loom folder")
		fmt.Println("errror: ", err)
		os.Exit(1)
	}

	loomPath := config.GetLoomPath()

	// Connect database
	err = store.ConnectDB(
		filepath.Join(loomPath, "loom.db"),
	)

	if err != nil {
		fmt.Println("connect database error")
		fmt.Println("error: ", err)
		os.Exit(1)
	}

	defer store.CloseDB()

	err = store.CreateTables()

	if err != nil {
		fmt.Println("create tables error")
		fmt.Println("error: ", err)
		os.Exit(1)
	}

	server.StartServer(*port)
}
