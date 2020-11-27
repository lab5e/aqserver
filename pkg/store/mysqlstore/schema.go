package mysqlstore

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

const schema = `
CREATE TABLE IF NOT EXISTS messages (
  id             BIGINT PRIMARY KEY auto_increment,
  device_id      VARCHAR(255) NOT NULL,
  received_time  BIGINT NOT NULL,
  packetsize     INTEGER NOT NULL,
  sysid          BIGINT NOT NULL,
  firmware_ver   BIGINT NOT NULL,
  uptime         BIGINT NOT NULL,
  boardtemp      REAL NOT NULL,
  board_rel_hum  REAL NOT NULL,
  status         BIGINT NOT NULL,

  gpstimestamp  REAL NOT NULL,
  lon           REAL NOT NULL,
  lat           REAL NOT NULL,
  alt           REAL NOT NULL,

  sensor1work   INTEGER NOT NULL,
  sensor1aux    INTEGER NOT NULL,
  sensor2work   INTEGER NOT NULL,
  sensor2aux    INTEGER NOT NULL,
  sensor3work   INTEGER NOT NULL,
  sensor3aux    INTEGER NOT NULL,
  afe3_temp_raw INTEGER NOT NULL,

  no2_ppb         DOUBLE NOT NULL,
  o3_ppb          DOUBLE NOT NULL,
  no_ppb          DOUBLE NOT NULL,
  afe3_temp_value DOUBLE NOT NULL,

  opcpma        INTEGER NOT NULL,
  opcpmb        INTEGER NOT NULL,
  opcpmc        INTEGER NOT NULL,

  pm1               REAL NOT NULL,
  pm10              REAL NOT NULL,
  pm25              REAL NOT NULL,

  opcsampleperiod   INTEGER NOT NULL,
  opcsampleflowrate INTEGER NOT NULL,
  opctemp           INTEGER NOT NULL,
  opchum            INTEGER NOT NULL,
  opcfanrevcount    INTEGER NOT NULL,
  opclaserstatus    INTEGER NOT NULL,

  opcbin_0          INTEGER NOT NULL,
  opcbin_1          INTEGER NOT NULL,
  opcbin_2          INTEGER NOT NULL,
  opcbin_3          INTEGER NOT NULL,
  opcbin_4          INTEGER NOT NULL,
  opcbin_5          INTEGER NOT NULL,
  opcbin_6          INTEGER NOT NULL,
  opcbin_7          INTEGER NOT NULL,
  opcbin_8          INTEGER NOT NULL,
  opcbin_9          INTEGER NOT NULL,
  opcbin_10         INTEGER NOT NULL,
  opcbin_11         INTEGER NOT NULL,
  opcbin_12         INTEGER NOT NULL,
  opcbin_13         INTEGER NOT NULL,
  opcbin_14         INTEGER NOT NULL,
  opcbin_15         INTEGER NOT NULL,
  opcbin_16         INTEGER NOT NULL,
  opcbin_17         INTEGER NOT NULL,
  opcbin_18         INTEGER NOT NULL,
  opcbin_19         INTEGER NOT NULL,
  opcbin_20         INTEGER NOT NULL,
  opcbin_21         INTEGER NOT NULL,
  opcbin_22         INTEGER NOT NULL,
  opcbin_23         INTEGER NOT NULL,
  opcsamplevalid    INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS cal (
  id                    INTEGER PRIMARY KEY auto_increment,
  device_id             VARCHAR(255) NOT NULL,
  sysid                 BIGINT NOT NULL,
  collection_id         VARCHAR(255) NOT NULL,
  valid_from            DATETIME NOT NULL,

  afe_serial            VARCHAR(255) NOT NULL,

  circuit_type          VARCHAR(255) NOT NULL,
  afe_type              VARCHAR(255) NOT NULL,
  sensor1_serial        VARCHAR(255) NOT NULL,
  sensor2_serial        VARCHAR(255) NOT NULL,
  sensor3_serial        VARCHAR(255) NOT NULL,

  afe_cal_date          DATETIME NOT NULL,

  vt20_offset            REAL NOT NULL,

  sensor1_we_e           REAL NOT NULL,
  sensor1_we_0           REAL NOT NULL,
  sensor1_ae_e           REAL NOT NULL,
  sensor1_ae_0           REAL NOT NULL,
  sensor1_pcb_gain       REAL NOT NULL,
  sensor1_we_sensitivity REAL NOT NULL,

  sensor2_we_e           REAL NOT NULL,
  sensor2_we_0           REAL NOT NULL,
  sensor2_ae_e           REAL NOT NULL,
  sensor2_ae_0           REAL NOT NULL,
  sensor2_pcb_gain       REAL NOT NULL,
  sensor2_we_sensitivity REAL NOT NULL,

  sensor3_we_e           REAL NOT NULL,
  sensor3_we_0           REAL NOT NULL,
  sensor3_ae_e           REAL NOT NULL,
  sensor3_ae_0           REAL NOT NULL,
  sensor3_pcb_gain       REAL NOT NULL,
  sensor3_we_sensitivity REAL NOT NULL,

  UNIQUE(device_id, collection_id, afe_serial, valid_from)
);
`

func createSchema(db *sqlx.DB) {
	for n, statement := range strings.Split(schema, ";") {
		if len(statement) > 5 {
			if _, err := db.Exec(statement); err != nil {
				panic(fmt.Sprintf("Statement %d failed: \"%s\" : %s", n+1, statement, err))
			}
		}
	}
}
