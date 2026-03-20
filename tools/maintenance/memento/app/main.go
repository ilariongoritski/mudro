package app

import (
	"fmt"
	"os"

	"github.com/goritskimihail/mudro/internal/bot"
)

func Run() {
	r := bot.NewRunner()
	out, err := r.Memento()
	if len(out) > 0 {
		fmt.Println(string(out))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
