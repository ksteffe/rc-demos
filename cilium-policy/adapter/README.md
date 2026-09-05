# Contributor-owned adapter

Implement the proof-of-concept CLI in this directory. Everything outside this
directory is a fixture provided by the project.

## Required command

The language and internal design are intentionally open, but the first
milestone should expose a command equivalent to:

~~~text
rc-cilium render \
  --profile ../profiles/request-coordinator.yaml \
  --backstage-url http://localhost:7007 \
  --kube-context kind-rc-cilium \
  --output ../out/request-coordinator.cnp.yaml
~~~

## Required behavior

- Match API conditions by required OpenAPI method/path capabilities, not by
  condition name.
- Use Backstage relations to find the Component that provides a matching API.
- Use the Component's Kubernetes annotations and live cluster metadata to
  resolve the actual Service name and namespace.
- Resolve datastore and cache conditions from the supplied Backstage Resource
  capability annotations using the same provenance path.
- Render one CiliumNetworkPolicy in applications.
- Select only the request-coordinator workload.
- Use toServices for the discovered destinations.
- Include only the requested HTTP methods and paths at L7.
- Produce deterministic YAML and useful errors for no match or ambiguous match.

Do not hard-code the supplied Service names. Tests may rename them while
preserving Backstage provenance.

The complete lookup and output contract is in ../docs/adapter-contract.md.

## Acceptance checks

- The workflow succeeds after the rendered policy is applied.
- Requests to undeclared Services remain blocked.
- Undeclared methods and paths on an allowed Service remain blocked.
- Renaming an RCProfile condition does not change the selected provider.
