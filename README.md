This is a simple SMTP simulator server for bounces, complaints, etc.

## Installation

<!--  -->

### Email Addresses

Send emails to the following addresses to simulate different scenarios.

#### Accept

| Email Local Part | Description                                            | Status Code | Enhanced Code |
| ---------------- | ------------------------------------------------------ | ----------- | ------------- |
| `accept@`        | Accepts the email and simulates a successful delivery. | 250         | 2.0.0         |

#### Synchronous Bounces

These emails respond with a bounce immediately within the SMTP transaction (when the DATA command is completed).

| Email Local Part | Description                                                          | Status Code | Enhanced Code |
| ---------------- | -------------------------------------------------------------------- | ----------- | ------------- |
| `busy@`          | Simulates a busy mailbox.                                            | 450         | 4.2.1         |
| `tempfail@`      | Simulates a temporary failure.                                       | 451         | 4.3.0         |
| `missing@`       | Simulates a hard bounce.                                             | 550         | 5.1.1         |
| `disabled@`      | Simulates a disabled email address.                                  | 550         | 5.1.2         |
| `spam@`          | Simulates a spam rejection (usually due to infrastructure problems). | 550         | 5.7.1         |

#### Asynchronous Bounces

These emails accept the message initially but later send a bounce notification (DSN) back to the sender. The bounce is sent as per RFC3464.

| Email Local Part  | Description                         | Status Code | Enhanced Code |
| ----------------- | ----------------------------------- | ----------- | ------------- |
| `missing+async@`  | Simulates a hard bounce.            | 550         | 5.1.1         |
| `disabled+async@` | Simulates a disabled email address. | 550         | 5.1.2         |
| `spam+async@`     | Simulates a spam rejection.         | 550         | 5.7.1         |

#### Complaints

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