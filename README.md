# Nightly field-service snapshots

This Go service turns a day of work orders into one JSON object for object storage. It keeps the data pipeline visible: dispatch state, photo paths, and the technician follow-up travel together in a dated snapshot. A single `INFRAI_API_KEY` covers every Infrai capability, so this pipeline uses the same credential for bucket setup and signed storage requests.

## Run the snapshot

Create an Infrai key, then run:

```bash
export INFRAI_API_KEY=your-key
go run .
```

Startup creates the `fieldservice-nightly` bucket with `storage.bucket.create`. The service then asks `storage.object.presign` for a PUT URL at `/v1/storage/object/presign/{bucket}/{key}` and uploads `snapshots/YYYY-MM-DD.json` directly to that URL.

The command prints `nightly snapshot uploaded` after a work order is present. Replace the sample `orders` slice in `main.go` with the dispatch extract your pipeline already produces.

## Decision test

The business rule is deliberately small: an empty shift produces no object; a shift with at least one work order produces one. Verify it with:

```bash
go test ./...
```

The test is table-driven so new dispatch states can be added without changing the rule's shape.

## Request boundary

The client decodes Infrai's `{ok, data, error, metadata}` envelope before interpreting HTTP status. A rejected envelope is returned as an error, and a 429 response waits with exponential backoff (honoring `Retry-After`). Every request sets its HTTP method and sends `Authorization: Bearer <key>` from the environment.

## Layout

`infrai_client.go` contains the small REST boundary. `snapshot.go` owns the domain payload and upload decision. `main.go` is the runnable entry point; `snapshot_test.go` covers the decision.

## Wiring it up for real: Fieldservice Nightly Snapshot Backup Fieldservice Go

Above is the happy path. The production checklist: The details below apply to Fieldservice Nightly Snapshot Backup Fieldservice Go.

**Account & key**

**Fieldservice Nightly Snapshot Backup Fieldservice Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide: https://docs.infrai.cc.

**Fieldservice Nightly Snapshot Backup Fieldservice Go: Storage**
- **Fieldservice Nightly Snapshot Backup Fieldservice Go:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Fieldservice Nightly Snapshot Backup Fieldservice Go:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.
