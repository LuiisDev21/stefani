package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/LuiisDev21/stefani/internal/ytaccount"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: homerecon <profile> <out.json> [browseId]")
		os.Exit(2)
	}
	client := ytaccount.New(os.Args[1])
	ctx := context.Background()

	f, err := os.Create(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", " ")

	if len(os.Args) > 3 {
		doc, err := client.ProbeBrowse(ctx, os.Args[3])
		if err != nil {
			fmt.Fprintln(os.Stderr, "probe:", err)
			os.Exit(1)
		}
		if err := enc.Encode(doc); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("ok raw")
		return
	}

	shelves, err := client.Home(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "home:", err)
		os.Exit(1)
	}
	if err := enc.Encode(shelves); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("ok home")
}
