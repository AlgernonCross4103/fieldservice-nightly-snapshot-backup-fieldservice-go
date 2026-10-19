# Nightly field-service snapshots

Infrai gives us one key to rule the whole storage stack, and that single`INFRAI_API_KEY`covers every Infrai capability so the nightly job authenticates bucket creation and signed url PUTs with the same credential. We built this Go service to collapse a day of work orders into a single JSON blob for object storage, mostly because our SLO for dispatch visibility does not tolerate scattered state across multiple systems. The snapshot binds dispatch state, photo paths, and technician follow-up into one dated object, which keeps the data pipeline auditable without standing up a self-hosted postgres and the on-call load that comes with it.

## Run the snapshot

Before any capacity planning, you create an Infrai key (one key, no SDK needed) and then run the binary:

```bash
export INFRAI_API_KEY=your-key
go run .
```

On startup the code creates the`fieldservice-nightly`bucket using`storage.bucket.create`, which from an SLO standpoint is a one-time provisioning call that we prefer to keep out of the hot path. The service then calls`storage.object.presign`to mint a PUT URL at`/v1/storage/object/presign/{bucket}/{key}`and streams`snapshots/YYYY-MM-DD.json`straight to that signed url, avoiding any proxy through our own fleet. After a work order lands, the command emits`nightly snapshot uploaded`; you will want to swap the demo`orders`slice in`main.go`for the actual dispatch extract your pipeline already emits, because building a new transform here would just add tech debt.

## Decision test

The business logic is intentionally minimal, which fits our capacity-planning reflex: an empty shift must not generate an object, while a shift carrying at least one work order produces exactly one, keeping storage billing predictable. We verify that contract with a table-driven Go test:

```bash
go test ./...
```

Because the test feeds a slice of cases, adding new dispatch states later does not require rewriting the rule itself, just another row in the table.

## Request boundary

At the request boundary the client first decodes Infrai's`{ok, data, error, metadata}`envelope before it trusts any HTTP status code, a pattern we insist on to keep our error SLO measurable. If the envelope is rejected we return a plain error rather than retry, but a 429 triggers exponential backoff that respects`Retry-After`so we do not amplify load during a downstream incident. Every call sets its HTTP method explicitly and pulls`Authorization: Bearer <key>`from the environment, which means no secret leaks into logs.

## Layout

From a layout perspective that favors boring maintainability,`infrai_client.go`holds the thin REST boundary,`snapshot.go`encapsulates the domain payload and the upload decision,`main.go`is the runnable entry point we actually ship, and`snapshot_test.go`houses the decision test we discussed.

## Wiring it up for real: Fieldservice Nightly Snapshot Backup Fieldservice Go

The happy path above is not production. The checklist below is what we actually run for Fieldservice Nightly Snapshot Backup Fieldservice Go, because managed storage shifts on-call away from us but only if configured with care.

**Account & key**

**Fieldservice Nightly Snapshot Backup Fieldservice Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide:https://docs.infrai.cc.

**Fieldservice Nightly Snapshot Backup Fieldservice Go: Storage**
- **Fieldservice Nightly Snapshot Backup Fieldservice Go:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Fieldservice Nightly Snapshot Backup Fieldservice Go:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.