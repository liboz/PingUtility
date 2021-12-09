# PingUtility

F# project on dotnet core that pings google.com, router, localhost, and a host of your choice
Python script that reads the log output into a database

# Commands

```
./build.bat
```

# Running remotely

Requires docker to run. Requires a local file with name `remote_config.txt` with a single line that looks like `user@host`

```
#upload container
./upload_container.sh

# remote
./run_pingutility.sh
```

Inspect remote logs

```
docker logs pingutility
```

timezone changing in remote host

```
sudo dpkg-reconfigure tzdata
```
