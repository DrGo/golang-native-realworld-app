# golang-native-realworld-app

A native Go implementation of the realworld app API https://github.com/gothinkster/realworld/tree/master/api#profile

## Features
- designed to be simple to install (one binary) and maintain but not to scale.
- no external dependencies (except sqlite3 and a low-level Go interface to sqlite3 developed by David Crawshaw https://github.com/crawshaw/sqlite).
- hand-written routing using 

## Security
### User authentication
https://stackoverflow.com/questions/549/the-definitive-guide-to-form-based-website-authentication

## Optimizations
### Sqlite
- https://stackoverflow.com/questions/1711631/improve-insert-per-second-performance-of-sqlite
- ```explain query plan```https://sqlite.org/eqp.html 
- added indexes as needed
- typical article sql ```
  SELECT json_object('articles', json_group_array(
        json_object('id', a.id, 'slug', slug, 'title', title, 'description', description,
'body', body, 'favorited', favourited, 'favoritesCount', favouritesCount,
'createdAt', DateTime(createdAt, 'unixepoch'), 'updatedAt', DateTime(updatedAt, 'unixepoch'),
'tagList', COALESCE(json_extract(tags, '$'), json_array()) ,
'author', json_object('username', u.username, 'bio', u.bio, 'image', u.image))), 'articlesCount',
        COUNT(a.id))
as JSON
FROM (select * from Article LIMIT $limit OFFSET $offset) a
INNER JOIN User AS u ON a.author=u.id
OUTER LEFT JOIN (SELECT articleID,json_group_array(tag) as tags FROM Tag GROUP BY articleID) t on a.id = t.articleID
 WHERE 1=1 AND a.id IN (SELECT articleID FROM Tag WHERE tag='js') ORDER BY a.createdAt DESC;
```
- use following optimizations https://phiresky.github.io/blog/2020/sqlite-performance-tuning/
-- pragma synchronous = normal; this is a per-connection pragma see https://github.com/crawshaw/sqlite/issues/101
-- pragma temp_store = memory;
-- pragma mmap_size = 30000000000;
-- pragma page_size = 32768;

## Benchmarking
- install wrk https://github.com/wg/wrk/blob/master/INSTALL
- run ```make bench``` after running the app.
```
wrk -c 80 -d 5  http://localhost:8080/api/articles
```
- typical result (v 0.0.1) without race detector but in debug mode (logging etc.)
```
Running 5s test @ http://localhost:8080/api/articles
  2 threads and 80 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    19.81ms   11.69ms 143.12ms   95.09%
    Req/Sec     2.17k   472.98     2.47k    92.00%
  21597 requests in 5.00s, 31.55MB read
Requests/sec:   4315.59
Transfer/sec:      6.31MB
```

typical result (v 0.0.1) without race detector and after removing the DB interface but in debug mode (logging etc.)
```
Running 5s test @ http://localhost:8080/api/articles
  2 threads and 80 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    17.74ms   13.51ms 143.08ms   93.77%
    Req/Sec     2.55k   656.94     2.94k    90.00%
  25414 requests in 5.00s, 37.13MB read
Requests/sec:   5078.44
Transfer/sec:      7.42MB
```

as above but without debugging output
```
Running 5s test @ http://localhost:8080/api/articles
  2 threads and 80 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency    11.35ms    0.97ms  20.98ms   86.49%
    Req/Sec     3.53k   327.57     6.40k    97.03%
  35521 requests in 5.10s, 51.90MB read
Requests/sec:   6963.32
Transfer/sec:     10.17MB
```