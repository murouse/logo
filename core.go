package logo

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger создает сырой экземпляр *zap.Logger, оптимизированный под JSON или Console вывод в Stdout.
// Этот инстанс не используется напрямую в бизнес-логике, а передается как Backend для slog.
func NewZapLogger(level zapcore.Level, format Format) *zap.Logger {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder // Стандартизируем таймштампы (ISO8601)

	var encoder zapcore.Encoder
	switch format {
	case FormatJSON:
		encoder = zapcore.NewJSONEncoder(cfg)
	case FormatConsole:
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // Расцвечивает уровни (INFO, ERROR) в терминале

		wd, _ := os.Getwd()
		cfg.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			if !caller.Defined {
				enc.AppendString("undefined")
				return
			}

			// По умолчанию берем полный путь к файлу
			filePath := caller.File

			// Если удалось вычислить относительный путь от папки запуска (soma) — берем его
			if wd != "" {
				if rel, err := filepath.Rel(wd, caller.File); err == nil {
					filePath = rel
				}
			}

			// Форматируем как file.go:line
			enc.AppendString(fmt.Sprintf("%s:%d", filePath, caller.Line))
		}

		encoder = zapcore.NewConsoleEncoder(cfg)
	}

	// Собираем ядро с прямой записью в Stdout
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	return zap.New(core)
}
