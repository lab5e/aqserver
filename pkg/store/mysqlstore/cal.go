// Package mysqlstore implements MySQL based store.
package mysqlstore

import (
	"github.com/lab5e/aqserver/pkg/model"
)

// PutCal ...
func (s *MySQLStore) PutCal(c *model.Cal) (int64, error) {
	r, err := s.db.NamedExec(`
INSERT INTO cal
(
  device_id,
  sysid,
  collection_id,
  valid_from,
  afe_serial,
  circuit_type,
  afe_type,
  sensor1_serial,
  sensor2_serial,
  sensor3_serial,
  afe_cal_date,
  vt20_offset,
  sensor1_we_e,
  sensor1_we_0,
  sensor1_ae_e,
  sensor1_ae_0,
  sensor1_pcb_gain, 
  sensor1_we_sensitivity, 
  sensor2_we_e,
  sensor2_we_0,
  sensor2_ae_e,          
  sensor2_ae_0,         
  sensor2_pcb_gain,
  sensor2_we_sensitivity,
  sensor3_we_e,
  sensor3_we_0,
  sensor3_ae_e,
  sensor3_ae_0,
  sensor3_pcb_gain,
  sensor3_we_sensitivity
)
VALUES(
  :device_id,
  :sysid,
  :collection_id,
  :valid_from,
  :afe_serial,
  :circuit_type,
  :afe_type,
  :sensor1_serial,
  :sensor2_serial,
  :sensor3_serial,
  :afe_cal_date,
  :vt20_offset,
  :sensor1_we_e,
  :sensor1_we_0,
  :sensor1_ae_e,
  :sensor1_ae_0,
  :sensor1_pcb_gain, 
  :sensor1_we_sensitivity, 
  :sensor2_we_e,
  :sensor2_we_0,
  :sensor2_ae_e,          
  :sensor2_ae_0,         
  :sensor2_pcb_gain,
  :sensor2_we_sensitivity,
  :sensor3_we_e,
  :sensor3_we_0,
  :sensor3_ae_e,
  :sensor3_ae_0,
  :sensor3_pcb_gain,
  :sensor3_we_sensitivity)
`, c)

	if err != nil {
		return -1, err
	}

	return r.LastInsertId()
}

// GetCal ...
func (s *MySQLStore) GetCal(id int64) (*model.Cal, error) {
	var c model.Cal
	err := s.db.QueryRowx("SELECT * FROM cal WHERE id = ?", id).StructScan(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteCal ...
func (s *MySQLStore) DeleteCal(id int64) error {
	_, err := s.db.Exec("DELETE FROM cal WHERE id = ?", id)
	return err
}

// ListCals ...
//
// TODO(borud): whomever implemented this didn't actually implement this
// correctly since it doesn't heed limit and offset.
func (s *MySQLStore) ListCals(_ int, _ int) ([]model.Cal, error) {
	var cals []model.Cal
	err := s.db.Select(&cals, "SELECT * FROM cal ORDER BY device_id, valid_from ASC")
	return cals, err
}

// ListCalsForDevice ...
func (s *MySQLStore) ListCalsForDevice(deviceID string) ([]model.Cal, error) {
	var cals []model.Cal
	err := s.db.Select(&cals, "SELECT * FROM cal WHERE device_id = ? ORDER BY valid_from DESC", deviceID)
	return cals, err
}
