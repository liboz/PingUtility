# Parser

Reads log files from servers given in `remote_config.json` and uses it to update a sqlite database

## Commands

```
docker run -v $(pwd)/remote_config.json:/remote_config.json image_id
```

`remote_config.json` looks like

```
{
  "targets": [{ "name": "NAME", "url": "http://localhost:1234/logFile" }]
}
```
