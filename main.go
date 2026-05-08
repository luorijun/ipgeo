package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/netip"
	"strings"

	"github.com/lionsoul2014/ip2region/binding/golang/service"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

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

	getClientIP := func(ip string) (*IpInfo, error) {
		result, err := server.SearchByStr(ip)
		if err != nil {
			slog.Debug(err.Error())
			return nil, errors.New("failed to search IP info")
		}

		infos := strings.Split(result, "|")
		if len(infos) < 4 {
			slog.Debug("invalid IP info format: " + result)
			return nil, errors.New("invalid IP info format")
		}

		return &IpInfo{
			Ip:     ip,
			Region: infos[0],
			Prov:   infos[1],
			City:   infos[2],
			Isp:    infos[3],
		}, nil
	}

	// 初始化 web 接口
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-Ip")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			addr, err := netip.ParseAddrPort(r.RemoteAddr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to parse client IP"))
				slog.Debug(err.Error())
				return
			}
			ip = addr.Addr().String()
		}

		geo, err := getClientIP(ip)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get IP info: " + err.Error()))
			return
		}

		bytes, err := json.Marshal(geo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal IP info"))
			slog.Debug(err.Error())
			return
		}

		w.Write(bytes)
	})

	http.HandleFunc("/{ip}", func(w http.ResponseWriter, r *http.Request) {
		ip := r.PathValue("ip")

		geo, err := getClientIP(ip)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to get IP info: " + err.Error()))
			return
		}

		bytes, err := json.Marshal(geo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal IP info"))
			slog.Debug(err.Error())
			return
		}

		w.Write(bytes)
	})

	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}
}

type IpInfo struct {
	Ip     string `json:"ip"`
	Region string `json:"region"`
	Prov   string `json:"prov"`
	City   string `json:"city"`
	Isp    string `json:"isp"`
}
