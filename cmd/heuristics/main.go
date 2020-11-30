package main

import (
	"fmt"
	"hackday-diversity/internal"
	"os"
)

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {

	rek, err := internal.NewRekognition()
	checkError(err)

	db, err := internal.NewDefaultHackdayDB()
	checkError(err)

	rows, err := db.GetJournalists()
	checkError(err)

	for rows.Next() {
		row, err := rows.Scan()
		checkError(err)

		if len(row.ImageName) == 0 || len(row.Gender) > 0 {
			continue
		}

		fmt.Printf("analysing image: %s\n", row.ImageName)
		output, err := internal.AnalyseFace(rek, row.ImageName)
		if err != nil {
			fmt.Printf("error analysing image %s\n", row.ImageName)
			continue
		}

		if len(output.FaceDetails) == 0 {
			continue
		}

		details := output.FaceDetails[0]
		if err := db.UpdateJournalistHeuristics(row.Name, details); err != nil {
			fmt.Printf("unable to update heuristics for journalist %s\n", row.Name)
			continue
		}
	}
}
