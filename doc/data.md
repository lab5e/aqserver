# data fields

- device_id - the unique ID of the device
- received_time - milliseconds since epoch (UNIX time in milliseconds)
- firmware_ver - firmware version
- boardtemp - circuit board temperature in C
- board_rel_hum - relative humidity measured at circiuit board
- lat - GPS latitude in radians
- lon - GPS longitude in radians
- no2_ppb - NO2 concentration in parts per billion
- o3_ppb - O3 concentration in parts per billion
- no_ppb - NO concentration in parts per billion
- pm1 - parts per million, 1 micron particles
- pm25 - parts per million, 2.5 micron particles
- pm10 - parts per million, 10 micron particles
- opctemp - temperature (C) inside particle sensor
- opchum - relative humidity inside particle sensor

Particle sensor used:
<https://www.alphasense.com/wp-content/uploads/2022/09/Alphasense_OPC-N3_datasheet.pdf>

AFE analogue frontent for gas sensors:
<https://www.alphasense.com/wp-content/uploads/2019/10/AFE.pdf>

Gas sensors

- NO-A4 <https://www.alphasense.com/wp-content/uploads/2019/09/NO-A4.pdf>
- OX-A431 <https://www.alphasense.com/wp-content/uploads/2019/09/OX-A431.pdf>
- NO2-A43F <https://www.alphasense.com/wp-content/uploads/2019/09/NO2-A43F.pdf>
