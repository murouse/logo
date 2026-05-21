package logo

import (
	"os"

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

		cfg.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
			// caller.File вернет полный путь, например: /Users/.../soma/entrypoint.go
			// Можно вырезать из него текущую рабочую директорию или имя модуля
			// Но самый простой способ — оставить стандартный trim, убедившись, что он не дублирует корень:
			enc.AppendString(caller.TrimmedPath())
		}

		encoder = zapcore.NewConsoleEncoder(cfg)
	}

	// Собираем ядро с прямой записью в Stdout
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	return zap.New(core)
}
