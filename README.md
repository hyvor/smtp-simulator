This is a simple SMTP simulator server for bounces and complaints. 

> For Hyvor Relay customers, the email domain is `simulator.relay.hyvor.com`. For example, to simulate a busy mailbox, send an email to `busy@simulator.relay.hyvor.com`.

## Installation

<!--  -->

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

### Complaints

These emails accept the message initially but later send a complaint notification back to the sender. The complaint is sent as per RFC5965.

| Email Local Part | Description                 |
| ---------------- | --------------------------- |
| `complaint@`     | Simulates a user complaint. |

### Custom Responses

Sometimes, you might need further customization of the server response. To simulate these scenarios, send emails to `custom@` and use the following headers to define the response:

```
X-Custom-Status-Code: <status-code>
X-Custom-Enhanced-Code: <enhanced-code>
X-Custom-Message: <custom-message>
X-Custom-Delay: <delay-in-seconds-for-async-responses>
```

It is also possible to extend the responses from any other email address with these custom headers.