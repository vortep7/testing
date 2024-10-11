package save

import (
	"errors"
	"log/slog"
	"net/http"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias"`
}

type UniqueURLSaver interface {
	SaveURL(urlName string, alias string) (int64, error)
	AliasChecker(alias string) (bool, error)
}

const (
	aliasLength = 6
)

func New(log *slog.Logger, uniqueURLSaver UniqueURLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			errs := err.(validator.ValidationErrors)
			log.Error("invalid_request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(errs))
			return

		}

		alias := req.Alias

		if alias == "" {
			alias = random.NewRandomString(aliasLength)
			exists, err := uniqueURLSaver.AliasChecker(alias)
			if err != nil {
				log.Error("cannot be connected to storage", sl.Err(err))
				render.JSON(w, r, resp.Error("cannot be connected to storage"))
				return
			}
			if exists {
				log.Error("not unique alias", slog.String("alias", alias))
				render.JSON(w, r, resp.Error("not unique alias"))
				return
			}
		}

		id, err := uniqueURLSaver.SaveURL(req.URL, alias)

		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				log.Info("url already exist", slog.String("url", req.URL))

				render.JSON(w, r, resp.Error("url already exist"))

				return
			} else {
				log.Error("failed to add URL", sl.Err(err))
				render.JSON(w, r, resp.Error("failed to add URL"))
				return
			}
		}

		log.Info("url added", slog.Any("id", id))

		render.JSON(w, r, Response{
			Response: resp.OK(),
			Alias:    alias,
		})
	}

}
