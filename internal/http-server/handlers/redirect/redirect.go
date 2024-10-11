package redirect

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlgetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.redirect.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("field alias is empty")
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		resURL, err := urlgetter.GetURL(alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLNotFound) {
				log.Info("url not found", slog.Any("alias", alias))
				render.JSON(w, r, resp.Error("not found"))
				return
			} else {
				log.Error("failed to get url")
				render.JSON(w, r, resp.Error("internal error"))
			}

		}

		log.Info("get url", slog.Any("url", resURL))
		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
