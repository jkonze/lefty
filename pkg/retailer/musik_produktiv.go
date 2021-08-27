package retailer

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chrismeh/lefty/pkg/products"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type MusikProduktiv struct {
	http interface {
		Get(url string) (*http.Response, error)
	}
	manufacturers []string
}

func (m *MusikProduktiv) LoadProducts(category string) (ProductResponse, error) {
	resp, err := m.http.Get(fmt.Sprintf("https://www.musik-produktiv.de/%s", category))
	if err != nil {
		return ProductResponse{}, fmt.Errorf("could not fetch products from musik-produktiv.de: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ProductResponse{}, fmt.Errorf("could not create goquery document from reader: %w", err)
	}
	categoryName := doc.Find("div.list_title h1").Text()

	manufacturerNodes := doc.Find(".mp-filtermenu ul").First().Find("li span")
	m.manufacturers = make([]string, len(manufacturerNodes.Nodes))
	manufacturerNodes.Each(func(i int, s *goquery.Selection) {
		m.manufacturers[i] = s.Text()
	})

	instrumentNodes := doc.Find("ul.artgrid li")
	instruments := make([]products.Product, len(instrumentNodes.Nodes))
	instrumentNodes.Each(func(i int, s *goquery.Selection) {
		p, err := m.parseProduct(s)
		if err != nil {
			return
		}

		p.Category = categoryName
		instruments[i] = p
	})

	currentPage, lastPage, err := m.parsePagination(doc)
	if err != nil {
		return ProductResponse{}, err
	}

	return ProductResponse{
		Products:    instruments,
		CurrentPage: uint(currentPage),
		LastPage:    uint(lastPage),
	}, nil
}

func (m *MusikProduktiv) parseProduct(s *goquery.Selection) (products.Product, error) {
	manufacturer, model := m.parseProductName(s.Find("b").First().Text())
	price, err := m.parsePrice(s.Find("i").Text())
	if err != nil {
		return products.Product{}, err
	}

	return products.Product{
		Retailer:     "Musik Produktiv",
		Manufacturer: manufacturer,
		Model:        model,
		Price:        price,
		IsAvailable:  !s.Find(".ampel").HasClass("zzz"),
		ProductURL:   s.Find("a").First().AttrOr("href", ""),
		ThumbnailURL: s.Find("img").First().AttrOr("src", ""),
	}, nil
}

func (m *MusikProduktiv) parseProductName(productName string) (manufacturer, model string) {
	for _, man := range m.manufacturers {
		if strings.HasPrefix(productName, man) {
			return man, strings.TrimPrefix(productName, man+" ")
		}
	}

	parts := strings.Split(productName, " ")
	return parts[0], strings.TrimPrefix(productName, parts[0]+" ")
}

func (m *MusikProduktiv) parsePrice(price string) (float64, error) {
	re := regexp.MustCompile("[^0-9]")
	p := re.ReplaceAllString(price, "")

	fPrice, err := strconv.ParseFloat(p, 32)
	if err != nil {
		return 0, err
	}

	return fPrice, nil
}

func (m *MusikProduktiv) parsePagination(s *goquery.Document) (currentPage, lastPage int, err error) {
	pagination := s.Find(".list_page div")
	if len(pagination.Nodes) == 1 {
		return 1, 1, nil
	}

	cp := pagination.Find("div").Text()
	lp := pagination.Find("a").Last().Text()

	currentPage, err = strconv.Atoi(cp)
	if err != nil {
		return 0, 0, fmt.Errorf("could not parse current page from pagination: %w", err)
	}

	lastPage, err = strconv.Atoi(lp)
	if err != nil {
		return 0, 0, fmt.Errorf("could not parse last page from pagination: %w", err)
	}

	return currentPage, lastPage, nil
}
