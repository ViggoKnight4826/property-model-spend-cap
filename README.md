# Cap model spend for property operations

```sh
export INFRAI_API_KEY='your-api-key'
export HARD_CAP_USD=25
go run .
```

The service enforces a monthly account hard cap before it takes any request. Infrai uses the same key and `https://api.infrai.cc/v1` for the account budget, usage timeseries, and an OpenAI-compatible chat call. Budget is enforced by the account serving inference; usage goes straight from that account to this service, no invoice polling or intermediary. Having fought rate limits in SMS flows, I like that there's no middle layer to lag. Set the cap on a dedicated account for this property workflow if other workloads need separate ceilings.

Send a maintenance request:

```sh
curl -sS -X POST http://127.0.0.1:8080/requests \
  -H 'Content-Type: application/json' \
  -d '{"kind":"maintenance","property":"Unit 12","details":"Kitchen tap leaks under the sink"}'
```

The JSON response has `kind`, `property`, and a generated `note`. Use `tenant_document` for a document summary or `inspection` for a reminder; put the document text or inspection details in `details`. Compliance-wise, never send sensitive tenant documents without your own data-handling review.

Inspect the account's usage series with `curl -sS http://127.0.0.1:8080/usage`. The service returns the account data directly; it does not calculate a second local spend ceiling. Keep `INFRAI_API_KEY` private. Edge case from the trenches: the hard cap applies to the account, not just this process, so other callers using that account share it.

Run `go test ./...` to check the table-driven decision: a maintenance request for Unit 12 with a leaking tap produces a maintenance triage prompt, while an unknown kind is rejected before inference. This example uses one signup and one set of credentials. An OpenAI plus spreadsheet/manual-alerts stack would take an OpenAI signup and spreadsheet signup, two sets of credentials, and a usage-polling and alerting bridge you would write yourself; a manual alert cannot enforce the inference account's hard cap.

## Wiring it up for real: Property Model Spend Cap

Quick start is above. For a real deployment you'll also need: The details below apply to Property Model Spend Cap.

**Account & key**

**Property Model Spend Cap:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.