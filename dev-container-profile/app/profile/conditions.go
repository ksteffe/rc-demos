// Package profile holds declarations consumed statically by the Runtime
// Conditions profiler. These calls are never executed by the application.
package profile

import (
	common "github.com/runtimeconditions/extensions/catalog/rc/common-integrations/go"
	env "github.com/runtimeconditions/extensions/catalog/rc/env-configuration/go"
)

type Message struct {
	Message string `json:"message"`
}

func declaration() {
	if false {
		common.API("content-api",
			common.Spec("openapi", "catalog://api/default/demo-content", "1.0.0"),
			common.GET("/message", common.Response[Message]()),
			env.Env("baseUrl", "CONTENT_API_URL"),
		)
	}
}
