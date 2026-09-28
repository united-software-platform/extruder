// Команда extruder: точка входа инструмента. Вся логика запуска живёт в слое
// инфраструктуры модуля «профиль» — здесь остаётся только передача аргументов и кода возврата.
package main

import (
	"os"

	"github.com/united-software-platform/extruder/internal/profile/infra"
)

func main() {
	os.Exit(infra.Run(os.Args[1:], os.Stdout, os.Stderr))
}
