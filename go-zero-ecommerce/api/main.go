package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

// ───────────── Config ─────────────

type Config struct {
	rest.RestConf
	Upstreams map[string]string // path_prefix → target_url
}

// ───────────── Main ─────────────

func main() {
	var c Config
	conf.MustLoad("etc/config.yaml", &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// ── 反向代理: 按路径前缀转发到对应微服务 ──
	// 例如: /api/user/* → http://localhost:8001
	for prefix, target := range c.Upstreams {
		targetURL, err := url.Parse(target)
		if err != nil {
			log.Fatalf("无效的 upstream 地址 %s: %v", target, err)
		}
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		wildcardPath := prefix + "/*"

		// GET + POST + PUT + DELETE 全部转发
		for _, method := range []string{
			http.MethodGet, http.MethodPost,
			http.MethodPut, http.MethodDelete,
		} {
			server.AddRoutes([]rest.Route{
				{Method: method, Path: wildcardPath, Handler: proxyHandler(proxy)},
			})
		}

		log.Printf("路由: %s/* → %s\n", prefix, target)
	}

	log.Printf("🚀 API 网关启动: :%d\n", c.Port)
	server.Start()
}

func proxyHandler(proxy *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}
