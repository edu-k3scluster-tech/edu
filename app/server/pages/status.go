package pages

import (
	"edu-portal/app/server/middleware"
	"edu-portal/app/server/utils"
	"fmt"
	"net/http"
)

func (p Pages) Status(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.UserFromCtx(r.Context())
	if !exists {
		utils.Render500(w, fmt.Errorf("user not found"))
		return
	}
	certificate, err := p.Store.GetUserCertificate(r.Context(), user.Id)
	if err != nil {
		utils.Render500(w, err)
		return
	}

	var k8sconfig string

	if certificate != nil {
		rawCertificate, err := certificate.GetCertificate()
		if err != nil {
			utils.Render500(w, err)
			return
		}
		privateKey, err := certificate.GetPrivateKey()
		if err != nil {
			utils.Render500(w, err)
			return
		}
		k8sconfig, err = p.Cluster.Config(r.Context(), []byte(rawCertificate), privateKey)
		if err != nil {
			utils.Render500(w, err)
			return
		}
	}

	p.render(w, 200, "status.tmpl.html", "", map[string]interface{}{
		"user":           user,
		"hasIntegration": k8sconfig != "",
		"k8sconfig":      k8sconfig,
	})
}
