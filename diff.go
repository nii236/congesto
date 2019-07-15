package main

import (
	"strconv"

	"github.com/r3labs/diff"
)

func statusDiff(prev, next *Status) ([]*UpdatedServer, error) {
	if prev == nil || next == nil {
		return nil, nil
	}
	d, err := diff.NewDiffer()
	if err != nil {
		return nil, err
	}

	changeLog, err := d.Diff(prev, next)
	if err != nil {
		return nil, err
	}
	if changeLog == nil {
		return nil, nil
	}
	updates, err := processDiff(changeLog, next)
	if err != nil {
		return nil, err
	}
	return updates, nil
}

func processDiff(changes diff.Changelog, next *Status) ([]*UpdatedServer, error) {
	result := []*UpdatedServer{}
	for _, cl := range changes {
		if len(cl.Path) != 7 {
			continue
		}
		iRegionStr := cl.Path[1]
		iRegion, err := strconv.Atoi(iRegionStr)
		if err != nil {
			return nil, err
		}
		iDataCentreStr := cl.Path[3]
		iDataCentre, err := strconv.Atoi(iDataCentreStr)
		if err != nil {
			return nil, err
		}
		iServerStr := cl.Path[5]
		iServer, err := strconv.Atoi(iServerStr)
		if err != nil {
			return nil, err
		}
		region := next.Regions[iRegion]
		dataCentre := region.DataCentres[iDataCentre]
		server := dataCentre.Servers[iServer]
		key := cl.Path[6]
		from := cl.From
		to := cl.To

		result = append(result, &UpdatedServer{
			Region:     region.Name,
			DataCentre: dataCentre.Name,
			Server:     server.Name,
			Key:        key,
			From:       from,
			To:         to,
		})
	}
	return result, nil

}

// UpdatedServer is a changed key in the struct
type UpdatedServer struct {
	Region     string
	DataCentre string
	Server     string
	Key        string
	From       interface{}
	To         interface{}
}
