package hosttech

import (
	"context"
	"fmt"
	"github.com/libdns/libdns"
	"time"
)

func main() {
	provider := Provider{
		APIToken: "Your API token",
	}

	//Set your zone with the domain
	zone := "example.com"

	//List all records
	allRecords, err := provider.GetRecords(context.Background(), zone)
	if err != nil {
		fmt.Println("Could not read record(s) because of", err.Error())
	}

	for _, record := range allRecords {
		fmt.Println(record)
	}

	//Create a new record...
	newlyCreatedRecords, err := provider.AppendRecords(context.Background(), zone, []libdns.Record{
		{
			ID:       "",
			Type:     "A",
			Name:     "sub",
			Value:    "1.2.3.4",
			TTL:      1800 * time.Second,
			Priority: 0,
		},
	})

	if err != nil {
		fmt.Println("Could not create record(s) because of", err.Error())
	}

	//... and delete it afterwards again
	for _, record := range newlyCreatedRecords {
		deletedRecords, err := provider.DeleteRecords(context.Background(), zone, []libdns.Record{record})

		if err != nil {
			fmt.Println(err.Error())
		}

		for _, deletedRecord := range deletedRecords {
			fmt.Println("Deleted record: ", deletedRecord)
		}
	}

	//List all available zones
	allZones, err := provider.ListZones(context.Background())
	if err != nil {
		fmt.Print(err.Error())
	}

	for _, zone := range allZones {
		fmt.Println("Zone: ", zone.Name)
	}
}
