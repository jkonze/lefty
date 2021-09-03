package inmem

import (
	"encoding/json"
	"fmt"
	"github.com/chrismeh/lefty/pkg/products"
	"io"
	"sort"
	"sync"
	"time"
)

type ProductStore struct {
	products map[string]products.Product
	mu       *sync.Mutex
}

func NewProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[string]products.Product),
		mu:       &sync.Mutex{},
	}
}

func (p *ProductStore) FindAll() ([]products.Product, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	prds := make([]products.Product, 0, len(p.products))

	for _, v := range p.products {
		prds = append(prds, v)
	}

	sort.Slice(prds, func(i, j int) bool {
		return prds[i].Price < prds[j].Price
	})
	return prds, nil
}

func (p *ProductStore) Upsert(products []products.Product) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	for _, product := range products {
		key := buildProductKey(product)
		if _, exists := p.products[key]; !exists {
			product.CreatedAt = now
		}
		product.UpdatedAt = now
		p.products[key] = product
	}

	return nil
}

func (p *ProductStore) Dump(w io.Writer) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return json.NewEncoder(w).Encode(p.products)
}

func (p *ProductStore) Load(r io.Reader) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return json.NewDecoder(r).Decode(&p.products)
}

func buildProductKey(p products.Product) string {
	return fmt.Sprintf("%s-%s-%s", p.Retailer, p.Manufacturer, p.Model)
}
