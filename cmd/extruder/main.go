// Package main — точка входа Extruder. На этом изменении инструмент только сообщает свою версию:
// опрос, конвейер генерации и запись начальной структуры приходят последующими изменениями.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

// version — версия инструмента. Файл версии и релизный сценарий появятся на этапе поставки,
// до тех пор строка фиксирована и означает невыпущенную версию.
const version = "0.0.0-dev"

// Version возвращает строку версии инструмента.
func Version() string {
	return version
}

// run печатает версию в переданный поток. Вынесена из main ради проверки теста: main
// завершает процесс, а run возвращает ошибку записи вызывающему.
func run(out io.Writer) error {
	_, err := fmt.Fprintln(out, Version())

	return err
}

func main() {
	if err := run(os.Stdout); err != nil {
		log.SetFlags(0)
		log.Fatalf("extruder: не удалось напечатать версию: %v", err)
	}
}
