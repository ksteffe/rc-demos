module github.com/runtimeconditions/rc-demos/dev-container-profile

go 1.25.0

require (
	github.com/runtimeconditions/extensions/catalog/rc/common-integrations/go v0.0.0
	github.com/runtimeconditions/extensions/catalog/rc/env-configuration/go v0.0.0
	github.com/runtimeconditions/go-rc-profiler v0.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace github.com/runtimeconditions/extensions/catalog/rc/common-integrations/go => ../../extensions/catalog/rc/common-integrations/go

replace github.com/runtimeconditions/extensions/catalog/rc/env-configuration/go => ../../extensions/catalog/rc/env-configuration/go

replace github.com/runtimeconditions/go-rc-profiler => ../../go-rc-profiler
