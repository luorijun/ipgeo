package main

import (
	"encoding/json"
	"net/http"
	"net/netip"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/service"
)

func main() {

	// 初始化查询服务
	v4config, err := service.NewV4Config(service.BufferCache, "data/ip2region_v4.xdb", 20)
	if err != nil {
		panic(err)
	}

	v6config, err := service.NewV6Config(service.BufferCache, "data/ip2region_v6.xdb", 20)
	if err != nil {
		panic(err)
	}

	server, err := service.NewIp2Region(v4config, v6config)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	// 初始化 web 接口
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		addr, err := netip.ParseAddrPort(r.RemoteAddr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Write([]byte(addr.Addr().String()))
	})

	http.HandleFunc("/{ip}", func(w http.ResponseWriter, r *http.Request) {
		ip := r.PathValue("ip")
		if ip == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result, err := server.SearchByStr(ip)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		infos := strings.Split(result, "|")
		if len(infos) < 4 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(map[string]string{
			"region": infos[0],
			"prov":   infos[1],
			"city":   infos[2],
			"isp":    infos[3],
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(bytes)
	})

	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}
}
