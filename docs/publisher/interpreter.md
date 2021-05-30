# Publisher `interpreter`

Run specified interpreter with generated content

## Config

```yaml
bin: /path/to/some/executable
baseArgs: [some, args]
```

## Reference Commands Mapping

```yaml
/discuss:
  as: /prepare
  description: prepare script for interpreter execution

/end:
  as: /run
  description: run the prepared script

# disable following commands since they are not used

/edit: {}
/list: {}
/delete: {}
/start: {}
/continue: {}
/ignore: {}
/include: {}
```
