# Community resource fulfillment services

This is one Go module and one container image with five runtime roles. Sharing
the image keeps the lab quick to build; each Kubernetes Deployment selects its
role with SERVICE_ROLE.

| Role | Purpose | Durable interaction |
| --- | --- | --- |
| request | Coordinates a user's resource request | Stores requests in Postgres and caches lookup state in Redis |
| inventory | Reports stock and reserves it transactionally | Stores inventory and reservations in Postgres |
| fulfillment | Creates delivery work | Stores fulfillment tasks in Postgres |
| notification | Consumes fulfillment events | Reads NATS and records notifications in Postgres |
| directory | Unrelated API used as a policy negative control | None |

The workflow is useful without a special demo mode. Inventory is seeded
idempotently with blankets, water filters, and first-aid kits. Submitting a
request decrements stock, creates a reservation, stores the request, creates a
fulfillment task, caches its status, and publishes a best-effort event.

## Public request

~~~http
POST /requests
Content-Type: application/json

{
  "itemId": "blankets",
  "quantity": 2,
  "destination": "Shelter A"
}
~~~

A successful call returns HTTP 201 and a request ID. Fetch it with
GET /requests/{id}; cached responses include X-Data-Source: redis.

## Intentional policy probes

The inventory API also exposes POST /admin/items, but that operation is absent
from the RCProfile. The donor directory is deployed and cataloged, but no
condition asks for it. Together they prove that generated policy grants only
the declared path/method pairs and destinations.

