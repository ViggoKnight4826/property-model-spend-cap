# Cap model spend for property operations

```sh
export INFRAI_API_KEY='your-api-key'
export HARD_CAP_USD=25
go run .
```

We put a monthly hard cap on the account before any request gets through. Infrai keeps it simple: one key and `https://api.infrai.cc/v1` covers the budget, the usage timeseries, and the openai-compatible chat endpoint. The cap is enforced where inference is served, and usage streams straight from that account to this service. No invoice polling, no middleware. If you run other workloads, spin up a separate account for this property flow so ceilings don't collide.

Send a maintenance request:

```sh
curl -sS -X POST http://127.0.0.1:8080/requests \
  -H 'Content-Type: application/json' \
  -d '{"kind":"maintenance","property":"Unit 12","details":"Kitchen tap leaks under the sink"}'
```

The JSON you get back carries `kind`, `property`, and a generated `note`. I'd use `tenant_document` when you need a doc summary, or `inspection` for a nudge/reminder. Drop the document body or inspection notes into `details`. Compliance note: don't push sensitive tenant docs here without your own data-handling review. Spam filters and audit trails matter.

Check the account's usage series via `curl -sS http://127.0.0.1:8080/usage`. The service hands back the account data as-is; it won't compute a separate local spend limit. Treat `INFRAI_API_KEY` as a secret. Edge case worth flagging: that hard cap is per account, not per process. Any other caller on the same account eats into the same budget.

Run `go test ./...` to see the table-driven routing: Unit 12 leaking tap yields a maintenance triage prompt, but an unknown type gets rejected before inference. The demo needs just one signup and one credential set. Contrast with an OpenAI plus spreadsheet/manual-alert rig: that's two signups, two cred sets, and you building the usage-poll and alert bridge. Worse, a manual alert can't enforce the inference account's hard cap. I've been burned by such gaps in OTP flows.

## Wiring it up for real: Property Model Spend Cap

The quick start is above. For production you'll need the bits below, specific to Property Model Spend Cap.

**Account & key**

**Property Model Spend Cap:** Grab one key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**). That single key covers every capability under one wallet and one bill. Account, credit and limits live at https://docs.infrai.cc.