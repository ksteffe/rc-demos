# Request coordinator workload identity

This path gives the request-coordinator workload a stable source identity for
its RCProfile and Backstage Component. The executable is built from the shared
parent application and selected with SERVICE_ROLE=request.

Keeping workload source provenance separate from its Kubernetes Deployment and
Service names lets the adapter discover the source selector without assuming
that those names match.

