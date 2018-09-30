with room_state as (
  SELECT
  System.UnixTime as timestamp,
  room_struct.CalculatedTemperature as Temperature,
  room_struct.CurrentSetPoint as TargetTemperature,
  room_struct.PercentageDemand as PercentageDemand,
  'room=' || room_struct.Name as room_tag
  FROM "your-db"."your-table"
  CROSS JOIN UNNEST(room) as t(room_struct)
)
SELECT metric, timestamp, AVG(value) as value, room_tag FROM
(
(
  SELECT 
  'BoilerState' as metric,
  System.UnixTime as timestamp,
  case when heatingchannel[1].HeatingRelayState = 'On' then 1 else 0 end as value,
  'boiler=1' as room_tag
  FROM "your-db"."your-table"
)
UNION
(
  SELECT 
  'Humidity' as metric,
  System.UnixTime as timestamp,
  roomstat[1].MeasuredHumidity as value,
  'room=Hallway' as room_tag
  FROM "your-db"."your-table"
)
UNION
(
  SELECT
  'Temperature' as metric,
  timestamp,
  Temperature as value,
  room_tag
  from room_state
)
UNION
(
  SELECT
  'TargetTemperature' as metric,
  timestamp,
  TargetTemperature as value,
  room_tag
  from room_state
)
UNION
(
  SELECT
  'PercentageDemand' as metric,
  timestamp,
  PercentageDemand as value,
  room_tag
  from room_state
)
)
GROUP BY metric, timestamp, room_tag
ORDER BY timestamp asc
