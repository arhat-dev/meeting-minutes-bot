# Publisher `interpreter`

Run specified interpreter with generated content

## Config

```yaml
bin: /path/to/some/executable
args:
- some
- --args
- with template {{- . -}} support
```

## Reference Commands Mapping

```yaml
/new:
  as: /prepare
  description: prepare script for interpreter execution

/end:
  as: /run
  description: run the prepared script

# keep `/cancel` command as is
# /cancel: {}

# disable following commands since they are not used

/edit: {}
/list: {}
/delete: {}
/start: {}
/resume: {}
/ignore: {}
/include: {}
```
