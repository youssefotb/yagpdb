package roletemplates

import (
	"net/http"

	"emperror.dev/errors"
	"github.com/jonas747/yagpdb/web"
	"goji.io"
	"goji.io/pat"
)

func (p *Plugin) InitWeb() {
	web.LoadHTMLTemplate("../../streaming/assets/roletemplates.html", "templates/plugins/roletemplates.html")

	mux := goji.SubMux()
	web.CPMux.Handle(pat.New("/roles/*"), mux)
	web.CPMux.Handle(pat.New("/roles"), mux)

	// Alll handlers here require guild channels present

	mainPageHandler := web.ControllerHandler(p.handleGetMainPage, "cp_roletemplates")

	// Get just renders the template, so let the renderhandler do all the work
	mux.Handle(pat.Get(""), mainPageHandler)
	mux.Handle(pat.Get("/"), mainPageHandler)

}

func (p *Plugin) handleGetMainPage(w http.ResponseWriter, r *http.Request) (web.TemplateData, error) {
	g, tmpl := web.GetBaseCPContextData(r.Context())
	roleTemplates, err := GetGuildRoleTemplates(g.ID)
	if err != nil {
		return tmpl, errors.WithStackIf(err)
	}

	tmpl["RoleTemplates"] = roleTemplates

	return tmpl, nil
}
