package pprof

import (
	"net/http"
	_ "net/http/pprof"
)

func InitPProf() {
	go func() {
		if err := http.ListenAndServe("0.0.0.0:0", nil); err != nil {
			panic(err)
		}
	}()
}
