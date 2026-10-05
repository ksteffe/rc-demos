module github.com/runtimeconditions/rc-demos/apps/request-logger-http

go 1.25.0

require (
	github.com/runtimeconditions/extensions/catalog/rc/common-integrations/go v0.0.0
	github.com/runtimeconditions/extensions/catalog/rc/env-configuration/go v0.0.0
)

replace github.com/runtimeconditions/extensions/catalog/rc/common-integrations/go => ../../../extensions/catalog/rc/common-integrations/go

replace github.com/runtimeconditions/extensions/catalog/rc/env-configuration/go => ../../../extensions/catalog/rc/env-configuration/go
