package middleware

import (
	"net/http"

	"github.com/javierg/hackathon-bqia/internal/infrastructure/config"
	"github.com/javierg/hackathon-bqia/internal/shared/response"
)

const LicenseKeyHeader = "X-License-Key"

func LicenseKey(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.LicenseKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			if r.Header.Get(LicenseKeyHeader) != cfg.LicenseKey {
				response.Error(w, http.StatusUnauthorized, "invalid or missing X-License-Key")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
