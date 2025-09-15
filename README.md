This is a simple SMTP simulator server for bounces, complaints, etc.

### Installation

<!--  -->

### Email Addresses

You can use the following email addresses to simulate different scenarios:

| Email Local Part | Description                                      | Status Code | Enhanced Code |
|------------------|--------------------------------------------------|-------------|---------------|
| `accept@`         | Accepts the email and simulates a successful delivery. | 250         | 2.0.0         |
| `bounce@`         | Simulates a hard bounce.                          | 550         | 5.1.1         |