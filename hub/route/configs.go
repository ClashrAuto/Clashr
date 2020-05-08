package route

import (
	"net/http"
	"path/filepath"

	"github.com/ClashrAuto/Clashr/hub/executor"
	"github.com/ClashrAuto/Clashr/log"
	P "github.com/ClashrAuto/Clashr/proxy"
	T "github.com/ClashrAuto/Clashr/tunnel"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

func configRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getConfigs)
	r.Put("/", updateConfigs)
	r.Patch("/", patchConfigs)
	return r
}

type configSchema struct {
	Port        *int          `json:"port"`
	SocksPort   *int          `json:"socks-port"`
	RedirPort   *int          `json:"redir-port"`
	AllowLan    *bool         `json:"allow-lan"`
	BindAddress *string       `json:"bind-address"`
	Mode        *T.Mode       `json:"mode"`
	LogLevel    *log.LogLevel `json:"log-level"`
}

func getConfigs(w http.ResponseWriter, r *http.Request) {
	general := executor.GetGeneral()
	render.JSON(w, r, general)
}

func pointerOrDefault(p *int, def int) int {
	if p != nil {
		return *p
	}

	return def
}

func patchConfigs(w http.ResponseWriter, r *http.Request) {
	general := &configSchema{}
	if err := render.DecodeJSON(r.Body, general); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrBadRequest)
		return
	}

	if general.AllowLan != nil {
		P.SetAllowLan(*general.AllowLan)
	}

	if general.BindAddress != nil {
		P.SetBindAddress(*general.BindAddress)
	}

	ports := P.GetPorts()
	_ = P.ReCreateHTTP(pointerOrDefault(general.Port, ports.Port))
	_ = P.ReCreateSocks(pointerOrDefault(general.SocksPort, ports.SocksPort))
	_ = P.ReCreateRedir(pointerOrDefault(general.RedirPort, ports.RedirPort))

	if general.Mode != nil {
		T.Instance().SetMode(*general.Mode)
	}

	if general.LogLevel != nil {
		log.SetLevel(*general.LogLevel)
	}

	render.NoContent(w, r)
}

type updateConfigRequest struct {
	Path string `json:"path"`
}

func updateConfigs(w http.ResponseWriter, r *http.Request) {
	req := updateConfigRequest{}
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, ErrBadRequest)
		return
	}

	if !filepath.IsAbs(req.Path) {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError("path is not a absoluted path"))
		return
	}

	force := r.URL.Query().Get("force") == "true"
	cfg, err := executor.ParseWithPath(req.Path)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(err.Error()))
		return
	}

	executor.ApplyConfig(cfg, force)
	render.NoContent(w, r)
}
