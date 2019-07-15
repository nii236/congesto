package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func scrape(url string) (RegionSlice, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	result, err := processResponse(res.Body)
	if err != nil {
		return nil, err
	}

	previousStatus.Lock()
	previousStatus.Regions = nextStatus.Regions
	previousStatus.Unlock()
	nextStatus.Lock()
	nextStatus.Regions = result
	nextStatus.Unlock()

	if len(previousStatus.Regions) > 0 && len(nextStatus.Regions) > 0 {
		updates, err := statusDiff(previousStatus, nextStatus)
		if err != nil {
			return nil, err
		}
		if len(updates) > 0 {
			processNotifications(updates)
		}
	}
	return result, nil
}

func processServer(i int, s *goquery.Selection) (*Server, error) {
	svrName := strings.TrimSpace(s.Find(".world-list__world_name").First().Text())
	categoryStr := strings.TrimSpace(s.Find(".world-list__world_category").First().Text())
	createCharacter := strings.TrimSpace(s.Find(".world-list__create_character").First().Find("i").AttrOr("data-tooltip", "create character attr not found"))
	createCharacterAvailable := false
	if createCharacter == "Creation of New Characters Available" {
		createCharacterAvailable = true
	}
	category := CategoryUnknown
	switch Category(categoryStr) {
	case CategoryNew:
		category = CategoryNew
	case CategoryStandard:
		category = CategoryStandard
	case CategoryPreferred:
		category = CategoryPreferred
	case CategoryCongested:
		category = CategoryCongested
	default:
		return nil, fmt.Errorf("Unknown category: %s", categoryStr)
	}

	svr := &Server{Name: svrName, Category: category, CreateCharacterAvailable: createCharacterAvailable}
	return svr, nil

}

func processDataCentre(i int, s *goquery.Selection) (*DataCentre, error) {
	dcName := s.Find("h2").First().Text()
	if dcName == "" {
		return nil, ErrEmpty
	}
	dc := &DataCentre{Name: dcName, Servers: []*Server{}}
	s.Find(".world-list__item").Each(func(i int, s *goquery.Selection) {
		svr, err := processServer(i, s)

		if err != nil {
			fmt.Println(err)
			return
		}
		dc.Servers = append(dc.Servers, svr)
	})
	return dc, nil
}

func processRegion(i int, doc *goquery.Document, s *goquery.Selection) (*Region, error) {
	regionName := strings.TrimSpace(strings.ReplaceAll(s.Text(), "Data Center", ""))
	region := &Region{Name: regionName, DataCentres: []*DataCentre{}}
	doc.Find(fmt.Sprintf("div[data-region='%d'].js--tab-content", i+1)).Find("ul").First().Find("li").Each(func(i int, s *goquery.Selection) {
		dc, err := processDataCentre(i, s)
		if err != nil && err == ErrEmpty {
			return
		}
		if err != nil {
			fmt.Println(err)
			return
		}

		region.DataCentres = append(region.DataCentres, dc)
	})
	return region, nil
}

func processResponse(r io.ReadCloser) (RegionSlice, error) {
	regions := RegionSlice{}
	defer r.Close()
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	doc.Find(".world__tab").First().Find("li").Each(func(i int, s *goquery.Selection) {
		region, err := processRegion(i, doc, s)
		if err != nil {
			fmt.Println(err)
			return
		}
		regions = append(regions, region)
	})

	return regions, err
}
