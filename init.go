package logo

import (
	"fmt"
	"log/slog"

	"github.com/murouse/logo/handlers"
	"go.uber.org/zap/exp/zapslog"
)

func init() {
	// Автоматическая базовая инициализация при импорте пакета.
	// Гарантирует, что slog не запаникует и будет писать в понятном формате даже до явного вызова Init().
	_ = Init()
}

// Init выполняет ручную оркестрацию: собирает низкоуровневый Zap, оборачивает его в zapslog.Handler,
// строит упорядоченный конвейер обработки и делает его дефолтным для всего Go-приложения.
func Init(opts ...Option) error {
	cfg, err := DefaultWith(opts...)
	if err != nil {
		return fmt.Errorf("default with: %w", err)
	}

	zapLogger := NewZapLogger(LevelToZapLevel(cfg.Level), cfg.Format) // Создаем производительный фундамент (Zap)

	baseHandler := zapslog.NewHandler(
		zapLogger.Core(),
		zapslog.WithCaller(cfg.WithCaller),
		zapslog.WithCallerSkip(cfg.CallerSkip),
	) // Создаем адаптер-мост из zap в стандартный интерфейс slog.Handler

	handler := slog.Handler(baseHandler) // Приведение к интерфейсу slog.Handler необходимо, чтобыMiddleware-обертки могли прозрачно мутировать типы

	// Если задано имя сервиса, пришиваем его к базовому хендлеру
	if cfg.ServiceName != nil {
		handler = handler.WithAttrs([]slog.Attr{Service(*cfg.ServiceName)})
	}

	// Собираем декораторы (Middleware) по принципу Матрешки (внутри -> наружу).
	// Порядок применения важен: ContextAttrsHandler должен отработать ДО BufferHandler,
	// чтобы буфер увидел уже обогащенные контекстом записи.
	for _, middleware := range []Middleware{
		handlers.NewContextAttrsHandler,
		handlers.NewBufferHandler,
	} {
		handler = middleware(handler)
	}

	// Инжектим собранный пайплайн в стандартную библиотеку Go в качестве глобального логгера
	slog.SetDefault(slog.New(handler))
	return nil
}

// Middleware описывает контракт обертки над хендлером slog, позволяя строить цепочки (Pipeline pattern).
type Middleware func(slog.Handler) slog.Handler
