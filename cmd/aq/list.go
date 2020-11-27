// The parse command is mostly there to check that we can parse the
// CSV file(s) and visually check that a calibration file looks okay.
//
package main

import (
	"fmt"
	"log"
)

// ListCommand defines the command line parameters for list command
type ListCommand struct {
	Offset int `long:"offset" description:"offset from lowest id" default:"0"`
	Limit  int `long:"limit" description:"Number of entries to show" default:"200"`
}

func init() {
	parser.AddCommand("list",
		"List calibration data",
		"List calibration data",
		&ListCommand{})
}

// Execute runs the list command
func (a *ListCommand) Execute(args []string) error {
	db, err := getDB()
	defer db.Close()

	cals, err := db.ListCals(a.Offset, a.Limit)
	if err != nil {
		log.Fatalf("Unable to list calibration data: %v", err)
	}

	if len(cals) == 0 {
		log.Printf("no entries to list")
		return nil
	}

	fmt.Print("\n---------------------------------------------------------------------------\n")
	fmt.Print("   ID  CollectionID    DeviceID         AFE Serial  ValidFrom\n")
	fmt.Print("---------------------------------------------------------------------------\n")
	for _, cal := range cals {
		fmt.Printf(" %4d  %14s  %14s  %10s   %20s\n", cal.ID, cal.CollectionID, cal.DeviceID, cal.AFESerial, cal.ValidFrom.Format(layout))
	}
	fmt.Print("---------------------------------------------------------------------------\n\n")
	return nil
}
