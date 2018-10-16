with room_state as (
  SELECT
  System.UnixTime as timestamp,
  row_number() over (partition by room_struct.Name order by System.UnixTime desc) as rowNum,
  roomstat[1].MeasuredHumidity as Humidity,
  room_struct.CalculatedTemperature as Temperature,
  room_struct.CurrentSetPoint as TargetTemperature,
  room_struct.PercentageDemand as PercentageDemand,
  room_struct.Name as roomName
  FROM "YOURDB"."YOURTABLE"
  CROSS JOIN UNNEST(room) as t(room_struct)
  WHERE cast(concat(year, '-', month, '-', day) as date) =
     current_date
 )
SELECT 
  roomName,
  date_diff('second', from_unixtime(timestamp, current_timezone()), now()) as lag, 
  TargetTemperature, Temperature, Humidity
  from room_state
  where rowNum = 1
