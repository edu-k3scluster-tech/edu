package pages

import (
	"edu-portal/app/server/middleware"
	"fmt"
	"net/http"
)

func (p Pages) Status(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.UserFromCtx(r.Context())
	if !exists {
		p.render500(w, fmt.Errorf("user not found"))
		return
	}
	certificate, err := p.Store.GetUserCertificate(r.Context(), user.Id)
	if err != nil {
		p.render500(w, err)
		return
	}

	var k8sconfig string

	if certificate != nil {
		k8sconfig, err = p.Cluster.Config(r.Context(), []byte(certificate.Certificate), []byte(certificate.PrivateKey))
		if err != nil {
			p.render500(w, err)
			return
		}
	}

	p.render(w, 200, "status.tmpl.html", "", map[string]interface{}{
		"user":           user,
		"hasIntegration": k8sconfig != "",
		"k8sconfig":      k8sconfig,
	})
}
