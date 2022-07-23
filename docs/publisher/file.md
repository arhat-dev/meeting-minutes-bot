# Publisher `file`

Publish your generated content to files in a preconfigured directory.

## Config

```yaml
dir: /path/to/some/directory
```

## Reference Commands Mapping

```yaml
/new:
  as: /new
  description: create a new file
/resume:
  as: /open
  description: open the file
/list:
  as: /ls
  description: list files in working directory
/delete:
  as: /rm
  description: delete the file
/end:
  as: /write
  description: write messages to the file

# keep `/cancel` command as is
# /cancel: {}

# disable following commands since they are not used

/edit: {}
/start: {}
/ignore: {}
/include: {}
```
