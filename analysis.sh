docker-compose up -d
opentsdb_id=`docker-compose ps -q db`
sleep 10
curl 'http://admin:admin@localhost/api/datasources' -X POST -H 'Content-Type: application/json;charset=UTF-8' --data-binary '{"name":"localopenTSDB","type":"opentsdb","url":"http://db:4242","access":"proxy","isDefault":true,"jsonData":{"tsdbVersion":3}}'

echo 'Fetching data from Athena'
go run queryAthena.go

docker cp wiser.tsdb $opentsdb_id:/
docker exec $opentsdb_id /usr/local/bin/tsdb import /wiser.tsdb 
