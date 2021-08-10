package config

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	conf := GetConfig()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	tk := time.NewTicker(time.Second)
LOOP:
	for {
		select {
		case <-tk.C:
			t.Logf("conf=%+v\n", conf)
			t.Logf("app=%+v\n", conf.App)
			t.Logf("server=%+v\n", conf.Server)
			t.Logf("databases=%+v\n", conf.Databases)
			t.Logf("database=%+v\n", conf.Databases["readfog"])
			t.Logf("caches=%+v\n", conf.Caches)
			t.Logf("jwt.method=%+v\n", conf.Get("jwt.method"))
			t.Logf("rf_var1=%v\n", conf.Get("var1"))
			t.Logf("rf_var_p=%v\n", conf.Get("var_p"))

		case <-signalChan:
			break LOOP
		}
	}
}
