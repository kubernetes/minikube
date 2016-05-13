package azureutil

import (
	"net/http"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/docker/machine/drivers/azure/logutil"
	"github.com/docker/machine/libmachine/log"
)

func withInspection() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			log.Debug("Azure request", logutil.Fields{
				"method":  r.Method,
				"request": r.URL.String(),
			})
			return p.Prepare(r)
		})
	}
}

func byInspecting() autorest.RespondDecorator {
	return func(r autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(resp *http.Response) error {
			log.Debug("Azure response", logutil.Fields{
				"status":          resp.Status,
				"method":          resp.Request.Method,
				"request":         resp.Request.URL.String(),
				"x-ms-request-id": azure.ExtractRequestID(resp),
			})
			return r.Respond(resp)
		})
	}
}
