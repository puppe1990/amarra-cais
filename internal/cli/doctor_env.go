package cli

import (
	"os"
	"path/filepath"

	"github.com/puppe1990/amarra-cais/pkg/cais/dotenv"
)

func isProduction(dir string) bool {
	return resolveEnvVar(dir, "ENV") == "production"
}

func resolveEnvVar(dir, key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	data, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		return ""
	}
	return dotenv.Parse(data)[key]
}

func checkAdminToken(dir string) doctorCheck {
	return checkRequiredProductionEnv(dir, "ADMIN_TOKEN")
}

func checkAppURL(dir string) doctorCheck {
	return checkRequiredProductionEnv(dir, "APP_URL")
}

func checkPort(dir string) doctorCheck {
	return checkRequiredProductionEnv(dir, "PORT")
}

func checkRequiredProductionEnv(dir, key string) doctorCheck {
	if resolveEnvVar(dir, key) != "" {
		return doctorCheck{Name: key, OK: true}
	}
	return doctorCheck{
		Name:     key,
		Optional: true,
		Detail:   "required when ENV=production",
		FixHint:  "set " + key + " in .env",
	}
}

func hasAuthHandler(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "internal/handlers/auth.go"))
	return err == nil
}

func checkSMTP(dir string) doctorCheck {
	if resolveEnvVar(dir, "SMTP_HOST") != "" && resolveEnvVar(dir, "SMTP_FROM") != "" {
		return doctorCheck{Name: "SMTP", OK: true}
	}
	return doctorCheck{
		Name:     "SMTP",
		Optional: true,
		Detail:   "password reset emails log to stdout without SMTP_HOST/SMTP_FROM",
		FixHint:  "set SMTP_HOST and SMTP_FROM in .env for outbound mail",
	}
}
