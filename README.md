# WiserLogging
Logging from Wiser Heat Hub to AWS, via NodeMCU, Lambda, S3, Athena, openTSDB and Grafana.

## Setup
* `poll-hub.ino` sets up a NodeMCU to poll the Wiser hub and PUT the data to an AWS API wrapping a lambda service.
* `s3-put-lambda.go` is the code for the Lambda service.
* `analysis.sh` can be run to retrieve data and set up an openTSDB/grafana stack to analyse it.

## Analysis
Running `analysis.sh` will launch openTSDB and grafana in docker containers and then query Athena for the logged data. The values are then imported into openTSDB and defined as a datasource in grafana ready to access on localhost:80.
