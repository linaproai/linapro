// Package main boots the development-only source upgrade tool and intentionally
// keeps the repository-root package limited to process startup.
package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"

	"upgrade-source/internal/sourceupgrade"
)

// main delegates all command parsing and execution to the internal source-upgrade package.
func main() {
	if err := sourceupgrade.Main(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
