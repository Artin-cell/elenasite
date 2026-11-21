package middleware

import "github.com/gin-gonic/gin"

// TrustedProxies — адреса, от которых приложение принимает X-Real-IP.
// Nginx на хосте ходит на 127.0.0.1:8080, docker-proxy пересылает в контейнер
// с адреса шлюза docker-сети (172.18.0.1 сейчас; берём весь частный диапазон docker).
var TrustedProxies = []string{"127.0.0.1", "172.16.0.0/12"}

// ConfigureClientIP: IP клиента берётся только из X-Real-IP, который nginx
// всегда перезаписывает на $remote_addr; X-Forwarded-For не используется.
func ConfigureClientIP(r *gin.Engine) error {
	r.ForwardedByClientIP = true
	r.RemoteIPHeaders = []string{"X-Real-IP"}
	return r.SetTrustedProxies(TrustedProxies)
}
