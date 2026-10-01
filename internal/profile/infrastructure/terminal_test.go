package infrastructure

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// statErrorStream — поток, о котором сведения получить не удалось.
type statErrorStream struct{}

// Stat отказывает, как отказал бы закрытый поток.
func (statErrorStream) Stat() (os.FileInfo, error) {
	return nil, errors.New("сведения о потоке недоступны")
}

func TestIsTerminalRejectsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ввод")

	if err := os.WriteFile(path, []byte("1\n"), 0o600); err != nil {
		t.Fatalf("подготовка файла вернула ошибку: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("открытие файла вернуло ошибку: %v", err)
	}

	t.Cleanup(func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Errorf("закрытие файла вернуло ошибку: %v", closeErr)
		}
	})

	if IsTerminal(file) {
		t.Error("файл признан терминалом")
	}
}

func TestIsTerminalRejectsPipe(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("создание канала вернуло ошибку: %v", err)
	}

	t.Cleanup(func() {
		if closeErr := reader.Close(); closeErr != nil {
			t.Errorf("закрытие канала вернуло ошибку: %v", closeErr)
		}

		if closeErr := writer.Close(); closeErr != nil {
			t.Errorf("закрытие канала вернуло ошибку: %v", closeErr)
		}
	})

	if IsTerminal(reader) {
		t.Error("канал признан терминалом")
	}
}

func TestIsTerminalRejectsStreamWithoutStat(t *testing.T) {
	if IsTerminal(statErrorStream{}) {
		t.Error("поток без сведений признан терминалом")
	}
}

func TestIsTerminalAcceptsCharDevice(t *testing.T) {
	device, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("открытие %s вернуло ошибку: %v", os.DevNull, err)
	}

	t.Cleanup(func() {
		if closeErr := device.Close(); closeErr != nil {
			t.Errorf("закрытие %s вернуло ошибку: %v", os.DevNull, closeErr)
		}
	})

	if !IsTerminal(device) {
		t.Errorf("символьное устройство %s не признано терминалом", os.DevNull)
	}
}
