This is a simple SMTP simulator server for bounces and complaints.

> For Hyvor Relay customers, the email domain is `simulator.relay.hyvor.com`. For example, to simulate a busy mailbox, send an email to `busy@simulator.relay.hyvor.com`.

## Installation

### Using Docker

`hyvor/smtp-simulator` is a minimal Docker image that exposes port `25` for SMTP.

```bash
docker run -p 25:25 hyvor/smtp-simulator
```

### Using Docker Compose

```yaml
services:
  smtp-simulator:
    image: hyvor/smtp-simulator
    ports:
      - "25:25"
```

### From Source

```bash
git clone https://github.com/hyvor/smtp-simulator
cd smtp-simulator
go build -o smtp-simulator .
./smtp-simulator
```

## Environment Variables

- `DOMAIN` (default: `localhost`): The domain name of the SMTP server. You should set up a MX record for this domain to point to the server's IP address. This is used in the `HELO`/`EHLO` command and in the `From` address of bounce emails.

## Email Addresses

Send emails to the following addresses to simulate different scenarios.

### Synchronous Responses

These emails respond with a bounce immediately within the SMTP transaction after the `DATA` command is completed.

| Email Local Part | Description                                            | Status Code | Enhanced Code |
| ---------------- | ------------------------------------------------------ | ----------- | ------------- |
| `accept@`        | Accepts the email and simulates a successful delivery. | 250         | 2.0.0         |
| `busy@`          | Simulates a busy mailbox.                              | 450         | 4.2.1         |
| `tempfail@`      | Simulates a temporary failure.                         | 451         | 4.3.0         |
| `missing@`       | Simulates a hard bounce.                               | 550         | 5.1.1         |
| `disabled@`      | Simulates a disabled email address.                    | 550         | 5.1.2         |
| `spam@`          | Simulates a spam rejection.                            | 550         | 5.7.1         |

### Asynchronous Bounces

These emails accept the message initially but later send a bounce notification (DSN) back to the sender. The bounce is sent as per RFC3464.

| Email Local Part  | Description                         | Status Code | Enhanced Code |
| ----------------- | ----------------------------------- | ----------- | ------------- |
| `missing+async@`  | Simulates a hard bounce.            | 550         | 5.1.1         |
| `disabled+async@` | Simulates a disabled email address. | 550         | 5.1.2         |
| `spam+async@`     | Simulates a spam rejection.         | 550         | 5.7.1         |

- If there are multiple recipients with different bounce types, a single DSN will be sent with all the failed recipients listed.

### At-Data Responses

These emails accept the message in the `RCPT` stage (return `250 OK`) but respond with the bounce error after the `DATA` command is completed.

| Email Local Part   | Description                         | Status Code | Enhanced Code |
| ------------------ | ----------------------------------- | ----------- | ------------- |
| `busy+atdata@`     | Simulates a busy mailbox.           | 450         | 4.2.1         |
| `tempfail+atdata@` | Simulates a temporary failure.      | 451         | 4.3.0         |
| `missing+atdata@`  | Simulates a hard bounce.            | 550         | 5.1.1         |
| `disabled+atdata@` | Simulates a disabled email address. | 550         | 5.1.2         |
| `spam+atdata@`     | Simulates a spam rejection.         | 550         | 5.7.1         |

### Complaints

These emails accept the message initially but later send a complaint notification back to the sender. The complaint is sent as per RFC5965.

| Email Local Part | Description                 |
| ---------------- | --------------------------- |
| `complaint@`     | Simulates a user complaint. |

### Custom Responses

For full control over the response, send to `custom+<hex>@` where `<hex>` is the hex-encoded JSON of a spec:

```json
{ "status": 550, "message": "5.7.1 bad reputation", "at": "rcpt" }
```

| Field     | Type   | Required | Description                                                                                           |
| --------- | ------ | -------- | ----------------------------------------------------------------------------------------------------- |
| `status`  | int    | yes      | SMTP status code to return (200–599).                                                                 |
| `message` | string | no       | Text emitted **verbatim** after the status code.                                                      |
| `at`      | string | no       | When to respond: `"rcpt"` (default) responds during `RCPT`; `"data"` accepts then fails after `DATA`. |

- The message is sent verbatim — the server does not inject an enhanced status code. To include one, put it in `message` (e.g. `"5.7.1 bad reputation"` → `550 5.7.1 bad reputation`); to omit it, leave it out (e.g. `"bad reputation"` → `550 bad reputation`).
- A `status` of `250` accepts the recipient.
- Malformed hex, malformed JSON, an out-of-range `status`, or an unknown `at` value are rejected with `501`.

#### Encoding the spec

```bash
printf '%s' '{"status":550,"message":"5.7.1 bad reputation","at":"rcpt"}' | xxd -p | tr -d '\n'
# 7b22737461747573223a3535302c226d657373616765223a22352e372e31206261642072657075746174696f6e222c226174223a2272637074227d
```

Then send to `custom+<that-hex>@simulator.relay.hyvor.com`, e.g. with [swaks](https://github.com/jetmore/swaks):

```bash
HEX=$(printf '%s' '{"status":550,"message":"5.7.1 bad reputation"}' | xxd -p | tr -d '\n')
swaks --server localhost:25 --from sender@example.com --to "custom+$HEX@localhost"
# -> 550 5.7.1 bad reputation
```
