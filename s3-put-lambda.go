package main

import (
 "log"
 "github.com/aws/aws-lambda-go/events"
 "github.com/aws/aws-lambda-go/lambda"
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/s3/s3manager"
  "fmt"
  "os"
  "strconv"
  "time"
)

func exitErrorf(msg string, args ...interface{}) {
  fmt.Fprintf(os.Stderr, msg+"\n", args...)
  os.Exit(1)
}

// Handler is your Lambda function handler
// It uses Amazon API Gateway request/responses provided by the aws-lambda-go/events package,
// However you could use other event sources (S3, Kinesis etc), or JSON-decoded primitive types such as 'string'.
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
 bucket := "your-bucket-here"
 now := time.Now().UTC()
 filename := strconv.FormatInt(now.UnixNano(), 10) + ".json"
 year, month, day := now.Date()
 partitions := fmt.Sprintf("year=%d/month=%d/day=%d/",year, month, day)
 jsonLength := len(request.Body);
 // stdout and stderr are sent to AWS CloudWatch Logs
 log.Printf("Processing Lambda request %s, %d bytes\n", request.RequestContext.RequestID, jsonLength)
 if jsonLength == 0 {
  exitErrorf("No JSON data to store")
 }
 file, err := os.Create("/tmp/" + filename)
 if err != nil {
  exitErrorf("Unable to open file $q, %v", err)
 }
 defer file.Close()
 file.WriteString(request.Body)
 file.Sync()
 _, err = file.Seek(0, 0)
 if err != nil {
  exitErrorf("Unable to seek to start of file", err)
 }
 sess, err := session.NewSession()
 uploader := s3manager.NewUploader(sess)

 _, err = uploader.Upload(&s3manager.UploadInput{
     Bucket: aws.String(bucket),
     Key: aws.String(partitions + filename),
     Body: file,
 })
 if err != nil {
     // Print the error and exit.
     exitErrorf("Unable to upload %q to %q, %v", filename, bucket, err)
 }

 fmt.Printf("Successfully uploaded %q to %q\n", filename, bucket)
 return events.APIGatewayProxyResponse{

  Body:       filename,
  StatusCode: 200,
 }, nil

}

func main() {
 lambda.Start(Handler)
}

