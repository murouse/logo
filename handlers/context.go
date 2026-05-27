package handlers

import (
	"context"
	"log/slog"

	"github.com/murouse/logo/logctx"
	"github.com/samber/lo"
)

// ContextAttrsHandler автоматически обогащает каждую запись лога атрибутами,
// сохраненными ранее в контексте выполнения горутины.
type ContextAttrsHandler struct {
	slog.Handler
}

// NewContextAttrsHandler создает хендлер-обогатитель контекста.
func NewContextAttrsHandler(next slog.Handler) slog.Handler {
	return &ContextAttrsHandler{next}
}

// Handle извлекает изолированные атрибуты из context.Context, выполняет дедупликацию
// и слияние одноименных slog.KindGroup между контекстом и локальными аргументами записи,
// после чего конструирует новую slog.Record и передает ее дальше по цепочке хэндлеров.
//
// Если контекст не содержит атрибутов лога, запись передается ниже без изменений
// и дополнительных аллокаций.
func (h *ContextAttrsHandler) Handle(ctx context.Context, r slog.Record) error {
	ctxAttrs := logctx.AttrsFromContext(ctx)
	if len(ctxAttrs) == 0 {
		return h.Handler.Handle(ctx, r)
	}

	// Хранилище для разделения: мапа для групп, слайс для плоских атрибутов
	ctxGroups := make(map[string][]slog.Attr)

	// finalAttrs инициализируем с запасом под контекст + локальные атрибуты
	finalAttrs := make([]slog.Attr, 0, len(ctxAttrs)+r.NumAttrs())

	// За один проход распределяем атрибуты контекста
	for _, attr := range ctxAttrs {
		if attr.Value.Kind() == slog.KindGroup {
			ctxGroups[attr.Key] = attr.Value.Group()
		} else {
			finalAttrs = append(finalAttrs, attr)
		}
	}

	// Обрабатываем локальные атрибуты текущей строки лога
	r.Attrs(func(attr slog.Attr) bool {
		if attr.Value.Kind() == slog.KindGroup {
			// Если группа с таким именем уже была в контексте — мержим её содержимое
			if group, ok := ctxGroups[attr.Key]; ok {
				ctxGroups[attr.Key] = append(group, attr.Value.Group()...)
				return true
			}
		}
		// Плоские атрибуты и новые группы просто дописываем в финальный слайс
		finalAttrs = append(finalAttrs, attr)
		return true
	})

	// Возвращаем склеенные из контекста группы обратно в общий пул
	finalAttrs = append(finalAttrs, lo.MapToSlice(ctxGroups, func(key string, attrs []slog.Attr) slog.Attr {
		return slog.GroupAttrs(key, attrs...)
	})...)

	// Собираем чистую изолированную запись
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)
	newRecord.AddAttrs(finalAttrs...)

	return h.Handler.Handle(ctx, newRecord)
}

// WithAttrs возвращает новый экземпляр ContextAttrsHandler с предопределенными
// атрибутами, делегируя их сохранение нижележащему хэндлеру.
func (h *ContextAttrsHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextAttrsHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup возвращает новый экземпляр ContextAttrsHandler, внутри которого
// все последующие атрибуты логирования будут изолированы в рамках указанной группы (name).
func (h *ContextAttrsHandler) WithGroup(name string) slog.Handler {
	return &ContextAttrsHandler{Handler: h.Handler.WithGroup(name)}
}
