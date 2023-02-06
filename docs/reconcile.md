# Reconcile

Basic reconcile support exists for v1 via a command line tool called `reconcile`.

See task `reconcile:run` for details.

A `reconcile.json` file would look something like:
```json
{
    "credit": 1000,
    "split_transactions": [
        {
            "memo": "Agent Fee",
            "account_id": 1,
            "credit": -110
        },
        {
            "memo": "Agent Fee GST",
            "account_id": 2,
            "credit": -11
        }
    ]
}
