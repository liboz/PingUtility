#!/bin/bash
docker load -i container
docker run --rm -v /root/logs:/App/old-logs -e "TZ=`cat /etc/timezone`" -d --name pingutility -i pingutility