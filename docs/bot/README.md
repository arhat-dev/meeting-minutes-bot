# Bots

A `bot` is a component receiving messages and act automatically.

## Concepts

- `workflow`: a set of bot commands to complete certain task
- `botcmd`: predefined text message triggering bot actions

We define following botcmds:

- `/new`: request a new workflow run
- `/resume`: resume a previously requested workflow run
- `/ignore`: ignore certain message during the workflow run
- `/include`: include extra information during the workflow run
- `/edit`: edit generated content of previous workflow run
- `/list`: list workflow run results
- `/delete`: delete certain workflow run result
- `/end`: end current workflow run
- `/cancel`: cancel the workflow run request

## State Machine of botcmds

```mermaid
stateDiagram-v2
  discuss: botcmd `/new ...`
  continue: bot cmd `/resume ...`
  cancel: botcmd `/cancel`
  state "user login" as login
  state login_state <<choice>>

  [*] --> discuss
  discuss --> login: redirect to private chat
  [*] --> continue
  continue --> login: redirect to private chat
  login --> login_state
  login_state --> Success: re
  login_state --> Fail: re
  Fail --> login
  login_state --> cancel
  cancel --> [*]
```

`/new` can be treated as the entrance of the workflow, once received
