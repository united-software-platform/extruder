// Команда extruder: точка входа инструмента. Вся логика запуска живёт в точке сборки
// internal/cli — здесь остаётся только передача аргументов и кода возврата.
package main

import (
	"os"

	"github.com/united-software-platform/extruder/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
