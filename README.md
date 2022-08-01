# PingUtility

Golang project that pings google.com, router, localhost, and a host of your choice.
Also runs a server that allows access to the logs generated.
Golang app that reads the log output into a database

# Database Schema

```
Create TABLE timeout_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
	location TEXT,
    timestamp DATETIME,
	hour_minute DATETIME,
	UNIQUE(name, location, timestamp)
);


CREATE INDEX idx_name_location
ON timeout_data (name, location);

CREATE INDEX idx_hour_minute_location
ON timeout_data (hour_minute, location);

CREATE INDEX idx_name_hour_minute_location
ON timeout_data (name, hour_minute, location);

explain query plan select strftime('%H:%M:%S', timestamp), count(*) from timeout_data
GROUP BY strftime('%H:%M:%S', timestamp)


EXPLAIN QUERY PLAN select * from timeout_data where name = 'google.com'

select hour_minute, location, avg(c) from (
 select hour_minute, location, strftime('%W',timestamp) AS weekofyear, count(*) as c from timeout_data
 --where location = 'Home-PC'
 GROUP BY hour_minute, location, strftime('%W',timestamp)
)
group by hour_minute, location

select hour_minute, location, avg(c) from (
 select hour_minute, location, strftime('%W',timestamp) AS weekofyear, count(*) as c from timeout_data
 --where location = 'Home-PC'
 GROUP BY hour_minute, location, strftime('%W',timestamp)
)
group by hour_minute, location
```

# TODO

- [x] Do DNS lookup more than just once at the beginning just in case
- [ ] Database schema info
