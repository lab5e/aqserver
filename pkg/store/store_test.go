package store

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/lab5e/aqserver/pkg/model"
	"github.com/lab5e/aqserver/pkg/store/sqlitestore"
	"github.com/stretchr/testify/assert"
)

var testCal = model.Cal{
	DeviceID:             "device1",
	CollectionID:         "mycollection",
	ValidFrom:            time.Now().Add(-24 * time.Hour),
	AFESerial:            "some-serial-character-sequence",
	AFECalDate:           time.Now().Add(-24 * time.Hour),
	Vt20Offset:           0.3195,
	Sensor1WEe:           312,
	Sensor1WE0:           -5,
	Sensor1AEe:           316,
	Sensor1AE0:           -5,
	Sensor1PCBGain:       -0.73,
	Sensor1WESensitivity: 0.203,
	Sensor2WEe:           411,
	Sensor2WE0:           -4,
	Sensor2AEe:           411,
	Sensor2AE0:           -3,
	Sensor2PCBGain:       -0.73,
	Sensor2WESensitivity: 0.363,
	Sensor3WEe:           271,
	Sensor3WE0:           19,
	Sensor3AEe:           256,
	Sensor3AE0:           23,
	Sensor3PCBGain:       0.8,
	Sensor3WESensitivity: 0.408,
}

func TestSqlitestore(t *testing.T) {
	// Cal tests
	{
		var db Store

		db, err := sqlitestore.New(":memory:")
		assert.Nil(t, err, "Error instantiating new sqlitestore")
		assert.NotNil(t, db)
		calTests(t, db)
		db.Close()
	}

	// Message tests
	{
		var db Store

		db, err := sqlitestore.New(":memory:")
		assert.Nil(t, err, "Error instantiating new sqlitestore")
		assert.NotNil(t, db)
		messageTests(t, db)
		db.Close()
	}

}

// calTests performs CRUD tests on Cal
func calTests(t *testing.T, db Store) {

	{
		// Put
		id, err := db.PutCal(&testCal)
		assert.Nil(t, err)
		assert.True(t, id > 0)

		testCal.ID = id

		// Get
		c, err := db.GetCal(id)
		assert.Nil(t, err)
		assert.NotNil(t, c)

		// TODO(borud): date returned from SQLite3 has different
		// precision and timezone that what we put in, so this has to
		// be fixed at some point.  It's not an error, we just can't
		// do naive equality test.
		// assert.Equal(t, testCal, *c)
	}

	// Delete
	{
		err := db.DeleteCal(testCal.ID)
		assert.Nil(t, err)

		c, err := db.GetCal(testCal.ID)
		assert.NotNil(t, err)
		assert.Nil(t, c)
	}

	// ListCals
	{
		var devices []string
		for i := 0; i < 20; i++ {
			did := fmt.Sprintf("cal-device-%d", i)

			id, err := db.PutCal(&model.Cal{
				DeviceID:  did,
				ValidFrom: time.Now().Add(-24 * time.Hour),
			})
			assert.Nil(t, err)
			assert.True(t, id > 0)
			devices = append(devices, did)
		}

		cals, err := db.ListCals(0, 100)
		assert.Nil(t, err)
		assert.Equal(t, 20, len(cals))

		// Make sure that there is one cal for each device
		for _, dev := range devices {
			devcals, err := db.ListCalsForDevice(dev)
			assert.Nil(t, err)
			assert.Equal(t, 1, len(devcals))
		}
	}

}

// messageTests performs CRUD tests on Messages
func messageTests(t *testing.T, db Store) {

	numDevices := 3
	numMessagesPerDevice := (60 * 24)
	t0 := time.Now()

	var messageIDs []int64

	// Populate some devices and messages
	for i := 0; i < numDevices; i++ {
		deviceID := fmt.Sprintf("msg-device-%d", i)

		for j := 0; j < numMessagesPerDevice; j++ {
			msg := &model.Message{
				DeviceID:     deviceID,
				ReceivedTime: ms(t0.Add(time.Duration(j) * time.Minute)),
			}
			id, err := db.PutMessage(msg)
			assert.Nil(t, err)
			assert.True(t, id > 0)
			messageIDs = append(messageIDs, id)
		}
	}

	totalMessages := numMessagesPerDevice * numDevices
	assert.Equal(t, totalMessages, len(messageIDs))

	// Fetch some random messages
	for i := 0; i < 20; i++ {
		id := rand.Int63n(int64(totalMessages))

		m, err := db.GetMessage(id)
		assert.Nil(t, err)
		assert.NotNil(t, m)
	}

	// ListMessages
	{
		msgs, err := db.ListMessages(0, 10)
		assert.Nil(t, err)
		assert.NotNil(t, msgs)
		assert.Equal(t, 10, len(msgs))
	}

	// ListMessagesByDate
	{
		duration := time.Minute * 10

		from := ms(t0)
		to := ms(t0.Add(duration))

		log.Printf("From: %d, To: %d", from, to)

		msgs, err := db.ListMessagesByDate(from, to)
		assert.Nil(t, err)
		assert.Equal(t, numDevices*10, len(msgs))
	}

	// ListDeviceMessagesByDate
	{
		for i := 0; i < numDevices; i++ {
			deviceID := fmt.Sprintf("msg-device-%d", i)
			duration := time.Minute * 10
			msgs, err := db.ListDeviceMessagesByDate(deviceID, ms(t0), ms(t0.Add(duration)))
			assert.Nil(t, err)
			assert.Equal(t, 10, len(msgs))
		}
	}
}

func ms(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
