package main

import (
	"fmt"
	"hackday-diversity/internal"
)

func main() {

	rek, err := internal.NewRekognition()
	internal.ExitOnError(err)

	db, err := internal.NewDefaultHackdayDB()
	internal.ExitOnError(err)

	rows, err := db.GetJournalists()
	internal.ExitOnError(err)

	for rows.Next() {
		row, err := rows.Scan()
		internal.ExitOnError(err)

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
