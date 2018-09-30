#include <ESP8266WiFi.h>
#include <WiFiClientSecure.h>
#include <ESP8266HTTPClient.h>

#define JSONBUFLEN 9000
#define DEBUG false

char ssid[] = "WIFI SSID";
char password[] = "your-pasword";

char lambda_endpoint[] = "api-endpoint-here";
char lambda_url[] = "/default/wiser_etl";
char lambda_api_key[] = "key-here";

char wiser_endpoint[] = "192.168.0.17";
char wiser_url[]= "/data/domain/";
char wiser_secret[] = "secret-here";

char wiserJSON[JSONBUFLEN] = "";
HTTPClient http;
unsigned long previousMillis = 0;

void setup() {
  Serial.begin(115200);
  delay(10);
  Serial.setDebugOutput(true);
  Serial.print("Connecting Wifi: ");
  Serial.println(ssid);

  delay(1000);
  WiFi.begin(ssid, password);
  delay(1000);
  while (WiFi.status() != WL_CONNECTED)
  {
    Serial.print(".");
    delay(500);
  }
  Serial.println("");
  Serial.println("WiFi connected");
  Serial.println("IP address: ");
  IPAddress ip = WiFi.localIP();
  Serial.println(ip);
  
}

String getHTTPHeaders(WiFiClientSecure client){
  String headers;
  while (client.connected()) {
    String line = client.readStringUntil('\n');
    if (line == "\r") {
      Serial.println("headers received");
      return headers;
    }
    else headers += line;
  }
}

String getHTTPChunk(WiFiClientSecure client){
  String chunk = "";
  int chunksize = 0;
  if (client.connected()) {
    String line = client.readStringUntil('\n');
    if (line == "\r") {
      chunksize = 0;
    }
    else if (line.length() > 0)
    {
      int lastIndex = line.length() - 1;
      line.remove(lastIndex);
      //Serial.printf("Parsing chunk size: %s\n", line.c_str());
      chunksize = strtol(line.c_str(),NULL,16);
    }
  }
  //Serial.printf("Reading chunk of size %d\n", chunksize);

  while (client.connected() && chunk.length() < chunksize){
    String line = client.readStringUntil('\n');
    int lastIndex = line.length() - 1;
    line.remove(lastIndex);
    chunk += line;
  }

  return chunk;
}

String pushJSONtoS3(String jsonBlob)
{
  String headers = "";
  String body = "";
  bool finishedHeaders = false;
  bool currentLineIsBlank = true;
  bool gotResponse = false;
  WiFiClientSecure client;

  if (client.connect(lambda_endpoint, 443))
  {
    Serial.println("Connected to Lambda");  
    client.print("POST ");
    client.print(lambda_url);
    client.println(" HTTP/1.1");
    client.print("Host: ");
    client.println(lambda_endpoint);
    client.print("x-api-key: ");
    client.println(lambda_api_key);
    client.println("User-Agent: NodeMCU/1.0");
    client.print("content-length: ");
    client.println(jsonBlob.length());
    client.println("");
    client.print(jsonBlob);

    while (!client.available()) {
      delay(100);
    }
    
    while (client.available())
    {
      char c = client.read();
      if (finishedHeaders)
      {
        body = body + c;
      }
      else
      {
        if (currentLineIsBlank && c == '\n')
        {
          finishedHeaders = true;
        }
        else
        {
          headers = headers + c;
        }
      }
      
      if (c == '\n')
      {
        currentLineIsBlank = true;
      }
      else if (c != '\r')
      {
        currentLineIsBlank = false;
      }
      gotResponse = true; 
    }
  }
  if (gotResponse) return body;
}

void getWiserJSON()
{
  String chunk = "";
  String headers = "";
  bool finishedHeaders = false;
  bool currentLineIsBlank = true;
  bool gotResponse = false;
  bool jsonStart = false;
  long now;
  int bytes_to_copy = 0;
  
  wiserJSON[0] = 0;
  WiFiClientSecure client;

  if (client.connect(wiser_endpoint, 443))
  {
    Serial.println("Connected to Wiser");
    client.print("GET ");
    client.print(wiser_url);
    client.println(" HTTP/1.1");
    client.print("Host: ");
    client.println(wiser_endpoint);
    client.print("SECRET: ");
    client.println(wiser_secret);
    client.println("User-Agent: NodeMCU/1.0");
    client.println("");

    while (!client.available()) {
      delay(100);
    }

    headers = getHTTPHeaders(client);

    chunk = getHTTPChunk(client);
    while (chunk.length() > 0) {
      bytes_to_copy = min(JSONBUFLEN - strlen(wiserJSON), chunk.length());
      strncat(wiserJSON, chunk.c_str(), bytes_to_copy);
      chunk = getHTTPChunk(client);
    }
  }
}

void loop() {
  unsigned long currentMillis = millis();
  
  if (currentMillis - previousMillis >= 60000) {
    previousMillis = currentMillis;
    getWiserJSON();
    Serial.printf("Received %d bytes from Wiser\n", strlen(wiserJSON));
    if (strlen(wiserJSON) > 0) 
    {
      pushJSONtoS3(wiserJSON);
      Serial.println("Pushed JSON to s3");
    }
  }
}
