# Site Subscribe

Site Subscribe simplifies monitoring static websites and emails you when changes are made.

## Requirements

A SendGrid API key.

## Configuration

First add a user.

```
> site-subscribe new-user -name "User Name" -email "user@email.com"
```

Then, subscribe to a site.

```
> site-subscribe sub -name "some site" -url "https://school.edu/professor/class/homework.html" -user "user@email.com"
```

Now, simply run the `update` command to check for updates.

```
> site-subscribe update -key "<YOUR SENDGRID API KEY>"
```

The update command should be scheduled to run at a regular interval using cron.
