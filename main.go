// file: buzz/main.go

package main

import (
	"net/http"

	"github.com/rskv-p/buzz/mod/m_bus"
	"github.com/rskv-p/buzz/pkg/x_init"
	"github.com/rskv-p/buzz/pkg/x_log"
)

func main() {

	x_init.Init()
	// Инициализация шины и модуля
	busModule := m_bus.NewBusModule("bus", "secretKey", nil, 10)

	// После создания модуля регистрируем публичные действия
	m_bus.RegisterPublicActions(&busModule.Module)

	// Запуск HTTP сервера
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			x_log.Error("Не удалось запустить HTTP сервер:", err)
		}
	}()

	x_log.Info("Сервер запущен на порту 8080")
	select {}

}
