package domain

import (
	"errors"
	"testing"
)

func TestValidateProjectIdentifier(t *testing.T) {
	cases := []struct {
		name     string
		value    string
		accepted bool
	}{
		{name: "путь go-модуля", value: "github.com/acme/orders", accepted: true},
		{name: "одно слово", value: "orders", accepted: true},
		{name: "владелец и проект", value: "acme/orders", accepted: true},
		{name: "дефис и точка", value: "example.com/acme-orders", accepted: true},
		{name: "пустая строка", value: "", accepted: false},
		{name: "строка с пробелом", value: "acme orders", accepted: false},
		{name: "только пробел", value: " ", accepted: false},
		{name: "табуляция", value: "acme\torders", accepted: false},
		{name: "перевод строки", value: "acme\norders", accepted: false},
		{name: "управляющий символ", value: "acme\x00orders", accepted: false},
		{name: "ведущий и замыкающий слеш", value: "/x/", accepted: false},
		{name: "ведущий слеш", value: "/acme/orders", accepted: false},
		{name: "замыкающий слеш", value: "acme/orders/", accepted: false},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateProjectIdentifier(test.value)

			if test.accepted {
				if err != nil {
					t.Fatalf("идентификатор %q отвергнут: %v", test.value, err)
				}

				return
			}

			if err == nil {
				t.Fatalf("идентификатор %q принят, ожидался отказ", test.value)
			}

			if !errors.Is(err, ErrInvalidProjectIdentifier) {
				t.Errorf("ошибка %v не сравнима с ErrInvalidProjectIdentifier", err)
			}
		})
	}
}
